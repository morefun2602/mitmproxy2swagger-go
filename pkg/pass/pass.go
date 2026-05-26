package pass

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
	captureopen "github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/open"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/swaggerutil"
	"github.com/vmihailenco/msgpack/v5"
)

// Options configures a single Pass run.
type Options struct {
	Input          string
	Output         string
	APIPrefix      string
	Examples       bool
	Headers        bool
	Format         string
	ParamRegex     string
	SuppressParams bool
	Reader         capture.Reader
}

func Run(opts Options) error {
	paramRegex, err := regexp.Compile("^" + opts.ParamRegex + "$")
	if err != nil {
		return fmt.Errorf("invalid path parameter regex: %w", err)
	}

	reader := opts.Reader
	if reader == nil {
		reader, err = captureopen.OpenReader(opts.Input, opts.Format, nil)
		if err != nil {
			return err
		}
	}

	doc, err := loadDocument(opts.Output, opts.Input)
	if err != nil {
		return err
	}

	apiPrefix := strings.TrimSuffix(opts.APIPrefix, "/")
	doc.EnsureDefaults(apiPrefix)

	pathTemplates := doc.PathTemplates()
	schema.SortPathTemplates(pathTemplates)
	pathRegexes := make([]*regexp.Regexp, len(pathTemplates))
	for i, tmpl := range pathTemplates {
		pathRegexes[i] = pathToRegex(tmpl)
	}

	newPathTemplates := make([]string, 0)
	seenNew := make(map[string]struct{})

	err = reader.Each(func(req capture.CapturedRequest) error {
		matchURL, ok := req.MatchingURL(apiPrefix)
		if !ok {
			return nil
		}

		method := strings.ToLower(req.Method())
		path := stripAPIPrefix(stripQueryString(matchURL), apiPrefix)
		status := req.ResponseStatusCode()

		pathTemplateIndex := -1
		for i, re := range pathRegexes {
			if re.MatchString(path) {
				pathTemplateIndex = i
				break
			}
		}
		if pathTemplateIndex < 0 {
			if _, exists := seenNew[path]; exists {
				return nil
			}
			seenNew[path] = struct{}{}
			newPathTemplates = append(newPathTemplates, path)
			return nil
		}

		pathTemplate := pathTemplates[pathTemplateIndex]
		doc.SetPathIfNotExists(pathTemplate, map[string]any{})

		ops, ok := doc.PathOperations(pathTemplate)
		if !ok {
			return nil
		}

		schema.SetKeyIfNotExists(ops, method, map[string]any{
			"summary":   swaggerutil.PathTemplateToEndpointName(method, pathTemplate),
			"responses": map[string]any{},
		})

		methodDoc, _ := ops[method].(map[string]any)

		if opts.Headers {
			headerParams := swaggerutil.RequestToHeaders(req.RequestHeaders())
			if len(headerParams) > 0 {
				schema.SetKeyIfNotExists(methodDoc, "parameters", headerParams)
			}
		}
		params := swaggerutil.URLToParams(matchURL, pathTemplate)
		if len(params) > 0 {
			schema.SetKeyIfNotExists(methodDoc, "parameters", params)
		}

		if method != "get" && method != "head" {
			if body := req.RequestBody(); body != nil {
				if bodyVal, contentType, ok := parseRequestBody(body); ok {
					content := map[string]any{
						"content": map[string]any{
							contentType: map[string]any{
								"schema": swaggerutil.ValueToSchema(bodyVal),
							},
						},
					}
					if opts.Examples {
						content["content"].(map[string]any)[contentType].(map[string]any)["example"] = swaggerutil.LimitExampleSize(bodyVal)
					}
					schema.SetKeyIfNotExists(methodDoc, "requestBody", content)
				}
			}
		}

		if respBody := req.ResponseBody(); respBody != nil {
			if parsed, contentType, ok := parseResponseBody(respBody); ok {
				resp := map[string]any{
					"description": req.ResponseReason(),
					"content": map[string]any{
						contentType: map[string]any{
							"schema": swaggerutil.ValueToSchema(parsed),
						},
					},
				}
				if opts.Examples {
					resp["content"].(map[string]any)[contentType].(map[string]any)["example"] = swaggerutil.LimitExampleSize(parsed)
				}
				if opts.Headers {
					if h := swaggerutil.ResponseToHeaders(req.ResponseHeaders()); len(h) > 0 {
						resp["headers"] = h
					}
				}
				responses, _ := methodDoc["responses"].(map[string]any)
				if responses == nil {
					responses = map[string]any{}
					methodDoc["responses"] = responses
				}
				schema.SetKeyIfNotExists(responses, fmt.Sprint(status), resp)
			}
		}

		responses, _ := methodDoc["responses"].(map[string]any)
		if responses != nil && len(responses) == 0 {
			responses["200"] = map[string]any{
				"description": "OK",
				"content":     map[string]any{},
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(newPathTemplates)
	suggestions := buildSuggestedTemplates(newPathTemplates, paramRegex, opts.SuppressParams)
	doc.XPathTemplates = append(doc.XPathTemplates, suggestions...)
	doc.XPathTemplates = schema.FilterXPathTemplates(doc.XPathTemplates, doc.Paths)
	doc.XPathTemplates = schema.DedupeStrings(doc.XPathTemplates)
	schema.SortPathTemplates(doc.XPathTemplates)

	return doc.Save(opts.Output)
}

func loadDocument(outputPath, inputPath string) (*schema.Document, error) {
	doc, err := schema.Load(outputPath)
	if err == nil {
		return doc, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	fmt.Println("No existing swagger file found. Creating new one.")
	return schema.New(captureTitle(inputPath)), nil
}

func captureTitle(inputPath string) string {
	clean := filepath.Clean(inputPath)
	if rel, err := filepath.Rel(".", clean); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		clean = rel
	}
	return filepath.ToSlash(clean)
}

func buildSuggestedTemplates(paths []string, paramRegex *regexp.Regexp, suppressParams bool) []string {
	var out []string
	seen := make(map[string]struct{})

	isParam := func(segment string) bool {
		return paramRegex.MatchString(segment)
	}

	for _, path := range paths {
		segments := strings.Split(path, "/")
		hasParam := false
		for _, seg := range segments {
			if seg != "" && isParam(seg) {
				hasParam = true
				break
			}
		}

		if hasParam {
			newSegments := make([]string, 0, len(segments))
			paramID := 0
			for _, seg := range segments {
				if seg != "" && isParam(seg) {
					name := "id"
					if paramID > 0 {
						name = fmt.Sprintf("id%d", paramID)
					}
					newSegments = append(newSegments, "{"+name+"}")
					paramID++
				} else {
					newSegments = append(newSegments, seg)
				}
			}
			suggested := strings.Join(newSegments, "/")
			entry := "ignore:" + suggested
			if _, ok := seen[entry]; !ok {
				seen[entry] = struct{}{}
				out = append(out, entry)
			}
		}

		if !hasParam || !suppressParams {
			entry := "ignore:" + path
			if _, ok := seen[entry]; !ok {
				seen[entry] = struct{}{}
				out = append(out, entry)
			}
		}
	}
	return out
}

func parseRequestBody(body []byte) (any, string, bool) {
	var jsonVal any
	if err := json.Unmarshal(body, &jsonVal); err == nil {
		return jsonVal, "application/json", true
	}
	var msgpackVal any
	if err := msgpack.Unmarshal(body, &msgpackVal); err == nil {
		return msgpackVal, "application/msgpack", true
	}
	form, err := url.ParseQuery(string(body))
	if err == nil && len(form) > 0 {
		out := map[string]any{}
		for key, values := range form {
			if len(values) > 0 {
				out[key] = values[0]
			}
		}
		if len(out) > 0 {
			return out, "application/x-www-form-urlencoded", true
		}
	}
	return nil, "", false
}

func parseResponseBody(body []byte) (any, string, bool) {
	var jsonVal any
	if err := json.Unmarshal(body, &jsonVal); err == nil {
		return jsonVal, "application/json", true
	}
	var msgpackVal any
	if err := msgpack.Unmarshal(body, &msgpackVal); err == nil {
		return msgpackVal, "application/msgpack", true
	}
	return nil, "", false
}

func pathToRegex(path string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(path)
	escaped = strings.ReplaceAll(escaped, "\\{", "(?P<")
	escaped = strings.ReplaceAll(escaped, "\\}", ">[^/]+)")
	escaped = strings.ReplaceAll(escaped, "\\*", ".*")
	re, err := regexp.Compile("^" + escaped + "$")
	if err != nil {
		return regexp.MustCompile("^$")
	}
	return re
}

func stripQueryString(rawURL string) string {
	if i := strings.Index(rawURL, "?"); i >= 0 {
		return rawURL[:i]
	}
	return rawURL
}

func stripAPIPrefix(rawURL, prefix string) string {
	return strings.TrimPrefix(rawURL, prefix)
}
