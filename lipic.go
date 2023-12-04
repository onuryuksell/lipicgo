package lipicgo

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
)

const TOP_LEFT = 0
const TOP_CENTER = 1
const TOP_RIGHT = 2
const CENTER_LEFT = 3
const CENTER = 4
const CENTER_RIGHT = 5
const BOTTOM_LEFT = 6
const BOTTOM_CENTER = 7
const BOTTOM_RIGHT = 8

type Image struct {
	Width  int
	Height int
	Type   string
	Pixels [][]color.RGBA
}

func Load(path string) Image {
	image, imageType, _ := read(path)

	return Image{
		Width:  image.Bounds().Dx(),
		Height: image.Bounds().Dy(),
		Type:   imageType,
		Pixels: createPixelsFromImage(image),
	}
}

func (image Image) At(x int, y int) color.RGBA {
	return image.Pixels[y][x]
}

func (image *Image) Set(x int, y int, c color.RGBA) {
	image.Pixels[y][x] = c
}

func blendPixelSourceOver(background color.RGBA, foreground color.RGBA) color.RGBA {
	bgR, bgG, bgB, uibgA := background.R, background.G, background.B, background.A
	fgR, fgG, fgB, uifgA := foreground.R, foreground.G, foreground.B, foreground.A

	// Keep alpha between 0-1 for multiplication
	fgA := float64(float64(uifgA) / 255)
	bgA := float64(float64(uibgA) / 255)

	alphaFinal := bgA + fgA - bgA*fgA
	bgRa := float64(bgR) * bgA
	bgGa := float64(bgG) * bgA
	bgBa := float64(bgB) * bgA

	fgRa := float64(fgR) * fgA
	fgGa := float64(fgG) * fgA
	fgBa := float64(fgB) * fgA

	finalColorRa := uint8(float64(fgRa) + bgRa*(float64(1)-fgA))
	finalColorGa := uint8(float64(fgGa) + bgGa*(float64(1)-fgA))
	finalColorBa := uint8(float64(fgBa) + bgBa*(float64(1)-fgA))

	finalColorR := uint8(float64(finalColorRa) / alphaFinal)
	finalColorG := uint8(float64(finalColorGa) / alphaFinal)
	finalColorB := uint8(float64(finalColorBa) / alphaFinal)

	// Make alpha between again 0-255
	alphaFinal *= 255

	return color.RGBA{uint8(finalColorR), uint8(finalColorG), uint8(finalColorB), uint8(alphaFinal)}
}

func (img *Image) ResizeByScale(scale float64) {
	h, w := float64(img.Height)*scale, float64(img.Width)*scale
	img.Resize(int(w), int(h))
}

func (img *Image) Resize(width int, height int) {
	img.Pixels = img.bilinearInterpolation(width, height)
	img.Width = len(img.Pixels[0])
	img.Height = len(img.Pixels)
}

func (img *Image) Opacity(opacity float64) {
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			var c color.RGBA = img.At(x, y)

			if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
				continue
			}

			if opacity > 100 {
				c.A = uint8(opacity)
			} else if opacity > 0 && opacity < 1 {
				c.A = uint8(float64(255) * opacity)
			} else if opacity < 0 {
				c.A = 0
			}

			img.Set(x, y, c)
		}
	}
}

func (img *Image) getAlign(image Image, alignH string, alignV string) (int, int) {
	x, y := 0, 0
	switch alignV {
	case "TOP":
		y = 0
	case "CENTER":
		y = int(img.Height/2 - image.Height/2)
	case "BOTTOM":
		y = img.Height - image.Height
	}

	switch alignH {
	case "LEFT":
		x = 0
	case "CENTER":
		x = int(img.Width/2 - image.Width/2)
	case "RIGHT":
		x = img.Width - image.Width
	}
	return x, y
}

func (img *Image) Place(image Image, x int, y int, ignoreBounds ...bool) {
	bounds := true
	if len(ignoreBounds) > 0 {
		bounds = false
	}
	if bounds && x+image.Width > img.Width {
		x = img.Width - image.Width
	}

	if bounds && y+image.Height > img.Height {
		y = img.Height - image.Height
	}

	for y0 := 0; y0 < image.Height; y0++ {
		for x0 := 0; x0 < image.Width; x0++ {
			if x+x0 < img.Width && y+y0 < img.Height {
				fmt.Println(x, x0)
				c := image.At(x0, y0)

				if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
					continue
				}

				c = blendPixelSourceOver(img.At(x+x0, y+y0), c)
				img.Set(x+x0, y+y0, c)
			}
		}
	}
}

func (img *Image) Watermark(image Image, alignOrPos ...int) {
	if len(alignOrPos) == 0 {
		fmt.Println("Align or PosX-Y must be given!")
		return
	}
}

func (img Image) bilinearInterpolation(width int, height int) [][]color.RGBA {
	wScaleFactor, hScaleFactor := float64(0), float64(0)
	if width != 0 {
		wScaleFactor = float64(img.Width) / float64(width)
	}

	if height != 0 {
		hScaleFactor = float64(img.Height) / float64(height)
	}

	resized := make([][]color.RGBA, height)
	println(len(resized))
	for y := 0; y < height; y++ {
		resized[y] = make([]color.RGBA, width)
		for x := 0; x < width; x++ {

			oldX := float64(float64(x) * wScaleFactor)
			oldY := float64(float64(y) * hScaleFactor)

			oldXFloored := math.Floor(oldX)
			oldXCeiled := min(float64(img.Width-1), math.Ceil(oldX))
			oldYFloored := math.Floor(oldY)
			oldYCeiled := min(float64(img.Height-1), math.Ceil(oldY))
			v1 := img.At(int(oldXFloored), int(oldYFloored))
			v2 := img.At(int(oldXCeiled), int(oldYFloored))
			v3 := img.At(int(oldXFloored), int(oldYCeiled))
			v4 := img.At(int(oldXCeiled), int(oldYCeiled))

			multiply := func(color uint8, val float64) uint8 {
				out := uint8(float64(color) * val)
				return out
			}

			var q color.RGBA
			if oldXCeiled == oldXFloored && oldYFloored == oldYCeiled {
				q = img.At(int(oldX), int(oldY))
			} else if oldXCeiled == oldXFloored {
				q1 := img.At(int(oldX), int(oldYFloored))
				q2 := img.At(int(oldX), int(oldYCeiled))
				q = sumColors(applyCalculationToColor(q1, oldYCeiled-float64(oldY), multiply), applyCalculationToColor(q2, float64(oldY)-oldYFloored, multiply))
			} else if oldYCeiled == oldYFloored {
				q1 := img.At(int(oldXFloored), int(oldY))
				q2 := img.At(int(oldXCeiled), int(oldY))
				q = sumColors(applyCalculationToColor(q1, oldXCeiled-float64(oldX), multiply), applyCalculationToColor(q2, float64(oldX)-oldXFloored, multiply))
			} else {
				q1 := sumColors(applyCalculationToColor(v1, (oldXCeiled-float64(oldX)), multiply), applyCalculationToColor(v2, (float64(oldX)-oldXFloored), multiply))
				q2 := sumColors(applyCalculationToColor(v3, (oldXCeiled-float64(oldX)), multiply), applyCalculationToColor(v4, (float64(oldX)-oldXFloored), multiply))
				q = sumColors(applyCalculationToColor(q1, (oldYCeiled-float64(oldY)), multiply), applyCalculationToColor(q2, (float64(oldY)-oldYFloored), multiply))
			}
			resized[y][x] = q
		}
	}
	return resized
}

func applyCalculationToColor(pixel color.RGBA, val float64, calc func(color uint8, val float64) uint8) color.RGBA {
	r := calc(pixel.R, val)
	g := calc(pixel.G, val)
	b := calc(pixel.B, val)
	a := calc(pixel.A, val)

	return color.RGBA{r, g, b, a}
}

func sumColors(color1 color.RGBA, color2 color.RGBA) color.RGBA {
	newR := color1.R + color2.R
	newG := color1.G + color2.G
	newB := color1.B + color2.B
	newA := color1.A + color2.A

	return color.RGBA{newR, newG, newB, newA}
}

func (img Image) Save(path string) {
	image := img.toImage()
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = png.Encode(file, image)
	if err != nil {
		panic(err)
	}
}

func (img Image) toImage() image.Image {
	image := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			image.Set(x, y, img.At(x, y))
		}
	}
	return image
}

func createPixelsFromImage(img image.Image) [][]color.RGBA {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	pixels := make([][]color.RGBA, height)
	for y := 0; y < height; y++ {
		pixels[y] = make([]color.RGBA, width)
		for x := 0; x < width; x++ {
			realColor := img.At(x, y)
			var rgbaColor color.RGBA
			if rgba, ok := realColor.(color.RGBA); ok {
				rgbaColor = rgba
			} else {
				rgbaColor = color.RGBAModel.Convert(realColor).(color.RGBA)
			}

			pixels[y][x] = rgbaColor
		}
	}

	return pixels
}

func read(path string) (image.Image, string, error) {
	file, err := GetFromPath(path)
	if err != nil {
		fmt.Println("An error ocurred while reading file")
	}
	defer file.Close()
	imgBuff := bufio.NewReader(file)
	image, imageType, err := image.Decode(imgBuff)
	return image, imageType, err
}
