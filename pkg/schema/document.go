package schema

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Document is an OpenAPI 3 schema file used during a Pass.
type Document struct {
	OpenAPI        string         `yaml:"openapi"`
	Info           map[string]any `yaml:"info"`
	Servers        []any          `yaml:"servers,omitempty"`
	Paths          MapSlice       `yaml:"paths,omitempty"`
	Tags           []any          `yaml:"tags,omitempty"`
	Components     map[string]any `yaml:"components,omitempty"`
	Security       []any          `yaml:"security,omitempty"`
	XTagGroups     []any          `yaml:"x-tagGroups,omitempty"`
	XPathTemplates []string       `yaml:"x-path-templates,omitempty"`
}

func Load(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func New(inputPath string) *Document {
	return &Document{
		OpenAPI: "3.0.0",
		Info: map[string]any{
			"title":   inputPath + " Mitmproxy2Swagger",
			"version": "1.0.0",
		},
		Servers:        []any{},
		Paths:          MapSlice{},
		XPathTemplates: []string{},
	}
}

func (d *Document) EnsureDefaults(apiPrefix string) {
	if d.Servers == nil {
		d.Servers = []any{}
	}
	hasServer := false
	for _, s := range d.Servers {
		m, _ := s.(map[string]any)
		if m != nil && m["url"] == apiPrefix {
			hasServer = true
			break
		}
	}
	if !hasServer {
		d.Servers = append(d.Servers, map[string]any{
			"url":         apiPrefix,
			"description": "The default server",
		})
	}
	if d.Paths == nil {
		d.Paths = MapSlice{}
	}
	if d.XPathTemplates == nil {
		d.XPathTemplates = []string{}
	}
}

func (d *Document) PathTemplates() []string {
	out := make([]string, 0, len(d.Paths)+len(d.XPathTemplates))
	for _, item := range d.Paths {
		out = append(out, fmt.Sprint(item.Key))
	}
	out = append(out, d.XPathTemplates...)
	return out
}

func (d *Document) PathOperations(pathTemplate string) (map[string]any, bool) {
	for _, item := range d.Paths {
		if fmt.Sprint(item.Key) == pathTemplate {
			m, ok := item.Value.(map[string]any)
			return m, ok
		}
	}
	return nil, false
}

func (d *Document) SetPathIfNotExists(pathTemplate string, operations map[string]any) {
	if _, ok := d.PathOperations(pathTemplate); ok {
		return
	}
	d.Paths = append(d.Paths, MapItem{Key: pathTemplate, Value: operations})
}

func (d *Document) Save(path string) error {
	data, err := Marshal(d)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func SetKeyIfNotExists(m map[string]any, key string, value any) {
	if _, ok := m[key]; !ok {
		m[key] = value
	}
}

func DedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func FilterXPathTemplates(templates []string, paths MapSlice) []string {
	pathKeys := make(map[string]struct{}, len(paths))
	for _, item := range paths {
		pathKeys[fmt.Sprint(item.Key)] = struct{}{}
	}
	out := make([]string, 0, len(templates))
	for _, t := range templates {
		if _, ok := pathKeys[t]; ok {
			continue
		}
		out = append(out, t)
	}
	return out
}
