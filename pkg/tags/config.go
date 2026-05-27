package tags

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TagDef is one entry for OpenAPI top-level tags:.
type TagDef struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// TagGroup is a Redoc x-tagGroups entry.
type TagGroup struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
}

// PrefixRule maps paths with a given prefix to tags (first tag wins on apply).
type PrefixRule struct {
	Prefix string   `yaml:"prefix"`
	Tags   []string `yaml:"tags"`
}

// TagsFile is the sidecar for tags apply.
type TagsFile struct {
	Tags         []TagDef          `yaml:"tags,omitempty"`
	XTagGroups   []TagGroup        `yaml:"x-tagGroups,omitempty"`
	PrefixRules  []PrefixRule      `yaml:"prefixRules,omitempty"`
	Operations   map[string][]string `yaml:"operations,omitempty"`
}

// LoadTagsFile reads a tags sidecar YAML.
func LoadTagsFile(path string) (*TagsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f TagsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (f *TagsFile) validate() error {
	if f == nil {
		return fmt.Errorf("tags file is nil")
	}
	seen := make(map[string]struct{})
	for _, t := range f.Tags {
		if strings.TrimSpace(t.Name) == "" {
			return fmt.Errorf("tags: entry with empty name")
		}
		if _, ok := seen[t.Name]; ok {
			return fmt.Errorf("tags: duplicate name %q", t.Name)
		}
		seen[t.Name] = struct{}{}
	}
	for i, r := range f.PrefixRules {
		p := strings.TrimSpace(r.Prefix)
		if p == "" {
			return fmt.Errorf("prefixRules[%d]: empty prefix", i)
		}
		if p[0] != '/' {
			return fmt.Errorf("prefixRules[%d]: prefix must start with /, got %q", i, p)
		}
		if len(r.Tags) == 0 {
			return fmt.Errorf("prefixRules[%d]: tags required", i)
		}
	}
	for key, tags := range f.Operations {
		if _, _, ok := parseOperationKey(key); !ok {
			return fmt.Errorf("operations: invalid key %q (want METHOD /path)", key)
		}
		if len(tags) == 0 {
			return fmt.Errorf("operations[%q]: tags required", key)
		}
	}
	return nil
}

func parseOperationKey(key string) (method, path string, ok bool) {
	key = strings.TrimSpace(key)
	i := strings.IndexByte(key, ' ')
	if i <= 0 {
		return "", "", false
	}
	method = strings.ToUpper(strings.TrimSpace(key[:i]))
	path = strings.TrimSpace(key[i+1:])
	if path == "" || path[0] != '/' {
		return "", "", false
	}
	return method, path, true
}
