package golden

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/schema"
)

// CompareYAML reports whether two Schema YAML documents are semantically equal.
// info.title is ignored because it embeds the capture file path.
func CompareYAML(expected, actual []byte) error {
	want, err := schema.LoadBytes(expected)
	if err != nil {
		return fmt.Errorf("parse expected schema: %w", err)
	}
	got, err := schema.LoadBytes(actual)
	if err != nil {
		return fmt.Errorf("parse actual schema: %w", err)
	}

	wantNorm := normalizeForCompare(want)
	gotNorm := normalizeForCompare(got)
	if cmp.Equal(wantNorm, gotNorm) {
		return nil
	}
	return fmt.Errorf("schema mismatch (-golden +generated):\n%s", cmp.Diff(wantNorm, gotNorm))
}

func normalizeForCompare(doc *schema.Document) *schema.Document {
	if doc == nil {
		return nil
	}
	out := *doc
	if out.Info != nil {
		info := make(map[string]any, len(out.Info))
		for key, value := range out.Info {
			if key == "title" {
				continue
			}
			info[key] = value
		}
		out.Info = info
	}
	return &out
}
