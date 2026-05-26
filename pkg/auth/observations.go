package auth

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ObservedCredential is one auth-related signal seen in a Capture.
type ObservedCredential struct {
	Kind         string   `yaml:"kind"`
	Name         string   `yaml:"name,omitempty"`
	Coverage     float64  `yaml:"coverage"`
	SamplePaths  []string `yaml:"sample_paths,omitempty"`
	SampleValue  string   `yaml:"sample_value,omitempty"`
	RequestCount int      `yaml:"request_count,omitempty"`
}

// SuggestedAuthPath is a weak hint that a path may be an Auth Endpoint.
type SuggestedAuthPath struct {
	Path   string `yaml:"path"`
	Reason string `yaml:"reason,omitempty"`
}

// ObservationsFile is the sidecar written by auth observe.
type ObservationsFile struct {
	APIPrefix           string               `yaml:"api_prefix,omitempty"`
	Capture             string               `yaml:"capture,omitempty"`
	SuccessRequestCount int                  `yaml:"success_request_count,omitempty"`
	Observed            []ObservedCredential `yaml:"observed"`
	Verified            []string             `yaml:"verified,omitempty"`
	Combination         string               `yaml:"combination,omitempty"`
	SuggestedAuthPaths  []SuggestedAuthPath  `yaml:"suggested_auth_paths,omitempty"`
}

// LoadObservationsFile reads an auth observations sidecar YAML.
func LoadObservationsFile(path string) (*ObservationsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f ObservationsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func saveObservationsFile(path string, f *ObservationsFile) error {
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
