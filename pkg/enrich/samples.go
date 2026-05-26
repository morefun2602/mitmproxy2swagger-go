package enrich

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/swaggerutil"
)

type endpointKey struct {
	pathTemplate string
	method       string
}

type RequestSample struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	RequestHeaders map[string][]string `json:"requestHeaders,omitempty"`
	RequestBody    string            `json:"requestBody,omitempty"`
	ResponseStatus int               `json:"responseStatus"`
	ResponseReason string            `json:"responseReason,omitempty"`
	ResponseBody   string            `json:"responseBody,omitempty"`
}

func endpointSamples(
	doc *schema.Document,
	reader capture.Reader,
	apiPrefix string,
	maxSamples int,
	mode RedactMode,
) (map[endpointKey][]RequestSample, error) {
	apiPrefix = strings.TrimSuffix(apiPrefix, "/")
	type pathMatch struct {
		template string
		re       *regexp.Regexp
	}
	var paths []pathMatch
	for _, item := range doc.Paths {
		tmpl := fmt.Sprint(item.Key)
		paths = append(paths, pathMatch{template: tmpl, re: pathToRegex(tmpl)})
	}

	out := make(map[endpointKey][]RequestSample)
	err := reader.Each(func(req capture.CapturedRequest) error {
		matchURL, ok := req.MatchingURL(apiPrefix)
		if !ok {
			return nil
		}
		path := stripAPIPrefix(stripQueryString(matchURL), apiPrefix)
		method := strings.ToLower(req.Method())

		var matchedTemplate string
		for _, pm := range paths {
			if !pm.re.MatchString(path) {
				continue
			}
			matchedTemplate = pm.template
			break
		}
		if matchedTemplate == "" {
			return nil
		}

		key := endpointKey{pathTemplate: matchedTemplate, method: method}
		if len(out[key]) >= maxSamples {
			return nil
		}

		reqBody := req.RequestBody()
		respBody := req.ResponseBody()
		if len(reqBody) > 0 {
			reqBody = []byte(truncateBody(reqBody, mode))
		}
		if len(respBody) > 0 {
			respBody = []byte(truncateBody(respBody, mode))
		}

		out[key] = append(out[key], RequestSample{
			URL:            matchURL,
			Method:         method,
			RequestHeaders: redactHeaders(req.RequestHeaders(), mode),
			RequestBody:    string(reqBody),
			ResponseStatus: req.ResponseStatusCode(),
			ResponseReason: req.ResponseReason(),
			ResponseBody:   string(respBody),
		})
		return nil
	})
	return out, err
}

func truncateBody(body []byte, mode RedactMode) string {
	var parsed any
	if err := json.Unmarshal(body, &parsed); err == nil {
		limited := swaggerutil.LimitExampleSize(parsed)
		redactValue(limited, mode)
		data, err := json.Marshal(limited)
		if err == nil {
			return string(data)
		}
	}
	return redactBodyText(body, mode)
}
