// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package ai

import (
	"fmt"
	"github.com/MatrixAINetwork/go-matrix/common/mt19937"
	"github.com/MatrixAINetwork/go-matrix/crypto/sha3"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
)

type Picture struct {
	picturePaths []string
	images       []image.Image
	backGround   *image.RGBA
}

const (
	RowInterceptNum    = 4
	CloInterceptNum    = 4
	Height             = 1024
	Width              = 1024
	PictureNum         = 16
	MinInterceptWidth  = Width / (RowInterceptNum * 2)
	MaxInterceptWidth  = Width / (RowInterceptNum)
	MinInterceptHeight = Height / (CloInterceptNum * 2)
	MaxInterceptHeight = Height / (CloInterceptNum)
)

var (
	errPictureNumber = errors.New("picture number err")
)

func New(pictures []string) (*Picture, error) {
	if len(pictures) != PictureNum {
		return nil, errPictureNumber
	}

	p := &Picture{
		picturePaths: make([]string, PictureNum),
	}

	// save picture paths
	copy(p.picturePaths, pictures)

	// load pictures
	for index, path := range p.picturePaths {
		if err := p.imgFileHandle(index, path); err != nil {
			return nil, err
		}
	}

	// create background
	p.backGround = createBackGround()

	return p, nil
}

func (p *Picture) imgFileHandle(index int, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Errorf("open No.%d picture(%s) failed: %v", index+1, path, err)
	}
	defer file.Close()

	// decode to img and save
	img, err := jpeg.Decode(file)
	if err != nil {
		return errors.Errorf("decode No.%d picture failed: %v", index+1, err)
	}
	p.images = append(p.images, img)
	return nil
}

func createBackGround() *image.RGBA {
	whiteBackground := image.NewRGBA(image.Rect(0, 0, 32, 32))
	whiteColor := color.RGBA{R: uint8(255), G: uint8(255), B: uint8(255), A: uint8(255)}
	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			whiteBackground.Set(i, j, &whiteColor)
		}
	}

	backGround := image.NewRGBA(image.Rect(0, 0, Width, Height))
	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if i%2 == 0 {
				if j%2 == 1 {
					startX := i * 32
					startY := j * 32
					draw.Draw(backGround, image.Rectangle{image.Point{startX, startY}, image.Point{startX + 31, startY + 31}}, whiteBackground, whiteBackground.Bounds().Min, draw.Over)
				}
			} else {
				if j%2 == 0 {
					startX := i * 32
					startY := j * 32
					draw.Draw(backGround, image.Rectangle{image.Point{startX, startY}, image.Point{startX + 31, startY + 31}}, whiteBackground, whiteBackground.Bounds().Min, draw.Over)
				}
			}
		}
	}
	return backGround
}

func (p *Picture) AIMine(randSeed int64) []byte {
	//copy target
	target := &image.RGBA{}
	target.Pix = make([]uint8, len(p.backGround.Pix))
	copy(target.Pix, p.backGround.Pix)
	target.Stride = p.backGround.Stride
	target.Rect = p.backGround.Rect

	//
	randHandle := mt19937.New()
	randHandle.Seed(randSeed)
	residueIndex := make([]int, PictureNum)
	for i := 0; i < PictureNum; i++ {
		residueIndex[i] = i
	}
	index := 0
	for len(residueIndex) != 0 {
		//挑选随机图片
		randPictureIndex := randHandle.Uint64() % uint64(len(residueIndex))
		picture := p.images[residueIndex[randPictureIndex]]
		//产生原始图像坐标，截取范围
		origin := GenFillOrigin(index, randHandle)
		originRect := GenFillRec(randHandle)
		//产生被截取图像起点
		interPoint := GenInterceptOrigin(picture.Bounds(), randHandle)
		originRect = Adjust(originRect, interPoint, picture.Bounds())
		drawRect := RecOffset(origin, originRect)
		//在背景图像上作图
		draw.Draw(target, drawRect, picture, interPoint, draw.Over)
		residueIndex = append(residueIndex[:randPictureIndex], residueIndex[randPictureIndex+1:]...)
		index++
	}

	//savePic(target)

	hash := sha3.Sum256(target.Pix)
	return hash[:]
}

func savePic(pic *image.RGBA) {
	file, err := os.Create("d:\\dst.jpg")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	jpeg.Encode(file, pic, nil)
}

func GenFillOrigin(index int, randhandel *mt19937.MT19937) image.Point {
	x := (index % RowInterceptNum) * (Height / RowInterceptNum)
	y := (index / CloInterceptNum) * (Width / CloInterceptNum)

	offsetX := randhandel.Uint64() % MaxInterceptHeight * 3 / 4
	offsetY := randhandel.Uint64() % MaxInterceptWidth * 3 / 4

	x += int(offsetX)
	y += int(offsetY)

	origin := image.Point{x, y}
	return origin
}

/*
func GenFillRec(origin image.Point, randhandel *mt19937.MT19937) image.Rectangle {
	x := int(randhandel.Uint64())%(MaxInterceptWidth-MinInterceptWidth) + MinInterceptWidth
	y := int(randhandel.Uint64())%(MaxInterceptHeight-MinInterceptHeight) + MinInterceptHeight

	min := image.Point{0 + origin.X, 0 + origin.Y}
	max := image.Point{x + origin.X, y + origin.Y}
	fmt.Println(max)
	return image.Rectangle{min, max}
}*/
func GenFillRec(randhandel *mt19937.MT19937) image.Rectangle {
	x := int(randhandel.Uint64())%(MaxInterceptWidth-MinInterceptWidth) + MinInterceptWidth
	y := int(randhandel.Uint64())%(MaxInterceptHeight-MinInterceptHeight) + MinInterceptHeight

	return image.Rectangle{image.Point{0, 0}, image.Point{x, y}}
}

func RecOffset(offset image.Point, rec image.Rectangle) image.Rectangle {
	min := image.Point{X: rec.Min.X + offset.X, Y: rec.Min.Y + offset.Y}
	max := image.Point{X: rec.Max.X + offset.X, Y: rec.Max.Y + offset.Y}
	return image.Rectangle{min, max}
}

func GenInterceptOrigin(rec image.Rectangle, randHandle *mt19937.MT19937) image.Point {
	maxWidth := uint64(rec.Max.X)
	maxHeight := uint64(rec.Max.Y)
	x := int(randHandle.Uint64() % maxWidth)
	y := int(randHandle.Uint64() % maxHeight)

	origin := image.Point{x, y}

	//fmt.Println(x,y,maxWidth,maxHeight)
	return origin
}

func Adjust(interSize image.Rectangle, inter image.Point, bound image.Rectangle) image.Rectangle {
	x := interSize.Max.X
	y := interSize.Max.Y

	if inter.X+x >= bound.Max.X {
		x = bound.Max.X - inter.X - 1
	}

	if inter.Y+y >= bound.Max.Y {
		y = bound.Max.Y - inter.Y - 1
	}

	return image.Rectangle{image.Point{0, 0}, image.Point{x, y}}
}
