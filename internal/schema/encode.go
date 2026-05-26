package schema

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

const yamlIndent = 2

// Marshal encodes v as YAML with 2-space indentation (matching Python ruamel.yaml output).
func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(yamlIndent)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
