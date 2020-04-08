package main

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"path/filepath"

	"github.com/mewkiz/pkg/imgutil"
	"github.com/pkg/errors"
)

func main() {
	if err := ren(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func ren() error {
	if err := load(); err != nil {
		return errors.WithStack(err)
	}
	if err := renderYenwood(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func renderYenwood() error {
	// BKG -> background
	// BKGSM -> background small
	// NM -> normal
	// HGT -> hight (z axis?)
	// AS -> unknown
	//
	// Ryyy -> row yyy
	// Cxxx -> column xxx
	yenwood := area{
		name:  "yenwood",
		nrows: 3, // number of rows
		ncols: 4, // number of columns
	}
	thumb := yenwood.thumb()
	thumbPath := fmt.Sprintf("%s_thumb.png", yenwood.name)
	if err := imgutil.WriteFile(thumbPath, thumb); err != nil {
		return errors.WithStack(err)
	}
	background := yenwood.background()
	backgroundPath := fmt.Sprintf("%s_background.png", yenwood.name)
	if err := imgutil.WriteFile(backgroundPath, background); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// row=0, col=0 starts at bottom left.
type area struct {
	name            string
	nrows           int
	ncols           int
	backgroundLayer draw.Image
	normalLayer     draw.Image
	hightLayer      draw.Image
	asLayer         draw.Image
}

func (a *area) getWidth() int {
	width := -1
	for row := 0; row < a.nrows; row++ {
		w := 0
		for col := 0; col < a.ncols; col++ {
			img := a.subimg(kindBackground, row, col)
			bounds := img.Bounds()
			w += bounds.Dx()
		}
		if width == -1 {
			width = w
		} else if w != width {
			panic(fmt.Errorf("mismatch between width of %q (prev=%d, new=%d)", a.name, width, w))
		}
	}
	return width
}

func (a *area) getHeight() int {
	height := -1
	for col := 0; col < a.ncols; col++ {
		h := 0
		for row := 0; row < a.nrows; row++ {
			img := a.subimg(kindBackground, row, col)
			bounds := img.Bounds()
			h += bounds.Dy()
		}
		if height == -1 {
			height = h
		} else if h != height {
			panic(fmt.Errorf("mismatch between height of %q (prev=%d, new=%d)", a.name, height, h))
		}
	}
	return height
}

func (a *area) background() image.Image {
	width, height := a.getWidth(), a.getHeight()
	bounds := image.Rect(0, 0, width, height)
	background := image.NewRGBA(bounds)
	y := 0
	for row := a.nrows - 1; row >= 0; row-- {
		x := 0
		var srcHeight int
		for col := 0; col < a.ncols; col++ {
			src := a.subimg(kindBackground, row, col)
			srcWidth := src.Bounds().Dx()
			srcHeight = src.Bounds().Dy()
			dr := image.Rect(x, y, x+srcWidth, y+srcHeight)
			sp := image.Pt(0, 0)
			draw.Draw(background, dr, src, sp, draw.Src)
			x += srcWidth
		}
		y += srcHeight
	}
	return background
}

func (a *area) thumb() image.Image {
	// BKGSM -> background small
	const kind = kindBackgroundSmall
	// "yenwood_BKGSM.png"
	imgName := fmt.Sprintf("%s_%s.png", a.name, kind)
	img, ok := imgs[imgName]
	if !ok {
		panic(fmt.Errorf("unable to locate %q (kind %q) of %q", imgName, kind, a.name))
	}
	return img
}

func (a *area) subimg(kind string, row, col int) image.Image {
	//"yenwood_BKG_R002_C002.png"
	imgName := fmt.Sprintf("%s_%s_R%03d_C%03d.png", a.name, kind, row, col)
	img, ok := imgs[imgName]
	if !ok {
		panic(fmt.Errorf("unable to locate %q (kind %q) of %q", imgName, kind, a.name))
	}
	return img
}

// load loads the game assets and populates the imgs map.
func load() error {
	imgNames := []string{
		Yenwood_AS_R000_C000,
		Yenwood_AS_R000_C001,
		Yenwood_AS_R000_C002,
		Yenwood_AS_R000_C003,
		Yenwood_AS_R001_C000,
		Yenwood_AS_R001_C001,
		Yenwood_AS_R001_C002,
		Yenwood_AS_R001_C003,
		Yenwood_AS_R002_C000,
		Yenwood_AS_R002_C001,
		Yenwood_AS_R002_C002,
		Yenwood_AS_R002_C003,
		Yenwood_BKG_R000_C000,
		Yenwood_BKG_R000_C001,
		Yenwood_BKG_R000_C002,
		Yenwood_BKG_R000_C003,
		Yenwood_BKG_R001_C000,
		Yenwood_BKG_R001_C001,
		Yenwood_BKG_R001_C002,
		Yenwood_BKG_R001_C003,
		Yenwood_BKG_R002_C000,
		Yenwood_BKG_R002_C001,
		Yenwood_BKG_R002_C002,
		Yenwood_BKG_R002_C003,
		Yenwood_BKGSM,
		Yenwood_HGT_R000_C000,
		Yenwood_HGT_R000_C001,
		Yenwood_HGT_R000_C002,
		Yenwood_HGT_R000_C003,
		Yenwood_HGT_R001_C000,
		Yenwood_HGT_R001_C001,
		Yenwood_HGT_R001_C002,
		Yenwood_HGT_R001_C003,
		Yenwood_HGT_R002_C000,
		Yenwood_HGT_R002_C001,
		Yenwood_HGT_R002_C002,
		Yenwood_HGT_R002_C003,
		Yenwood_NM_R000_C000,
		Yenwood_NM_R000_C001,
		Yenwood_NM_R000_C002,
		Yenwood_NM_R000_C003,
		Yenwood_NM_R001_C000,
		Yenwood_NM_R001_C001,
		Yenwood_NM_R001_C002,
		Yenwood_NM_R001_C003,
		Yenwood_NM_R002_C000,
		Yenwood_NM_R002_C001,
		Yenwood_NM_R002_C002,
		Yenwood_NM_R002_C003,
	}
	for _, imgName := range imgNames {
		imgPath := filepath.Join("_assets_", "yenwood", imgName)
		img, err := imgutil.ReadFile(imgPath)
		if err != nil {
			return errors.WithStack(err)
		}
		imgs[imgName] = img
	}
	return nil
}

// imgs maps from image name to image content.
var imgs = make(map[string]image.Image)

const (
	// BKG -> background
	kindBackground = "BKG"
	// BKGSM -> background small
	kindBackgroundSmall = "BKGSM"
	// NM -> normal
	kindNormal = "NM"
	// HGT -> hight (z axis?)
	kindHeight = "HGT"
	// AS -> unknown
	kindAS = "AS" // TODO: rename once we know what this does.
)

// Ryyy -> row yyy
// Cxxx -> column xxx
const (
	Yenwood_AS_R000_C000  = "yenwood_AS_R000_C000.png"
	Yenwood_AS_R000_C001  = "yenwood_AS_R000_C001.png"
	Yenwood_AS_R000_C002  = "yenwood_AS_R000_C002.png"
	Yenwood_AS_R000_C003  = "yenwood_AS_R000_C003.png"
	Yenwood_AS_R001_C000  = "yenwood_AS_R001_C000.png"
	Yenwood_AS_R001_C001  = "yenwood_AS_R001_C001.png"
	Yenwood_AS_R001_C002  = "yenwood_AS_R001_C002.png"
	Yenwood_AS_R001_C003  = "yenwood_AS_R001_C003.png"
	Yenwood_AS_R002_C000  = "yenwood_AS_R002_C000.png"
	Yenwood_AS_R002_C001  = "yenwood_AS_R002_C001.png"
	Yenwood_AS_R002_C002  = "yenwood_AS_R002_C002.png"
	Yenwood_AS_R002_C003  = "yenwood_AS_R002_C003.png"
	Yenwood_BKG_R000_C000 = "yenwood_BKG_R000_C000.png"
	Yenwood_BKG_R000_C001 = "yenwood_BKG_R000_C001.png"
	Yenwood_BKG_R000_C002 = "yenwood_BKG_R000_C002.png"
	Yenwood_BKG_R000_C003 = "yenwood_BKG_R000_C003.png"
	Yenwood_BKG_R001_C000 = "yenwood_BKG_R001_C000.png"
	Yenwood_BKG_R001_C001 = "yenwood_BKG_R001_C001.png"
	Yenwood_BKG_R001_C002 = "yenwood_BKG_R001_C002.png"
	Yenwood_BKG_R001_C003 = "yenwood_BKG_R001_C003.png"
	Yenwood_BKG_R002_C000 = "yenwood_BKG_R002_C000.png"
	Yenwood_BKG_R002_C001 = "yenwood_BKG_R002_C001.png"
	Yenwood_BKG_R002_C002 = "yenwood_BKG_R002_C002.png"
	Yenwood_BKG_R002_C003 = "yenwood_BKG_R002_C003.png"
	Yenwood_BKGSM         = "yenwood_BKGSM.png"
	Yenwood_HGT_R000_C000 = "yenwood_HGT_R000_C000.png"
	Yenwood_HGT_R000_C001 = "yenwood_HGT_R000_C001.png"
	Yenwood_HGT_R000_C002 = "yenwood_HGT_R000_C002.png"
	Yenwood_HGT_R000_C003 = "yenwood_HGT_R000_C003.png"
	Yenwood_HGT_R001_C000 = "yenwood_HGT_R001_C000.png"
	Yenwood_HGT_R001_C001 = "yenwood_HGT_R001_C001.png"
	Yenwood_HGT_R001_C002 = "yenwood_HGT_R001_C002.png"
	Yenwood_HGT_R001_C003 = "yenwood_HGT_R001_C003.png"
	Yenwood_HGT_R002_C000 = "yenwood_HGT_R002_C000.png"
	Yenwood_HGT_R002_C001 = "yenwood_HGT_R002_C001.png"
	Yenwood_HGT_R002_C002 = "yenwood_HGT_R002_C002.png"
	Yenwood_HGT_R002_C003 = "yenwood_HGT_R002_C003.png"
	Yenwood_NM_R000_C000  = "yenwood_NM_R000_C000.png"
	Yenwood_NM_R000_C001  = "yenwood_NM_R000_C001.png"
	Yenwood_NM_R000_C002  = "yenwood_NM_R000_C002.png"
	Yenwood_NM_R000_C003  = "yenwood_NM_R000_C003.png"
	Yenwood_NM_R001_C000  = "yenwood_NM_R001_C000.png"
	Yenwood_NM_R001_C001  = "yenwood_NM_R001_C001.png"
	Yenwood_NM_R001_C002  = "yenwood_NM_R001_C002.png"
	Yenwood_NM_R001_C003  = "yenwood_NM_R001_C003.png"
	Yenwood_NM_R002_C000  = "yenwood_NM_R002_C000.png"
	Yenwood_NM_R002_C001  = "yenwood_NM_R002_C001.png"
	Yenwood_NM_R002_C002  = "yenwood_NM_R002_C002.png"
	Yenwood_NM_R002_C003  = "yenwood_NM_R002_C003.png"
)
