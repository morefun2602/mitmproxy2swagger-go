package pass

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture"
	captureopen "github.com/morefun2602/mitmproxy2swagger-go/internal/capture/open"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/schema"
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

	runner := newPassRunner(doc, opts, apiPrefix, paramRegex)
	if err := reader.Each(runner.processRequest); err != nil {
		return err
	}

	runner.finalizeDiscovery()
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
	return schema.New(inputPath), nil
}

type passRunner struct {
	doc              *schema.Document
	opts             Options
	apiPrefix        string
	paramRegex       *regexp.Regexp
	pathTemplates    []string
	pathRegexes      []*regexp.Regexp
	newPathTemplates []string
	seenNew          map[string]struct{}
}

func newPassRunner(doc *schema.Document, opts Options, apiPrefix string, paramRegex *regexp.Regexp) *passRunner {
	pathTemplates := doc.PathTemplates()
	pathRegexes := make([]*regexp.Regexp, len(pathTemplates))
	for i, tmpl := range pathTemplates {
		pathRegexes[i] = pathToRegex(tmpl)
	}
	return &passRunner{
		doc:           doc,
		opts:          opts,
		apiPrefix:     apiPrefix,
		paramRegex:    paramRegex,
		pathTemplates: pathTemplates,
		pathRegexes:   pathRegexes,
		seenNew:       make(map[string]struct{}),
	}
}

func (r *passRunner) processRequest(req capture.CapturedRequest) error {
	matchURL, ok := req.MatchingURL(r.apiPrefix)
	if !ok {
		return nil
	}

	path := stripAPIPrefix(stripQueryString(matchURL), r.apiPrefix)
	pathTemplateIndex := r.matchPathTemplate(path)
	if pathTemplateIndex < 0 {
		r.discoverPath(path)
		return nil
	}

	return r.materializeEndpoint(req, matchURL, path, pathTemplateIndex)
}

func (r *passRunner) matchPathTemplate(path string) int {
	for i, re := range r.pathRegexes {
		if re.MatchString(path) {
			return i
		}
	}
	return -1
}

func (r *passRunner) finalizeDiscovery() {
	sort.Strings(r.newPathTemplates)
	suggestions := buildSuggestedTemplates(r.newPathTemplates, r.paramRegex, r.opts.SuppressParams)
	r.doc.XPathTemplates = append(r.doc.XPathTemplates, suggestions...)
	r.doc.XPathTemplates = schema.FilterXPathTemplates(r.doc.XPathTemplates, r.doc.Paths)
	r.doc.XPathTemplates = schema.DedupeStrings(r.doc.XPathTemplates)
}
