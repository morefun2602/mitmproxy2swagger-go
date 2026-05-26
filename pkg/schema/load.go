package schema

import "gopkg.in/yaml.v3"

// LoadBytes parses YAML schema bytes into a Document.
func LoadBytes(data []byte) (*Document, error) {
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
