package main

import (
	"os"
	"path/filepath"
	"strings"
)

// InferLang guesses the language from the file extension.
func InferLang(filename string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "js":
		return "javascript"
	case "py":
		return "python"
	case "go":
		return "go"
	case "ts":
		return "typescript"
	case "rs":
		return "rust"
	case "java":
		return "java"
	case "c":
		return "c"
	case "cpp", "cc", "cxx", "h", "hpp":
		return "cpp"
	case "sh", "bash":
		return "bash"
	case "md":
		return "markdown"
	case "html", "htm":
		return "html"
	case "css":
		return "css"
	case "json":
		return "json"
	default:
		return ext
	}
}

func TempFileWithExt(ext string) string {
	f, _ := os.CreateTemp("", "codeshot-*"+ext)
	f.Close()
	return f.Name()
}
