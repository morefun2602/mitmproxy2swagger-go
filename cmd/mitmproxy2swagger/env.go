package main

func envOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
