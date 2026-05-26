package golden

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/pass"
)

// RunCase executes the two-pass workflow for a case and returns the resulting Schema YAML.
func RunCase(repoRoot string, c Case) ([]byte, error) {
	capture, err := ResolveCapture(repoRoot, c.Capture)
	if err != nil {
		return nil, err
	}

	tmp, err := os.CreateTemp("", "golden-*.yaml")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	if err := os.Remove(tmpPath); err != nil {
		return nil, err
	}
	defer os.Remove(tmpPath)

	opts, err := caseToPassOptions(c, capture, tmpPath)
	if err != nil {
		return nil, err
	}

	if err := pass.Run(opts); err != nil {
		return nil, fmt.Errorf("first pass: %w", err)
	}
	if err := StripIgnorePrefixes(tmpPath); err != nil {
		return nil, fmt.Errorf("curation: %w", err)
	}
	if err := pass.Run(opts); err != nil {
		return nil, fmt.Errorf("second pass: %w", err)
	}

	return os.ReadFile(tmpPath)
}

// VerifyOne runs a case and compares the output with the committed golden file.
func VerifyOne(repoRoot, goldenDir string, c Case) error {
	goldenPath := filepath.Join(goldenDir, c.ID+".yaml")
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		return fmt.Errorf("[%s] read golden %s: %w", c.ID, goldenPath, err)
	}

	actual, err := RunCase(repoRoot, c)
	if err != nil {
		return fmt.Errorf("[%s] run case: %w", c.ID, err)
	}

	if err := CompareYAML(expected, actual); err != nil {
		return fmt.Errorf("[%s] %w", c.ID, err)
	}
	fmt.Printf("[%s] OK\n", c.ID)
	return nil
}

// GenerateOne runs a case and writes the output to testdata/golden/<id>.yaml.
func GenerateOne(repoRoot, goldenDir string, c Case) error {
	if err := os.MkdirAll(goldenDir, 0o755); err != nil {
		return err
	}

	fmt.Printf("[%s] generating …\n", c.ID)
	actual, err := RunCase(repoRoot, c)
	if err != nil {
		return err
	}

	outPath := filepath.Join(goldenDir, c.ID+".yaml")
	if err := os.WriteFile(outPath, actual, 0o644); err != nil {
		return err
	}

	fmt.Printf("[%s] → %s\n", c.ID, outPath)
	return nil
}
