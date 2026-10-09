package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type iconTarget struct {
	path     string
	size     int
	maskable bool
}

const sourceIcon = "assets/brand/logo-options/option-2.png"

func main() {
	srcFile, err := os.Open(sourceIcon)
	if err != nil {
		panic(err)
	}
	src, err := png.Decode(srcFile)
	if closeErr := srcFile.Close(); closeErr != nil {
		panic(closeErr)
	}
	if err != nil {
		panic(err)
	}

	targets := []iconTarget{
		{"apps/h5/public/favicon.png", 64, false},
		{"apps/h5/public/icons/icon-192.png", 192, false},
		{"apps/h5/public/icons/icon-512.png", 512, false},
		{"apps/h5/public/icons/maskable-192.png", 192, true},
		{"apps/h5/public/icons/maskable-512.png", 512, true},
		{"apps/mobile/web/favicon.png", 64, false},
		{"apps/mobile/web/icons/Icon-192.png", 192, false},
		{"apps/mobile/web/icons/Icon-512.png", 512, false},
		{"apps/mobile/web/icons/Icon-maskable-192.png", 192, true},
		{"apps/mobile/web/icons/Icon-maskable-512.png", 512, true},
		{"apps/mobile/android/app/src/main/res/mipmap-mdpi/ic_launcher.png", 48, false},
		{"apps/mobile/android/app/src/main/res/mipmap-hdpi/ic_launcher.png", 72, false},
		{"apps/mobile/android/app/src/main/res/mipmap-xhdpi/ic_launcher.png", 96, false},
		{"apps/mobile/android/app/src/main/res/mipmap-xxhdpi/ic_launcher.png", 144, false},
		{"apps/mobile/android/app/src/main/res/mipmap-xxxhdpi/ic_launcher.png", 192, false},
		{"apps/mobile/ios/Runner/Assets.xcassets/LaunchImage.imageset/LaunchImage.png", 168, false},
		{"apps/mobile/ios/Runner/Assets.xcassets/LaunchImage.imageset/LaunchImage@2x.png", 336, false},
		{"apps/mobile/ios/Runner/Assets.xcassets/LaunchImage.imageset/LaunchImage@3x.png", 504, false},
	}

	ios := map[string]int{
		"Icon-App-20x20@1x.png":     20,
		"Icon-App-20x20@2x.png":     40,
		"Icon-App-20x20@3x.png":     60,
		"Icon-App-29x29@1x.png":     29,
		"Icon-App-29x29@2x.png":     58,
		"Icon-App-29x29@3x.png":     87,
		"Icon-App-40x40@1x.png":     40,
		"Icon-App-40x40@2x.png":     80,
		"Icon-App-40x40@3x.png":     120,
		"Icon-App-60x60@2x.png":     120,
		"Icon-App-60x60@3x.png":     180,
		"Icon-App-76x76@1x.png":     76,
		"Icon-App-76x76@2x.png":     152,
		"Icon-App-83.5x83.5@2x.png": 167,
		"Icon-App-1024x1024@1x.png": 1024,
	}
	for name, size := range ios {
		targets = append(targets, iconTarget{
			path: filepath.Join("apps/mobile/ios/Runner/Assets.xcassets/AppIcon.appiconset", name),
			size: size,
		})
	}

	for _, target := range targets {
		if err := os.MkdirAll(filepath.Dir(target.path), 0o755); err != nil {
			panic(err)
		}
		img := resizeIcon(src, target.size, target.maskable)
		file, err := os.Create(target.path)
		if err != nil {
			panic(err)
		}
		if err := png.Encode(file, img); err != nil {
			_ = file.Close()
			panic(err)
		}
		if err := file.Close(); err != nil {
			panic(err)
		}
	}
}

func resizeIcon(src image.Image, size int, maskable bool) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	srcBounds := src.Bounds()
	inset := 0
	if maskable {
		inset = int(math.Round(float64(size) * 0.1))
		fill(dst, sample(src, srcBounds.Min.X, srcBounds.Min.Y))
	}
	drawSize := size - inset*2
	for y := 0; y < drawSize; y++ {
		for x := 0; x < drawSize; x++ {
			sx := float64(srcBounds.Min.X) + float64(x)*(float64(srcBounds.Dx()-1)/float64(drawSize-1))
			sy := float64(srcBounds.Min.Y) + float64(y)*(float64(srcBounds.Dy()-1)/float64(drawSize-1))
			dst.SetRGBA(x+inset, y+inset, sampleBilinear(src, sx, sy))
		}
	}
	return dst
}

func fill(img *image.RGBA, c color.RGBA) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func sample(img image.Image, x, y int) color.RGBA {
	r, g, b, a := img.At(x, y).RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func sampleBilinear(img image.Image, x, y float64) color.RGBA {
	b := img.Bounds()
	x0 := clampInt(int(math.Floor(x)), b.Min.X, b.Max.X-1)
	y0 := clampInt(int(math.Floor(y)), b.Min.Y, b.Max.Y-1)
	x1 := clampInt(x0+1, b.Min.X, b.Max.X-1)
	y1 := clampInt(y0+1, b.Min.Y, b.Max.Y-1)
	tx := x - float64(x0)
	ty := y - float64(y0)

	c00 := sample(img, x0, y0)
	c10 := sample(img, x1, y0)
	c01 := sample(img, x0, y1)
	c11 := sample(img, x1, y1)

	return color.RGBA{
		R: uint8(lerp(lerp(float64(c00.R), float64(c10.R), tx), lerp(float64(c01.R), float64(c11.R), tx), ty)),
		G: uint8(lerp(lerp(float64(c00.G), float64(c10.G), tx), lerp(float64(c01.G), float64(c11.G), tx), ty)),
		B: uint8(lerp(lerp(float64(c00.B), float64(c10.B), tx), lerp(float64(c01.B), float64(c11.B), tx), ty)),
		A: uint8(lerp(lerp(float64(c00.A), float64(c10.A), tx), lerp(float64(c01.A), float64(c11.A), tx), ty)),
	}
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func init() {
	if !strings.HasSuffix(sourceIcon, ".png") {
		panic("source icon must be a PNG")
	}
}
