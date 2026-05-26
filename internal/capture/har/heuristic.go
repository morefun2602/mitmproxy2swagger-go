package har

import (
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ArchiveHeuristic scores how likely path is a HAR archive (higher = more likely).
// Matches alufers/mitmproxy2swagger v0.14.0 har_archive_heuristic.
func ArchiveHeuristic(path string) int {
	val := 0
	if strings.HasSuffix(path, ".har") {
		val += 25
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return val
	}
	if len(data) > 2048 {
		data = data[:2048]
	}
	if len(data) == 0 {
		return val
	}

	if isPrintableASCII(data) {
		val += 25
	}
	if data[0] == '{' {
		val += 23
	}
	if bytesContains(data, `"WebInspector"`) || bytesContains(data, `"Firefox"`) {
		val += 15
	}
	if bytesContains(data, `"entries"`) {
		val += 15
	}
	if bytesContains(data, `"version"`) {
		val += 15
	}
	return val
}

func isPrintableASCII(data []byte) bool {
	s := strings.NewReplacer("\r", "", "\n", "").Replace(string(data))
	for _, r := range s {
		if r == utf8.RuneError {
			continue
		}
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func bytesContains(data []byte, needle string) bool {
	return strings.Contains(string(data), needle)
}
