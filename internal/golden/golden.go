package golden

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/pass"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/schema"
	"gopkg.in/yaml.v3"
)

type Case struct {
	ID        string   `yaml:"id"`
	Capture   string   `yaml:"capture"`
	APIPrefix string   `yaml:"api_prefix"`
	ExtraArgs []string `yaml:"extra_args"`
}

type CasesFile struct {
	Cases []Case `yaml:"cases"`
}

func LoadCases(path string) ([]Case, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc CasesFile
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Cases) == 0 {
		return nil, fmt.Errorf("no cases defined in %s", path)
	}
	return doc.Cases, nil
}

func FilterCases(cases []Case, ids []string) ([]Case, error) {
	if len(ids) == 0 {
		return cases, nil
	}
	want := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		want[id] = struct{}{}
	}
	var filtered []Case
	for _, c := range cases {
		if _, ok := want[c.ID]; ok {
			filtered = append(filtered, c)
			delete(want, c.ID)
		}
	}
	if len(want) > 0 {
		unknown := make([]string, 0, len(want))
		for id := range want {
			unknown = append(unknown, id)
		}
		return nil, fmt.Errorf("unknown case id(s): %s", strings.Join(unknown, ", "))
	}
	return filtered, nil
}

func ResolveCapture(repoRoot, capture string) (string, error) {
	if filepath.IsAbs(capture) {
		return "", fmt.Errorf("capture must be a path relative to repository root: %s", capture)
	}
	path := filepath.Clean(filepath.Join(repoRoot, capture))
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("capture must stay inside repository: %s", capture)
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("capture not found: %s", path)
	}
	return path, nil
}

func caseToPassOptions(c Case, capture, output string) (pass.Options, error) {
	opts := pass.Options{
		Input:      capture,
		Output:     output,
		APIPrefix:  c.APIPrefix,
		ParamRegex: "[0-9]+",
	}
	for i := 0; i < len(c.ExtraArgs); i++ {
		arg := c.ExtraArgs[i]
		switch arg {
		case "--format", "-f":
			if i+1 >= len(c.ExtraArgs) {
				return opts, fmt.Errorf("missing value for %s", arg)
			}
			i++
			opts.Format = c.ExtraArgs[i]
		case "--headers", "-hd":
			opts.Headers = true
		case "--examples", "-e":
			opts.Examples = true
		case "--suppress-params", "-s":
			opts.SuppressParams = true
		case "--param-regex", "-r":
			if i+1 >= len(c.ExtraArgs) {
				return opts, fmt.Errorf("missing value for %s", arg)
			}
			i++
			opts.ParamRegex = c.ExtraArgs[i]
		default:
			return opts, fmt.Errorf("unknown extra arg %q", arg)
		}
	}
	return opts, nil
}

func StripIgnorePrefixes(schemaPath string) error {
	doc, err := schema.Load(schemaPath)
	if err != nil {
		return err
	}
	for i, entry := range doc.XPathTemplates {
		doc.XPathTemplates[i] = strings.ReplaceAll(entry, "ignore:", "")
	}
	return doc.Save(schemaPath)
}
