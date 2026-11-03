package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

func createTerminalIcon(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background: dark charcoal rounded canvas
	bgColor := color.RGBA{R: 24, G: 24, B: 27, A: 255}
	borderColor := color.RGBA{R: 63, G: 63, B: 70, A: 255}
	accentColor := color.RGBA{R: 34, G: 197, B: 94, A: 255} // Emerald green
	cursorColor := color.RGBA{R: 96, G: 165, B: 250, A: 255} // Sky blue
	dotRed := color.RGBA{R: 239, G: 68, B: 68, A: 255}
	dotYellow := color.RGBA{R: 234, G: 179, B: 8, A: 255}
	dotGreen := color.RGBA{R: 34, G: 197, B: 94, A: 255}

	corner := size / 8
	borderW := size / 32
	if borderW < 1 {
		borderW = 1
	}

	// Draw rounded rectangle
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Check rounded corners
			dx := 0
			if x < corner {
				dx = corner - x
			} else if x >= size-corner {
				dx = x - (size - corner - 1)
			}
			dy := 0
			if y < corner {
				dy = corner - y
			} else if y >= size-corner {
				dy = y - (size - corner - 1)
			}

			if dx*dx+dy*dy > corner*corner {
				img.Set(x, y, color.Transparent)
				continue
			}

			if x < borderW || x >= size-borderW || y < borderW || y >= size-borderW {
				img.Set(x, y, borderColor)
			} else {
				img.Set(x, y, bgColor)
			}
		}
	}

	// Window top header dots
	dotRadius := size / 30
	if dotRadius < 2 {
		dotRadius = 2
	}
	dotY := size / 7
	dotGap := size / 14

	drawDot := func(cx, cy, r int, col color.RGBA) {
		for y := cy - r; y <= cy+r; y++ {
			for x := cx - r; x <= cx+r; x++ {
				if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r {
					img.Set(x, y, col)
				}
			}
		}
	}

	dotStartX := size / 6
	drawDot(dotStartX, dotY, dotRadius, dotRed)
	drawDot(dotStartX+dotGap, dotY, dotRadius, dotYellow)
	drawDot(dotStartX+dotGap*2, dotY, dotRadius, dotGreen)

	// Draw prompt '>' symbol
	promptSize := size / 3
	startX := size / 5
	startY := size * 38 / 100
	thick := size / 24
	if thick < 2 {
		thick = 2
	}

	for i := 0; i < promptSize/2; i++ {
		for t := 0; t < thick; t++ {
			img.Set(startX+i, startY+i+t, accentColor)
			img.Set(startX+i, startY+promptSize-i-1+t, accentColor)
		}
	}

	// Draw cursor '_' block
	cursorX := startX + promptSize*3/4
	cursorY := startY + promptSize - thick*2
	cursorW := promptSize / 2
	cursorH := thick * 2
	draw.Draw(img, image.Rect(cursorX, cursorY, cursorX+cursorW, cursorY+cursorH), &image.Uniform{C: cursorColor}, image.Point{}, draw.Src)

	return img
}

func main() {
	targets := []struct {
		path string
		size int
	}{
		{"package/ICON.PNG", 64},
		{"package/ICON_256.PNG", 256},
		{"package/app/ui/images/icon_64.png", 64},
		{"package/app/ui/images/icon_256.png", 256},
	}

	for _, t := range targets {
		if err := os.MkdirAll(filepath.Dir(t.path), 0755); err != nil {
			panic(err)
		}
		img := createTerminalIcon(t.size)
		f, err := os.Create(t.path)
		if err != nil {
			panic(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}
