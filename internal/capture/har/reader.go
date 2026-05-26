package har

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture"
)

type headerKV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type postData struct {
	Text string `json:"text"`
}

type harRequest struct {
	Method  string     `json:"method"`
	URL     string     `json:"url"`
	Headers []headerKV `json:"headers"`
	PostData *postData `json:"postData"`
}

type harContent struct {
	Text     string `json:"text"`
	Encoding string `json:"encoding"`
}

type harResponse struct {
	Status     int        `json:"status"`
	StatusText string     `json:"statusText"`
	Headers    []headerKV `json:"headers"`
	Content    *harContent `json:"content"`
}

type harEntry struct {
	Request  harRequest  `json:"request"`
	Response harResponse `json:"response"`
}

type harDocument struct {
	Log struct {
		Entries []harEntry `json:"entries"`
	} `json:"log"`
}

type capturedRequest struct {
	entry harEntry
}

func (r capturedRequest) URL() string {
	return r.entry.Request.URL
}

func (r capturedRequest) MatchingURL(prefix string) (string, bool) {
	if strings.HasPrefix(r.entry.Request.URL, prefix) {
		return r.entry.Request.URL, true
	}
	return "", false
}

func (r capturedRequest) Method() string {
	return r.entry.Request.Method
}

func (r capturedRequest) RequestHeaders() map[string][]string {
	return headersFromKV(r.entry.Request.Headers)
}

func (r capturedRequest) RequestBody() []byte {
	if r.entry.Request.PostData == nil {
		return nil
	}
	text := r.entry.Request.PostData.Text
	if text == "" {
		return nil
	}
	return []byte(text)
}

func (r capturedRequest) ResponseStatusCode() int {
	return r.entry.Response.Status
}

func (r capturedRequest) ResponseReason() string {
	return r.entry.Response.StatusText
}

func (r capturedRequest) ResponseHeaders() map[string][]string {
	return headersFromKV(r.entry.Response.Headers)
}

func (r capturedRequest) ResponseBody() []byte {
	content := r.entry.Response.Content
	if content == nil || content.Text == "" {
		return nil
	}
	if content.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(content.Text)
		if err != nil {
			return nil
		}
		if !utf8Valid(decoded) {
			return nil
		}
		return decoded
	}
	return []byte(content.Text)
}

func headersFromKV(items []headerKV) map[string][]string {
	headers := make(map[string][]string, len(items))
	for _, kv := range items {
		headers[kv.Name] = append(headers[kv.Name], kv.Value)
	}
	return headers
}

func utf8Valid(b []byte) bool {
	return utf8.Valid(b)
}

// CaptureReader reads CapturedRequest values from a HAR archive file.
type CaptureReader struct {
	path     string
	progress capture.ProgressFunc
}

func NewCaptureReader(path string, progress capture.ProgressFunc) *CaptureReader {
	return &CaptureReader{path: path, progress: progress}
}

func (r *CaptureReader) Name() string {
	return "har"
}

func (r *CaptureReader) Each(fn func(capture.CapturedRequest) error) error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return err
	}

	var doc harDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse HAR: %w", err)
	}

	entries := doc.Log.Entries
	total := len(entries)
	for i, entry := range entries {
		if r.progress != nil && total > 0 {
			r.progress(float64(i+1) / float64(total))
		}
		if err := fn(capturedRequest{entry: entry}); err != nil {
			return err
		}
	}
	return nil
}

// Ensure CaptureReader implements capture.Reader.
var _ capture.Reader = (*CaptureReader)(nil)
