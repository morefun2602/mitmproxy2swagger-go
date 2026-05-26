package curate

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TemplateSuggestion is one LLM-assisted merge proposal for x-path-templates.
type TemplateSuggestion struct {
	ProposedTemplate string   `yaml:"proposed_template"`
	Replaces         []string `yaml:"replaces"`
	Confidence       string   `yaml:"confidence,omitempty"`
	SamplePaths      []string `yaml:"sample_paths,omitempty"`
	Reason           string   `yaml:"reason,omitempty"`
}

// SuggestionsFile is the sidecar YAML written by --llm-suggest and read by --apply-suggestions.
type SuggestionsFile struct {
	Suggestions []TemplateSuggestion `yaml:"suggestions"`
}

// LoadSuggestionsFile reads a template suggestions sidecar YAML.
func LoadSuggestionsFile(path string) (*SuggestionsFile, error) {
	return loadSuggestionsFile(path)
}

func loadSuggestionsFile(path string) (*SuggestionsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f SuggestionsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func saveSuggestionsFile(path string, f *SuggestionsFile) error {
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
