package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/fogleman/gg"
)

const (
	horizontalPadding = 32
	verticalPadding   = 32
	lineSpacingFactor = 1.4
	tabWidthSpaces    = 4

	windowBarHeight = 40
	cornerRadius    = 20

	trafficLightR   = 8
	trafficLightGap = 16
)

func toNRGBA(src image.Image) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
	return dst
}

// Sanitize code input to avoid missing glyphs and hidden Unicode
func sanitizeCode(code string) string {
	if strings.HasPrefix(code, "\uFEFF") {
		code = strings.TrimPrefix(code, "\uFEFF")
	}
	code = strings.ReplaceAll(code, "\t", strings.Repeat(" ", tabWidthSpaces))
	code = strings.ReplaceAll(code, "\u00A0", " ")
	code = strings.ReplaceAll(code, "\u200B", "")
	code = strings.ReplaceAll(code, "\r\n", "\n")

	var sanitized strings.Builder
	for _, r := range code {
		switch {
		case r == '\n' || r == '\r':
			sanitized.WriteRune(r)
		case r >= 32 && r <= 126:
			sanitized.WriteRune(r)
		case r >= 0xA0 && r <= 0x17F:
			sanitized.WriteRune(r)
		default:
			sanitized.WriteRune(' ')
		}
	}
	return sanitized.String()
}

// Post-process: mask for rounded corners
func makeRoundedRectMask(width, height int, radius float64) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			inside := false
			switch {
			case x < int(radius) && y < int(radius): // top-left
				inside = math.Hypot(float64(x)-radius, float64(y)-radius) <= radius
			case x >= width-int(radius) && y < int(radius): // top-right
				inside = math.Hypot(float64(x)-float64(width-1)+radius, float64(y)-radius) <= radius
			case x < int(radius) && y >= height-int(radius): // bottom-left
				inside = math.Hypot(float64(x)-radius, float64(y)-float64(height-1)+radius) <= radius
			case x >= width-int(radius) && y >= height-int(radius): // bottom-right
				inside = math.Hypot(float64(x)-float64(width-1)+radius, float64(y)-float64(height-1)+radius) <= radius
			default:
				inside = true
			}
			if inside {
				mask.SetAlpha(x, y, color.Alpha{A: 255})
			} else {
				mask.SetAlpha(x, y, color.Alpha{A: 0})
			}
		}
	}
	return mask
}

func RenderCodeImage(code, lang, theme, font string, fontsize float64) ([]byte, error) {
	code = sanitizeCode(code)
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(code)
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return nil, fmt.Errorf("chroma tokenize error: %w", err)
	}

	style := styles.Get(theme)
	if style == nil {
		style = styles.Fallback
	}

	// --- Tokenize lines: split tokens only on \n ---
	var lines [][]chroma.Token
	var currentLine []chroma.Token
	for token := iterator(); token != chroma.EOF; token = iterator() {
		val := token.Value
		for {
			idx := strings.IndexByte(val, '\n')
			if idx == -1 {
				if val != "" {
					currentLine = append(currentLine, chroma.Token{Type: token.Type, Value: val})
				}
				break
			}
			if idx > 0 {
				currentLine = append(currentLine, chroma.Token{Type: token.Type, Value: val[:idx]})
			}
			lines = append(lines, currentLine)
			currentLine = []chroma.Token{}
			val = val[idx+1:]
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}
	if len(lines) == 0 {
		lines = append(lines, []chroma.Token{})
	}

	// Load font face ONCE and reuse
	fontFace, err := gg.LoadFontFace(font, fontsize)
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}

	// Measure maximum width
	measureDC := gg.NewContext(100, 100)
	measureDC.SetFontFace(fontFace)

	maxWidth := 0.0
	for _, line := range lines {
		w := 0.0
		for _, tok := range line {
			tw, _ := measureDC.MeasureString(tok.Value)
			w += tw
		}
		if w > maxWidth {
			maxWidth = w
		}
	}
	imgWidth := int(maxWidth) + horizontalPadding*2
	lineHeight := fontsize * lineSpacingFactor
	imgHeight := int(float64(len(lines))*lineHeight) + verticalPadding*2 + windowBarHeight

	// Set up context with transparent background (NO CLIPPING)
	dc := gg.NewContext(imgWidth, imgHeight)
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	// -- Set background color from theme --
	bgEntry := style.Get(chroma.Background)
	r, g, b := 40, 42, 54 // Dracula fallback
	if bgEntry.Background.IsSet() {
		r, g, b = int(bgEntry.Background.Red()), int(bgEntry.Background.Green()), int(bgEntry.Background.Blue())
	}
	dc.SetRGB255(r, g, b)

	// Draw window bar (same as background)
	dc.DrawRectangle(0, 0, float64(imgWidth), windowBarHeight)
	dc.Fill()

	// Draw code background (below bar, same color)
	dc.DrawRectangle(0, windowBarHeight, float64(imgWidth), float64(imgHeight-windowBarHeight))
	dc.Fill()

	// Draw traffic light buttons (red, yellow, green)
	trafficColors := []string{"#FF5F56", "#FFBD2E", "#27C93F"}
	for i, color := range trafficColors {
		dc.SetHexColor(color)
		x := float64(horizontalPadding) + float64(i)*(trafficLightR*2+trafficLightGap)
		y := windowBarHeight / 2.0
		dc.DrawCircle(x, y, trafficLightR)
		dc.Fill()
	}

	// Use loaded font face
	dc.SetFontFace(fontFace)

	// Render code lines, shifted down by windowBarHeight
	y := float64(verticalPadding) + fontsize + windowBarHeight
	for _, line := range lines {
		x := float64(horizontalPadding)
		for _, tok := range line {
			entry := style.Get(tok.Type)
			if entry.Colour.IsSet() {
				r, g, b := entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue()
				dc.SetRGB255(int(r), int(g), int(b))
			} else {
				dc.SetRGB(1, 1, 1)
			}
			dc.DrawString(tok.Value, x, y)
			tw, _ := dc.MeasureString(tok.Value)
			x += tw
		}
		y += lineHeight
	}

	// --- Post-process: rounded corner mask ---
	img := toNRGBA(dc.Image())
	mask := makeRoundedRectMask(imgWidth, imgHeight, cornerRadius)
	out := image.NewNRGBA(img.Bounds())
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			src := img.NRGBAAt(x, y)
			alpha := mask.AlphaAt(x, y).A
			out.SetNRGBA(x, y, color.NRGBA{
				R: src.R,
				G: src.G,
				B: src.B,
				A: uint8(uint16(src.A) * uint16(alpha) / 255),
			})
		}
	}

	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(&buf, out); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %v", err)
	}
	return buf.Bytes(), nil
}
