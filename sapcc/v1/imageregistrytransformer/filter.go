package main

import (
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const imageRegistryDockerHub = "dockerhub"

var (
	_ kio.Filter          = Filter{}
	_ kio.TrackableFilter = &Filter{}

	regexImageName = regexp.MustCompile("^(?P<registry>.*?)(\\/.*)+(?P<name>\\/.+)\\:(?P<tag>.+)$")
)

type Filter struct {
	ImageRegistries []ImageRegistry
	FsSlice         types.FsSlice
	trackableSetter filtersutil.TrackableSetter
}

func (f *Filter) WithMutationTracker(callback func(key, value, tag string, node *yaml.RNode)) {
	f.trackableSetter.WithMutationTracker(callback)
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(
		func(node *yaml.RNode) (*yaml.RNode, error) {
			if err := node.PipeE(fsslice.Filter{
				FsSlice:  f.FsSlice,
				SetValue: f.setImageRegistry,
			}); err != nil {
				return nil, err
			}
			return node, nil
		})).Filter(nodes)
	return nodes, err
}

func (f Filter) setImageRegistry(rn *yaml.RNode) error {
	image := rn.YNode().Value
	for _, imageRegistry := range f.ImageRegistries {
		if strings.HasPrefix(image, imageRegistry.Name) {
			return f.trackableSetter.SetScalar(replaceImageRegistry(
				image, imageRegistry.Name, imageRegistry.NewName,
			))(rn)
		}
	}
	// Assume dockerhub.
	if match := regexImageName.FindStringSubmatch(image); len(match) == 0 {
		if dockerImageRegistry, ok := getNewImageFromImageRegistryConfig(f.ImageRegistries, imageRegistryDockerHub); ok {
			return f.trackableSetter.SetScalar(replaceImageRegistry(
				image, "", dockerImageRegistry,
			))(rn)
		}
	}
	return nil
}

func replaceImageRegistry(image, oldRegistry, newRegistry string) string {
	image = strings.TrimLeft(image, oldRegistry)
	image = strings.TrimLeft(image, "/")
	image = fmt.Sprintf("%s/%s", newRegistry, image)
	return image
}

func getNewImageFromImageRegistryConfig(registries []ImageRegistry, name string) (string, bool) {
	for _, r := range registries {
		if r.Name == name {
			return r.NewName, true
		}
	}
	return "", false
}
