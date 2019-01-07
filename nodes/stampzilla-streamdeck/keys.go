package main

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"./streamdeck"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

func drawTempToKey(deck *streamdeck.StreamDeck, label string, value float32, key int) {
	dest := image.NewRGBA(image.Rect(0, 0, ICON_SIZE, ICON_SIZE))
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(color.RGBA{0xff, 0x66, 0x00, 0xff})
	gc.SetStrokeColor(color.RGBA{0xff, 0xff, 0xff, 0xff})

	gc.SetFontSize(28)
	gc.SetFontData(draw2d.FontData{
		Name: "Roboto",
	})

	text := fmt.Sprintf("%.0f", value)
	left, top, right, bottom := gc.GetStringBounds(text)
	gc.FillStringAt(text, 72/2-((right-left)/2), 72/2)

	// Label
	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.SetFontSize(14)
	left, top, right, bottom = gc.GetStringBounds(label)
	gc.FillStringAt(label, 72/2-((right-left)/2), 72-((bottom-top)/2))

	deck.WriteImageToKey(dest, key)
}

func drawStateToKey(deck *streamdeck.StreamDeck, label string, value interface{}, key int) {
	dest := image.NewRGBA(image.Rect(0, 0, ICON_SIZE, ICON_SIZE))
	gc := draw2dimg.NewGraphicContext(dest)

	text := ""
	switch state := value.(type) {
	case bool:
		text = strconv.FormatBool(state)
	case int:
		text = "int"
	default:
		text = "unkn"
	}

	// State
	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.SetStrokeColor(color.RGBA{0xff, 0xff, 0xff, 0xff})

	gc.SetFontSize(28)
	gc.SetFontData(draw2d.FontData{
		Name: "Roboto",
	})

	left, top, right, bottom := gc.GetStringBounds(text)
	gc.FillStringAt(text, 72/2-((right-left)/2), 72/2)

	// Label
	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.SetFontSize(14)
	left, top, right, bottom = gc.GetStringBounds(label)
	gc.FillStringAt(label, 72/2-((right-left)/2), 72-((bottom-top)/2))

	deck.WriteImageToKey(dest, key)
}
