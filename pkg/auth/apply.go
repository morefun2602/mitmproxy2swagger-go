package auth

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

// ApplyOptions configures auth apply.
type ApplyOptions struct {
	Schema           string
	Observations     string
	Force            bool
	TagAuthEndpoints bool
	CookieName       string
}

type resolvedCredential struct {
	observed ObservedCredential
	schemeID string
}

var schemeIDSanitizer = regexp.MustCompile(`[^a-zA-Z0-9]+`)

var preferredCookieNames = []string{
	"sessionid", "jsessionid", "session", "sid",
}

// RunApply writes components.securitySchemes and root security from verified observations.
func RunApply(opts ApplyOptions) error {
	if opts.Schema == "" {
		return fmt.Errorf("schema path is required")
	}
	if opts.Observations == "" {
		if opts.Schema == "" {
			return fmt.Errorf("observations path is required")
		}
		opts.Observations = filepath.Join(filepath.Dir(opts.Schema), "auth-observations.yaml")
	}

	obs, err := LoadObservationsFile(opts.Observations)
	if err != nil {
		return err
	}
	if len(obs.Verified) == 0 {
		return fmt.Errorf("%s: verified is empty; edit the file after curl/probe (e.g. verified: [cookie])", opts.Observations)
	}

	resolved, err := resolveVerified(obs)
	if err != nil {
		return err
	}

	doc, err := schema.Load(opts.Schema)
	if err != nil {
		return err
	}

	schemes := buildSecuritySchemes(resolved, opts.CookieName)
	if err := mergeSecuritySchemes(doc, schemes, opts.Force); err != nil {
		return err
	}

	combination := obs.Combination
	if combination == "" {
		if len(resolved) == 1 {
			combination = "single"
		} else {
			combination = "and"
		}
	}
	security, err := buildSecurityRequirements(schemeIDs(resolved), combination)
	if err != nil {
		return err
	}
	doc.Security = security

	if opts.TagAuthEndpoints {
		tagSuggestedAuthPaths(doc, obs.SuggestedAuthPaths)
	}

	return doc.Save(opts.Schema)
}

func resolveVerified(obs *ObservationsFile) ([]resolvedCredential, error) {
	out := make([]resolvedCredential, 0, len(obs.Verified))
	for _, v := range obs.Verified {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" {
			continue
		}
		cred, ok := matchVerified(obs.Observed, v)
		if !ok {
			return nil, fmt.Errorf("verified entry %q does not match any observed credential", v)
		}
		out = append(out, resolvedCredential{
			observed: cred,
			schemeID: schemeIDFor(cred),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("verified has no usable entries")
	}
	return out, nil
}

func matchVerified(observed []ObservedCredential, verified string) (ObservedCredential, bool) {
	wantKind, wantName := verified, ""
	if strings.HasPrefix(verified, "header:") {
		wantKind = kindHeader
		wantName = strings.TrimSpace(strings.ToLower(strings.TrimPrefix(verified, "header:")))
	}
	for _, o := range observed {
		kind := strings.ToLower(o.Kind)
		name := strings.ToLower(o.Name)
		switch wantKind {
		case kindCookie, "cookies":
			if kind == kindCookie {
				return o, true
			}
		case kindBearer:
			if kind == kindBearer {
				return o, true
			}
		case kindAPIKey, "apikey", "api-key":
			if kind == kindAPIKey {
				return o, true
			}
		default:
			if kind == kindHeader || kind == kindAPIKey {
				if wantName == "" {
					if name == verified || name == strings.TrimPrefix(verified, "header:") {
						return o, true
					}
				} else if name == wantName {
					return o, true
				}
			}
		}
	}
	return ObservedCredential{}, false
}

func schemeIDFor(c ObservedCredential) string {
	switch strings.ToLower(c.Kind) {
	case kindCookie:
		return "sessionCookie"
	case kindBearer:
		return "bearerAuth"
	case kindAPIKey:
		return "apiKey_" + sanitizeSchemeID(c.Name)
	default:
		return "header_" + sanitizeSchemeID(c.Name)
	}
}

func sanitizeSchemeID(name string) string {
	s := schemeIDSanitizer.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "custom"
	}
	if unicode.IsDigit(rune(s[0])) {
		return "h_" + s
	}
	return s
}

func buildSecuritySchemes(resolved []resolvedCredential, cookieNameOverride string) map[string]map[string]any {
	out := make(map[string]map[string]any, len(resolved))
	for _, r := range resolved {
		out[r.schemeID] = schemeDefinition(r.observed, cookieNameOverride)
	}
	return out
}

func schemeDefinition(c ObservedCredential, cookieNameOverride string) map[string]any {
	switch strings.ToLower(c.Kind) {
	case kindCookie:
		name := primaryCookieName(c.SampleValue, cookieNameOverride)
		desc := "Session cookies observed in capture."
		if c.SampleValue != "" && c.SampleValue != "[REDACTED]" {
			desc = fmt.Sprintf("Session cookies observed in capture (includes %s). Primary OpenAPI cookie name: %s.", c.SampleValue, name)
		}
		return map[string]any{
			"type":        "apiKey",
			"in":          "cookie",
			"name":        name,
			"description": desc,
		}
	case kindBearer:
		return map[string]any{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
			"description":  "Bearer token in Authorization header.",
		}
	default:
		return map[string]any{
			"type":        "apiKey",
			"in":          "header",
			"name":        c.Name,
			"description": fmt.Sprintf("API key header %q observed in capture.", c.Name),
		}
	}
}

func primaryCookieName(sampleValue, override string) string {
	if override != "" {
		return override
	}
	names := cookieNamesFromSample(sampleValue)
	for _, pref := range preferredCookieNames {
		for _, n := range names {
			if strings.EqualFold(n, pref) {
				return n
			}
		}
	}
	if len(names) > 0 {
		return names[0]
	}
	return "session"
}

func cookieNamesFromSample(sample string) []string {
	if sample == "" || sample == "[REDACTED]" {
		return nil
	}
	var names []string
	for _, part := range strings.Split(sample, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, _, _ := strings.Cut(part, "=")
		n = strings.TrimSpace(n)
		if n != "" {
			names = append(names, n)
		}
	}
	return names
}

func mergeSecuritySchemes(doc *schema.Document, schemes map[string]map[string]any, force bool) error {
	if doc.Components == nil {
		doc.Components = map[string]any{}
	}
	existing, _ := doc.Components["securitySchemes"].(map[string]any)
	if existing == nil {
		existing = map[string]any{}
	}
	for id, def := range schemes {
		if force {
			existing[id] = def
			continue
		}
		schema.SetKeyIfNotExists(existing, id, def)
	}
	doc.Components["securitySchemes"] = existing
	return nil
}

func buildSecurityRequirements(schemeIDs []string, combination string) ([]any, error) {
	switch combination {
	case "single":
		if len(schemeIDs) != 1 {
			return nil, fmt.Errorf("combination single requires exactly one verified credential, got %d", len(schemeIDs))
		}
		return []any{map[string]any{schemeIDs[0]: []any{}}}, nil
	case "and":
		req := make(map[string]any, len(schemeIDs))
		for _, id := range schemeIDs {
			req[id] = []any{}
		}
		return []any{req}, nil
	case "or":
		out := make([]any, len(schemeIDs))
		for i, id := range schemeIDs {
			out[i] = map[string]any{id: []any{}}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unknown combination %q (use single, and, or)", combination)
	}
}

func schemeIDs(resolved []resolvedCredential) []string {
	ids := make([]string, len(resolved))
	for i, r := range resolved {
		ids[i] = r.schemeID
	}
	return ids
}

func tagSuggestedAuthPaths(doc *schema.Document, suggested []SuggestedAuthPath) {
	if len(suggested) == 0 {
		return
	}
	want := make(map[string]struct{}, len(suggested))
	for _, s := range suggested {
		if s.Path != "" {
			want[s.Path] = struct{}{}
		}
	}
	for _, item := range doc.Paths {
		pathKey := strings.TrimPrefix(fmt.Sprint(item.Key), "ignore:")
		if _, ok := want[pathKey]; !ok {
			continue
		}
		ops, ok := item.Value.(map[string]any)
		if !ok {
			continue
		}
		for method, raw := range ops {
			if !isHTTPMethod(method) {
				continue
			}
			op, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			addAuthTag(op)
		}
	}
}

func isHTTPMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}

func addAuthTag(op map[string]any) {
	const authTag = "auth"
	seen := map[string]struct{}{authTag: {}}
	out := []any{authTag}
	switch tags := op["tags"].(type) {
	case []any:
		for _, t := range tags {
			s := fmt.Sprint(t)
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	case []string:
		for _, s := range tags {
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	op["tags"] = out
}
