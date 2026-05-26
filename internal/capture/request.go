package capture

// ProgressFunc reports capture read progress in the range [0, 1].
type ProgressFunc func(progress float64)

// CapturedRequest is a single HTTP request/response pair from a Capture file.
type CapturedRequest interface {
	URL() string
	// MatchingURL returns the request URL when prefix matches; ok is false otherwise.
	MatchingURL(prefix string) (url string, ok bool)
	Method() string
	RequestHeaders() map[string][]string
	RequestBody() []byte
	ResponseStatusCode() int
	ResponseReason() string
	ResponseHeaders() map[string][]string
	ResponseBody() []byte
}

// Reader reads CapturedRequest values from a Capture file.
type Reader interface {
	Name() string
	Each(fn func(CapturedRequest) error) error
}
