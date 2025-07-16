package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	filePath := flag.String("file", "", "Path to code file")
	text := flag.String("text", "", "Raw code string")
	lang := flag.String("lang", "", "Language for syntax highlighting (required, or inferred from --file)")
	out := flag.String("out", "", "Output file (PNG). If omitted, defaults to ./screenshot.png")
	theme := flag.String("theme", "dracula", "Chroma theme")
	font := flag.String("font", "FiraCode-Regular.ttf", "Font file (TTF)")
	fontsize := flag.Float64("fontsize", 18, "Font size")

	flag.Parse()

	var code string
	var err error

	if *filePath != "" {
		data, err := ioutil.ReadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		code = string(data)
		// Infer lang from file extension if not set
		if *lang == "" {
			*lang = InferLang(*filePath)
		}
	} else if *text != "" {
		code = *text
	} else {
		// Try to read from stdin
		stdinBytes, _ := ioutil.ReadAll(os.Stdin)
		code = string(stdinBytes)
		if strings.TrimSpace(code) == "" {
			fmt.Fprintf(os.Stderr, "No input supplied. Use --file, --text, or pipe input.\n")
			os.Exit(1)
		}
	}

	if *lang == "" {
		fmt.Fprintf(os.Stderr, "--lang required (cannot infer)\n")
		os.Exit(1)
	}

	imgBytes, err := RenderCodeImage(code, *lang, *theme, *font, *fontsize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to render image: %v\n", err)
		os.Exit(1)
	}

	outputFile := *out
	if outputFile == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get current directory: %v\n", err)
			os.Exit(1)
		}
		outputFile = filepath.Join(cwd, "codeshot.png")
	}

	if err := ioutil.WriteFile(outputFile, imgBytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write image: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Image written to", outputFile)
}
