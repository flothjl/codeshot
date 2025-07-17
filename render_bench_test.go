package main

import (
	"fmt"
	"strings"
	"testing"
)

// Helper: Generate N lines of Go code
func generateGoCode(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		fmt.Fprintf(&b, "fmt.Println(\"This is line %d\")\n", i)
	}
	return b.String()
}

func BenchmarkRenderCodeImage(b *testing.B) {
	codetests := []struct {
		lines int
	}{
		{20},
		{40},
		{80},
		{100},
	}
	for _, bm := range codetests {
		b.Run(fmt.Sprintf("%d_lines", bm.lines), func(b *testing.B) {
			code := generateGoCode(bm.lines)
			for i := 0; i < b.N; i++ {
				_, err := RenderCodeImage(code, "go", "dracula", "fonts/FiraCode-Regular.ttf", 18)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
