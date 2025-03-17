package main

import (
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"sync"
)

func newImage(r io.Reader) (image.Image, string) {
	img, format, err := image.Decode(r)
	if err != nil {
		panic(errors.New("Error decoding image\n" + err.Error()))
	}
	return img, format
}

func openFile(path string, callback func(f *os.File)) {
	file, err := os.Open(path)
	if err != nil {
		panic(errors.New("Error opening file\n" + err.Error()))
	}
	defer file.Close()
	callback(file)
}

func createFile(path string, callback func(f *os.File)) {
	file, err := os.Create(path)
	if err != nil {
		panic(errors.New("Error creating file\n" + err.Error()))
	}
	defer file.Close()
	callback(file)
}

func _readerCheck(img image.Image, subImg *image.RGBA) {
	rect := subImg.Rect
	left, top, width, height := rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y
	pix := subImg.Pix
	stride := subImg.Stride
	for y := top; y < height; y++ {
		t := (y - top) * stride
		for x := left; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			i := t + (x-left)*4
			s := pix[i : i+4 : i+4]
			s[0] = uint8(r * 0xffff / a >> 8)
			s[1] = uint8(g * 0xffff / a >> 8)
			s[2] = uint8(b * 0xffff / a >> 8)
			s[3] = uint8(a >> 8)
		}
	}
}
func unPremultipliedImage(path string) {
	openFile(path, func(file *os.File) {
		img, _ := newImage(file)
		result := image.NewRGBA(img.Bounds())
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		numGoroutines := 16
		step := height / numGoroutines
		var wg sync.WaitGroup
		for i := range numGoroutines {
			startY := i * step
			endY := (i + 1) * step
			if i == numGoroutines-1 {
				endY = height
			}
			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				_readerCheck(img, result.SubImage(image.Rect(0, start, width, end)).(*image.RGBA))
			}(startY, endY)
		}
		wg.Wait()
		createFile(path, func(file *os.File) {
			png.Encode(file, result)
		})
	})
}

func main() {
	if len(os.Args) != 2 {
		println("Usage: program <filepath>")
		return
	}
	unPremultipliedImage(os.Args[1])
}
