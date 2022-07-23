/*******************************************************************************
*
* Copyright 2020 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	dataKey            = "data"
	resourceKindKey    = "kind"
	resourceKindSecret = "Secret"
)

type (
	// ValueTransformer ...
	ValueTransformer struct {
		replacements map[string]string
	}

	config struct {
		ApiVersion string `yaml:"apiVersion,omitempty"`
		Kind       string `yaml:"kind,omitempty"`
		Metadata   struct {
			Name string `yaml:"name,omitempty"`
		} `yaml:"metadata,omitempty"`

		Values map[string]string `yaml:"values,omitempty"`
	}

	resource map[string]interface{}
)

func main() {
	// Ignore plugin config.
	args := os.Args
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, "no replacements configured")
		return
	}

	t := New(args[1], args[2:])
	if err := t.Transform(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

// New creates a new ValueTransformer.
func New(pluginConfigPath string, args []string) *ValueTransformer {
	replacements := make(map[string]string, 0)
	lenArgs := len(args)

	for idx, a := range args {
		// Handle key:value
		itm := strings.Split(a, ":")
		if len(itm) == 2 && itm[1] != "" {
			replacements[strings.TrimSpace(itm[0])] = strings.TrimSpace(itm[1])
			continue
		}

		// Handle key: value (2 arguments)
		if strings.HasSuffix(a, ":") && idx+1 <= lenArgs {
			key := strings.TrimSuffix(a, ":")
			key = strings.TrimSpace(key)
			replacements[key] = strings.TrimSpace(args[idx+1])
		}
	}

	if cfg, err := loadConfig(pluginConfigPath); err == nil {
		for k, v := range cfg.Values {
			replacements[k] = v
		}
	}

	return &ValueTransformer{
		replacements: replacements,
	}
}

// Transform ...
func (t *ValueTransformer) Transform() error {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "error reading from stdin")
	}

	// Replace in plain text.
	b = t.replace(b)

	// Shortcut for resources without base64 encoded content.
	if !bytes.Contains(b, []byte(fmt.Sprintf("%s: %s", resourceKindKey, resourceKindSecret))) {
		_, err := fmt.Fprint(os.Stdout, string(b))
		return errors.Wrap(err, "error printing to stdout")
	}

	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		var res resource

		err := dec.Decode(&res)
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "error decoding yaml")
		}

		if kind, ok := res[resourceKindKey]; ok && kind == resourceKindSecret {
			res, err = t.replaceSecret(res)
			if err != nil {
				return err
			}
		}

		resBytes, err := yaml.Marshal(res)
		if err != nil {
			return errors.Wrap(err, "error marshalling resources")
		}

		_, err = fmt.Fprintf(os.Stdout, "---\n%s", string(resBytes))
		if err != nil {
			return errors.Wrap(err, "error printing to stdout")
		}
	}

	return nil
}

func (t *ValueTransformer) replace(b []byte) []byte {
	for old, newBy := range t.replacements {
		oldByte := []byte(fmt.Sprintf("$%s", old))
		newByte := []byte(newBy)
		b = bytes.ReplaceAll(b, oldByte, newByte)
	}
	return b
}

func (t *ValueTransformer) replaceSecret(res resource) (resource, error) {
	data, ok := res[dataKey]
	if !ok {
		return res, nil
	}

	dataMap, err := toStringMap(data)
	if err != nil {
		return nil, errors.New("cannot convert data")
	}
	if dataMap == nil || len(dataMap) == 0 {
		return res, nil
	}

	newData := make(map[string]string, len(dataMap))
	for k, v := range dataMap {
		valDec, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, errors.Wrap(err, "error base64 decoding")
		}

		newVal := t.replace(valDec)
		newData[k] = base64.StdEncoding.EncodeToString(newVal)
	}

	res[dataKey] = newData
	return res, nil
}

func toStringMap(in interface{}) (map[string]string, error) {
	switch in.(type) {
	case resource:
		data := in.(resource)
		m := make(map[string]string, 0)
		for k, v := range data {
			m[k] = fmt.Sprintf("%v", v)
		}
		return m, nil
	case map[string]string:
		return in.(map[string]string), nil
	}

	return nil, fmt.Errorf("cannot convert type %s", reflect.TypeOf(in).String())
}

func loadConfig(path string) (*config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config
	err = yaml.Unmarshal(b, &cfg)
	return &cfg, err
}
