package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// defaultImagePaths specify the paths to look for images.
var defaultImagePaths = []types.FieldSpec{
	{Path: "spec/template/spec/containers[]/image"},
	{Path: "spec/template/spec/initContainers[]/image"},
	{Path: "spec/jobTemplate/spec/template/spec/containers[]/image"},
	{Path: "spec/jobTemplate/spec/template/spec/initContainers[]/image"},
}

type (
	ImageRegistryTransformer struct {
		t resmap.Transformer

		// ImageRegistries is a list of replacements for image registries.
		ImageRegistries []ImageRegistry
		FieldPaths      []types.FieldSpec
	}

	ImageRegistry struct {
		// Name is the current name of the image registry, e.g. k8s.gcr.io .
		Name string `json:"name" yaml:"name"`

		// NewName is the new name of the image registry, e.g. my-k8s-mirror.cloud.sap .
		NewName string `json:"newName" yaml:"newName"`
	}

	config struct {
		ApiVersion string `yaml:"apiVersion,omitempty"`
		Kind       string `yaml:"kind,omitempty"`
		Metadata   struct {
			Name string `yaml:"name,omitempty"`
		} `yaml:"metadata,omitempty"`

		ImageRegistries []ImageRegistry `json:"imageRegistries,omitempty" yaml:"imageRegistries,omitempty"`
		FieldPaths      []string        `json:"fieldPaths,omitempty" yaml:"FieldPaths,omitempty"`
	}
)

func New(pluginConfigPath string, _ []string) (*ImageRegistryTransformer, error) {
	b, err := ioutil.ReadFile(pluginConfigPath)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	fieldPaths := defaultImagePaths
	for _, path := range cfg.FieldPaths {
		fieldPaths = append(fieldPaths, types.FieldSpec{Path: path})
	}
	if err := validateImageRegistries(cfg.ImageRegistries); err != nil {
		return nil, err
	}
	return &ImageRegistryTransformer{
		ImageRegistries: cfg.ImageRegistries,
		FieldPaths:      fieldPaths,
	}, nil
}

func (p *ImageRegistryTransformer) Config(_ *resmap.PluginHelpers, b []byte) error {
	var cfg config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return err
	}
	p.ImageRegistries = cfg.ImageRegistries
	p.FieldPaths = defaultImagePaths
	for _, path := range cfg.FieldPaths {
		p.FieldPaths = append(p.FieldPaths, types.FieldSpec{Path: path})
	}
	return validateImageRegistries(p.ImageRegistries)
}

func (p *ImageRegistryTransformer) Transform(m resmap.ResMap) error {
	if err := m.ApplyFilter(Filter{ImageRegistries: p.ImageRegistries, FsSlice: p.FieldPaths}); err != nil {
		return err
	}
	resYaml, err := m.AsYaml()
	if err != nil {
		return err
	}
	fmt.Println(string(resYaml))
	return nil
}

func main() {
	// Ignore ImageRegistryTransformer config.
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "no image registries configured. usage: ImageRegistryTransformer <path to config>")
		return
	}

	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	resMap, err := bytesToResmap(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	t, err := New(args[1], args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := t.Transform(resMap); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func bytesToResmap(b []byte) (resmap.ResMap, error) {
	resMap := resmap.New()
	factory := resource.NewFactory(&hasher.Hasher{})
	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		var res map[string]interface{}
		err := dec.Decode(&res)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if err := resMap.Append(factory.FromMap(res)); err != nil {
			return nil, err
		}
	}
	return resMap, nil
}

func validateImageRegistries(imageRegistries []ImageRegistry) error {
	for _, imageRegistry := range imageRegistries {
		if imageRegistry.Name == "" || imageRegistry.NewName == "" {
			return fmt.Errorf("image registry name and newName cannot be empty")
		}
	}
	return nil
}
