package curate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	captureopen "github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/open"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/openai/openai-go/v3"
)

func runLLMSuggest(ctx context.Context, opts Options) error {
	if opts.Capture == "" || opts.APIPrefix == "" {
		return fmt.Errorf("required flags for --llm-suggest: -i, -p")
	}

	paramRegex := opts.ParamRegex
	if paramRegex == "" {
		paramRegex = `[0-9]+`
	}
	numericRe, err := regexp.Compile("^" + paramRegex + "$")
	if err != nil {
		return fmt.Errorf("invalid path parameter regex: %w", err)
	}

	doc, err := schema.Load(opts.Schema)
	if err != nil {
		return err
	}

	candidates := findSlugMergeCandidates(doc.XPathTemplates, numericRe)
	if len(candidates) == 0 {
		fmt.Println("curate --llm-suggest: no slug merge candidates found")
		return saveSuggestionsFile(opts.SuggestionsOut, &SuggestionsFile{})
	}

	maxGroups := opts.MaxGroups
	if maxGroups < 1 {
		maxGroups = 50
	}
	if len(candidates) > maxGroups {
		candidates = candidates[:maxGroups]
		fmt.Printf("curate --llm-suggest: limiting to %d candidate groups\n", maxGroups)
	}

	reader, err := captureopen.OpenReader(opts.Capture, opts.Format, nil)
	if err != nil {
		return err
	}

	var client openai.Client
	if opts.EmitPromptsDir == "" {
		client, err = newCurateClient(opts)
		if err != nil {
			return err
		}
	}

	outPath := opts.SuggestionsOut
	if outPath == "" {
		outPath = filepath.Join(filepath.Dir(opts.Schema), "template-suggestions.yaml")
	}

	var suggestions []TemplateSuggestion
	for i, group := range candidates {
		want := make(map[string]struct{}, len(group.entries))
		replaces := make([]string, 0, len(group.entries))
		for _, e := range group.entries {
			want[e.path] = struct{}{}
			replaces = append(replaces, formatTemplateEntry(e))
		}

		samplePaths, err := collectSamplePaths(reader, opts.APIPrefix, want, 5)
		if err != nil {
			return err
		}

		userPrompt := buildTemplateSuggestionUserPrompt(group, samplePaths)
		if opts.EmitPromptsDir != "" {
			if err := os.MkdirAll(opts.EmitPromptsDir, 0o755); err != nil {
				return err
			}
			name := fmt.Sprintf("template-suggest-%03d.txt", i+1)
			if err := os.WriteFile(filepath.Join(opts.EmitPromptsDir, name), []byte(templateSuggestionSystemPrompt+"\n\n---\n\n"+userPrompt), 0o644); err != nil {
				return err
			}
			continue
		}

		fmt.Printf("[%d/%d] suggesting merge for %d paths\n", i+1, len(candidates), len(group.entries))
		llm, err := callTemplateSuggestionLLM(ctx, client, opts.Model, templateSuggestionSystemPrompt, userPrompt)
		if err != nil {
			return fmt.Errorf("LLM for group %d: %w", i+1, err)
		}
		if strings.TrimSpace(llm.ProposedTemplate) == "" {
			continue
		}

		proposed := llm.ProposedTemplate
		if !strings.HasPrefix(proposed, "/") {
			proposed = "/" + strings.TrimPrefix(proposed, "/")
		}
		if group.entries[0].ignore {
			proposed = "ignore:" + strings.TrimPrefix(proposed, "ignore:")
		}

		if !validateProposedMatchesGroup(proposed, group.entries) {
			fmt.Printf("warning: skipping group %q: proposed %q does not match all paths\n", group.prefix, proposed)
			continue
		}

		suggestions = append(suggestions, TemplateSuggestion{
			ProposedTemplate: proposed,
			Replaces:         replaces,
			Confidence:       llm.Confidence,
			SamplePaths:      samplePaths,
			Reason:           llm.Reason,
		})
	}

	if opts.EmitPromptsDir != "" {
		fmt.Printf("Wrote %d prompts to %s\n", len(candidates), opts.EmitPromptsDir)
		return nil
	}

	if err := saveSuggestionsFile(outPath, &SuggestionsFile{Suggestions: suggestions}); err != nil {
		return err
	}
	fmt.Printf("curate --llm-suggest: wrote %d suggestions to %s\n", len(suggestions), outPath)
	return nil
}
