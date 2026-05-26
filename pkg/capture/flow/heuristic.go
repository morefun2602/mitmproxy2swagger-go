package flow

import (
	"os"
	"strings"
	"unicode"
)

// DumpHeuristic scores how likely path is a mitmproxy Flow dump (higher = more likely).
func DumpHeuristic(path string) int {
	val := 0
	if strings.Contains(path, "flow") {
		val++
	}
	if strings.Contains(path, "mitmproxy") {
		val++
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
	compact := strings.NewReplacer("\r", "", "\n", "").Replace(string(data))
	if !isPrintable(compact) {
		val += 50
	}
	if data[0] >= '0' && data[0] <= '9' {
		val += 5
	}
	if strings.Contains(string(data), "status_code") {
		val += 5
	}
	if strings.Contains(string(data), "regular") {
		val += 10
	}
	return val
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
