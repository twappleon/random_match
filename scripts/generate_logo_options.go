package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

type colorF struct {
	r, g, b, a float64
}

type pt struct {
	x, y float64
}

func main() {
	outDir := "assets/brand/logo-options"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	for i := 1; i <= 5; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
		switch i {
		case 1:
			optionCameraPeople(img)
		case 2:
			optionShuffleFaces(img)
		case 3:
			optionVideoBubble(img)
		case 4:
			optionSplitMatch(img)
		case 5:
			optionVoiceVideo(img)
		}
		savePNG(filepath.Join(outDir, filename(i)), img)
	}
	savePNG(filepath.Join(outDir, "contact-sheet.png"), contactSheet(outDir))
}

func filename(i int) string {
	return "option-" + string(rune('0'+i)) + ".png"
}

func optionCameraPeople(img *image.RGBA) {
	bgRadial(img, colorF{38, 70, 255, 255}, colorF{11, 20, 40, 255})
	fillRoundRect(img, 110, 110, 914, 914, 210, colorF{0, 0, 0, 38})
	fillRoundRect(img, 172, 290, 650, 704, 132, colorF{255, 255, 255, 255})
	fillPoly(img, []pt{{620, 405}, {808, 302}, {858, 336}, {858, 660}, {808, 698}, {620, 595}}, colorF{255, 255, 255, 255})
	fillCircle(img, 390, 500, 132, colorF{21, 31, 51, 255})
	fillCircle(img, 342, 462, 52, colorF{93, 243, 235, 255})
	fillCircle(img, 454, 462, 52, colorF{255, 92, 159, 255})
	fillRoundRect(img, 284, 535, 500, 617, 42, colorF{255, 255, 255, 238})
	strokeCircle(img, 390, 500, 172, 24, colorF{93, 243, 235, 210})
}

func optionShuffleFaces(img *image.RGBA) {
	bgLinear(img, colorF{9, 18, 31, 255}, colorF{36, 31, 88, 255})
	fillCircle(img, 326, 470, 152, colorF{93, 243, 235, 255})
	fillCircle(img, 698, 470, 152, colorF{255, 92, 159, 255})
	fillCircle(img, 278, 438, 30, colorF{9, 18, 31, 235})
	fillCircle(img, 374, 438, 30, colorF{9, 18, 31, 235})
	drawArc(img, 326, 492, 62, .38, 2.76, 22, colorF{9, 18, 31, 235})
	fillCircle(img, 650, 438, 30, colorF{9, 18, 31, 235})
	fillCircle(img, 746, 438, 30, colorF{9, 18, 31, 235})
	drawArc(img, 698, 492, 62, .38, 2.76, 22, colorF{9, 18, 31, 235})
	drawLine(img, pt{420, 356}, pt{604, 356}, 34, colorF{255, 255, 255, 230})
	drawLine(img, pt{604, 356}, pt{558, 310}, 34, colorF{255, 255, 255, 230})
	drawLine(img, pt{604, 356}, pt{558, 402}, 34, colorF{255, 255, 255, 230})
	drawLine(img, pt{604, 640}, pt{420, 640}, 34, colorF{255, 255, 255, 230})
	drawLine(img, pt{420, 640}, pt{466, 594}, 34, colorF{255, 255, 255, 230})
	drawLine(img, pt{420, 640}, pt{466, 686}, 34, colorF{255, 255, 255, 230})
	fillRoundRect(img, 386, 676, 638, 810, 52, colorF{255, 255, 255, 255})
	fillPoly(img, []pt{{618, 710}, {742, 646}, {742, 840}, {618, 776}}, colorF{255, 255, 255, 255})
}

func optionVideoBubble(img *image.RGBA) {
	bgRadial(img, colorF{255, 92, 159, 255}, colorF{18, 22, 48, 255})
	fillRoundRect(img, 168, 226, 760, 670, 122, colorF{255, 255, 255, 255})
	fillPoly(img, []pt{{298, 660}, {244, 812}, {440, 676}}, colorF{255, 255, 255, 255})
	fillRoundRect(img, 292, 344, 560, 552, 72, colorF{18, 22, 48, 255})
	fillPoly(img, []pt{{540, 396}, {710, 310}, {758, 340}, {758, 556}, {710, 588}, {540, 500}}, colorF{18, 22, 48, 255})
	fillCircle(img, 425, 448, 58, colorF{93, 243, 235, 255})
	fillCircle(img, 425, 448, 28, colorF{18, 22, 48, 255})
	fillCircle(img, 650, 300, 36, colorF{255, 209, 102, 255})
	fillHeart(img, 622, 702, 82, colorF{255, 92, 159, 255})
}

func optionSplitMatch(img *image.RGBA) {
	bgLinear(img, colorF{9, 132, 148, 255}, colorF{18, 22, 48, 255})
	fillRoundRect(img, 148, 204, 876, 820, 92, colorF{255, 255, 255, 255})
	fillRoundRect(img, 206, 270, 478, 648, 56, colorF{25, 38, 66, 255})
	fillRoundRect(img, 546, 270, 818, 648, 56, colorF{25, 38, 66, 255})
	fillCircle(img, 342, 408, 66, colorF{93, 243, 235, 255})
	fillRoundRect(img, 270, 514, 414, 596, 40, colorF{93, 243, 235, 255})
	fillCircle(img, 682, 408, 66, colorF{255, 92, 159, 255})
	fillRoundRect(img, 610, 514, 754, 596, 40, colorF{255, 92, 159, 255})
	drawLine(img, pt{474, 458}, pt{548, 458}, 28, colorF{255, 209, 102, 255})
	drawLine(img, pt{548, 458}, pt{518, 428}, 28, colorF{255, 209, 102, 255})
	drawLine(img, pt{548, 458}, pt{518, 488}, 28, colorF{255, 209, 102, 255})
	fillRoundRect(img, 332, 696, 618, 766, 35, colorF{18, 22, 48, 255})
	fillPoly(img, []pt{{602, 712}, {690, 668}, {690, 794}, {602, 750}}, colorF{18, 22, 48, 255})
}

func optionVoiceVideo(img *image.RGBA) {
	bgRadial(img, colorF{93, 243, 235, 255}, colorF{13, 24, 44, 255})
	fillCircle(img, 512, 512, 326, colorF{255, 255, 255, 255})
	fillRoundRect(img, 294, 318, 546, 538, 68, colorF{18, 22, 48, 255})
	fillPoly(img, []pt{{528, 376}, {690, 292}, {738, 324}, {738, 532}, {690, 566}, {528, 480}}, colorF{18, 22, 48, 255})
	fillCircle(img, 420, 428, 42, colorF{93, 243, 235, 255})
	fillCircle(img, 420, 428, 18, colorF{18, 22, 48, 255})
	drawArc(img, 512, 560, 128, .42, 2.72, 32, colorF{255, 92, 159, 255})
	drawArc(img, 512, 560, 190, .36, 2.78, 26, colorF{255, 209, 102, 230})
	fillCircle(img, 292, 744, 50, colorF{93, 243, 235, 255})
	fillCircle(img, 732, 744, 50, colorF{255, 92, 159, 255})
}

func contactSheet(dir string) *image.RGBA {
	const cell = 256
	const gap = 24
	w := cell*5 + gap*6
	h := cell + gap*2
	sheet := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(sheet, sheet.Bounds(), &image.Uniform{color.RGBA{246, 248, 252, 255}}, image.Point{}, draw.Src)
	for i := 1; i <= 5; i++ {
		file, err := os.Open(filepath.Join(dir, filename(i)))
		if err != nil {
			panic(err)
		}
		src, err := png.Decode(file)
		_ = file.Close()
		if err != nil {
			panic(err)
		}
		dst := image.Rect(gap+(i-1)*(cell+gap), gap, gap+(i-1)*(cell+gap)+cell, gap+cell)
		scaleNearest(sheet, dst, src)
	}
	return sheet
}

func scaleNearest(dst *image.RGBA, rect image.Rectangle, src image.Image) {
	sb := src.Bounds()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		sy := sb.Min.Y + (y-rect.Min.Y)*sb.Dy()/rect.Dy()
		for x := rect.Min.X; x < rect.Max.X; x++ {
			sx := sb.Min.X + (x-rect.Min.X)*sb.Dx()/rect.Dx()
			dst.Set(x, y, src.At(sx, sy))
		}
	}
}

func savePNG(path string, img image.Image) {
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		panic(err)
	}
}

func bgLinear(img *image.RGBA, a, b colorF) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			t := (float64(x) + float64(y)) / float64(bounds.Dx()+bounds.Dy())
			img.SetRGBA(x, y, toRGBA(lerp(a, b, t)))
		}
	}
}

func bgRadial(img *image.RGBA, a, b colorF) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			nx := float64(x)/float64(bounds.Dx()) - .3
			ny := float64(y)/float64(bounds.Dy()) - .2
			t := clamp(math.Sqrt(nx*nx+ny*ny)*1.35, 0, 1)
			img.SetRGBA(x, y, toRGBA(lerp(a, b, t)))
		}
	}
}

func fillRoundRect(img *image.RGBA, x0, y0, x1, y1, r float64, c colorF) {
	for y := int(y0); y <= int(y1); y++ {
		for x := int(x0); x <= int(x1); x++ {
			fx, fy := float64(x), float64(y)
			cx := clamp(fx, x0+r, x1-r)
			cy := clamp(fy, y0+r, y1-r)
			if dist(fx, fy, cx, cy) <= r {
				blend(img, x, y, c)
			}
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, r float64, c colorF) {
	rr := r * r
	for y := int(cy - r); y <= int(cy+r); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			if sq(float64(x)-cx)+sq(float64(y)-cy) <= rr {
				blend(img, x, y, c)
			}
		}
	}
}

func strokeCircle(img *image.RGBA, cx, cy, r, width float64, c colorF) {
	outer := r * r
	inner := (r - width) * (r - width)
	for y := int(cy - r); y <= int(cy+r); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			d := sq(float64(x)-cx) + sq(float64(y)-cy)
			if d <= outer && d >= inner {
				blend(img, x, y, c)
			}
		}
	}
}

func fillPoly(img *image.RGBA, pts []pt, c colorF) {
	minX, minY, maxX, maxY := pts[0].x, pts[0].y, pts[0].x, pts[0].y
	for _, p := range pts {
		minX, minY = math.Min(minX, p.x), math.Min(minY, p.y)
		maxX, maxY = math.Max(maxX, p.x), math.Max(maxY, p.y)
	}
	for y := int(minY); y <= int(maxY); y++ {
		for x := int(minX); x <= int(maxX); x++ {
			if inPoly(float64(x), float64(y), pts) {
				blend(img, x, y, c)
			}
		}
	}
}

func fillHeart(img *image.RGBA, cx, cy, r float64, c colorF) {
	fillCircle(img, cx-r*.32, cy-r*.18, r*.34, c)
	fillCircle(img, cx+r*.32, cy-r*.18, r*.34, c)
	fillPoly(img, []pt{{cx - r*.7, cy - r*.1}, {cx + r*.7, cy - r*.1}, {cx, cy + r*.78}}, c)
}

func drawLine(img *image.RGBA, a, b pt, width float64, c colorF) {
	minX := int(math.Min(a.x, b.x) - width)
	maxX := int(math.Max(a.x, b.x) + width)
	minY := int(math.Min(a.y, b.y) - width)
	maxY := int(math.Max(a.y, b.y) + width)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if lineDist(pt{float64(x), float64(y)}, a, b) <= width/2 {
				blend(img, x, y, c)
			}
		}
	}
}

func drawArc(img *image.RGBA, cx, cy, r, start, end, width float64, c colorF) {
	steps := 120
	prev := pt{cx + math.Cos(start)*r, cy + math.Sin(start)*r}
	for i := 1; i <= steps; i++ {
		t := start + (end-start)*float64(i)/float64(steps)
		next := pt{cx + math.Cos(t)*r, cy + math.Sin(t)*r}
		drawLine(img, prev, next, width, c)
		prev = next
	}
}

func blend(img *image.RGBA, x, y int, c colorF) {
	if !image.Pt(x, y).In(img.Bounds()) {
		return
	}
	dst := img.RGBAAt(x, y)
	a := c.a / 255
	ia := 1 - a
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(c.r*a + float64(dst.R)*ia),
		G: uint8(c.g*a + float64(dst.G)*ia),
		B: uint8(c.b*a + float64(dst.B)*ia),
		A: uint8(c.a + float64(dst.A)*ia),
	})
}

func toRGBA(c colorF) color.RGBA {
	return color.RGBA{uint8(c.r), uint8(c.g), uint8(c.b), uint8(c.a)}
}

func lerp(a, b colorF, t float64) colorF {
	t = clamp(t, 0, 1)
	return colorF{
		r: a.r + (b.r-a.r)*t,
		g: a.g + (b.g-a.g)*t,
		b: a.b + (b.b-a.b)*t,
		a: a.a + (b.a-a.a)*t,
	}
}

func inPoly(x, y float64, pts []pt) bool {
	inside := false
	j := len(pts) - 1
	for i := range pts {
		if (pts[i].y > y) != (pts[j].y > y) && x < (pts[j].x-pts[i].x)*(y-pts[i].y)/(pts[j].y-pts[i].y)+pts[i].x {
			inside = !inside
		}
		j = i
	}
	return inside
}

func lineDist(p, a, b pt) float64 {
	l2 := sq(b.x-a.x) + sq(b.y-a.y)
	if l2 == 0 {
		return dist(p.x, p.y, a.x, a.y)
	}
	t := clamp(((p.x-a.x)*(b.x-a.x)+(p.y-a.y)*(b.y-a.y))/l2, 0, 1)
	return dist(p.x, p.y, a.x+t*(b.x-a.x), a.y+t*(b.y-a.y))
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}

func dist(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(sq(x1-x2) + sq(y1-y2))
}

func sq(v float64) float64 {
	return v * v
}
