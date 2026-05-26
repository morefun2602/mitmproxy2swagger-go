package schema

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// MapItem is an ordered key/value pair in a YAML mapping.
type MapItem struct {
	Key   any
	Value any
}

// MapSlice preserves insertion order for YAML mappings.
type MapSlice []MapItem

func (m *MapSlice) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", value.Kind)
	}
	*m = nil
	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valNode := value.Content[i+1]
		var key string
		if err := keyNode.Decode(&key); err != nil {
			return err
		}
		var val any
		if err := valNode.Decode(&val); err != nil {
			return err
		}
		*m = append(*m, MapItem{Key: key, Value: val})
	}
	return nil
}

func (m MapSlice) MarshalYAML() (any, error) {
	node := yaml.Node{Kind: yaml.MappingNode}
	for _, item := range m {
		keyNode := yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprint(item.Key)}
		var valNode yaml.Node
		if err := valNode.Encode(item.Value); err != nil {
			return nil, err
		}
		node.Content = append(node.Content, &keyNode, &valNode)
	}
	return node, nil
}
