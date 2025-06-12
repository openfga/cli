/*
Copyright © 2023 OpenFGA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package authorizationmodel

import (
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"
)

// modFile represents the contents of an fga.mod file.
type modFile struct {
	Schema   string   `yaml:"schema"`
	Contents []string `yaml:"contents"`
}

// parseModFile reads and validates an fga.mod file allowing relative parent paths.
func parseModFile(data []byte) (*modFile, error) {
	yamlMod := modFile{}
	if err := yaml.Unmarshal(data, &yamlMod); err != nil {
		return nil, fmt.Errorf("failed to parse fga.mod: %w", err)
	}

	if yamlMod.Schema == "" {
		return nil, fmt.Errorf("missing schema field")
	}
	if yamlMod.Schema != "1.2" {
		return nil, fmt.Errorf("unsupported schema version, fga.mod only supported in version `1.2`")
	}
	if len(yamlMod.Contents) == 0 {
		return nil, fmt.Errorf("missing contents field")
	}

	normalized := make([]string, 0, len(yamlMod.Contents))
	for _, file := range yamlMod.Contents {
		decoded, err := url.QueryUnescape(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode path: %s", file)
		}
		path := strings.ReplaceAll(decoded, "\\", "/")

		if strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("invalid contents item %s", file)
		}
		if !strings.HasSuffix(path, ".fga") {
			return nil, fmt.Errorf("contents items should use fga file extension, got %s", file)
		}

		normalized = append(normalized, path)
	}

	yamlMod.Contents = normalized

	return &yamlMod, nil
}
