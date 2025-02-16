package utility

func JsonError(message string) string {
	return `{"error": "` + message + `"}`
}
