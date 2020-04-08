package assets

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/mewkiz/pkg/imgutil"
	"github.com/pkg/errors"
)

//go:generate stringer -linecomment -type LayerKind

// LayerKind specifies a layer kind.
type LayerKind uint8

// Layer kinds.
const (
	// background layer
	LayerKindBackground LayerKind = iota + 1 // BKG
	// background (small) layer
	LayerKindBackgroundSmall // BKGSM
	// normal layer
	LayerKindNormal // NM
	// height layer (z axis)
	LayerKindHeight // HGT
	// as layer
	// TODO: figure out what AS is used for.
	LayerKindAS // AS
)

// Area is an area of the map.
type Area struct {
	// Area name.
	Name string
	// Background layer.
	BackgroundLayer image.Image
	// Normal layer.
	NormalLayer image.Image
	// Height layer.
	HeightLayer image.Image
	// AS layer.
	// TODO: figure out what AS is used for.
	ASLayer image.Image
}

// LoadArea loads the graphics layers of the given area.
func LoadArea(name string) (*Area, error) {
	area := &Area{
		Name: name,
	}
	// Background layer.
	backgroundLayerPath := FullPath(area.layerFileName(LayerKindBackground))
	backgroundLayer, err := imgutil.ReadFile(backgroundLayerPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	area.BackgroundLayer = backgroundLayer
	// Normal layer.
	normalLayerPath := FullPath(area.layerFileName(LayerKindNormal))
	normalLayer, err := imgutil.ReadFile(normalLayerPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	area.NormalLayer = normalLayer
	// Height layer.
	heightLayerPath := FullPath(area.layerFileName(LayerKindHeight))
	heightLayer, err := imgutil.ReadFile(heightLayerPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	area.HeightLayer = heightLayer
	// AS layer.
	asLayerPath := FullPath(area.layerFileName(LayerKindAS))
	asLayer, err := imgutil.ReadFile(asLayerPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	area.ASLayer = asLayer
	return area, nil
}

// layerFileName returns the file name of the specified layer for the given
// area.
func (a *Area) layerFileName(kind LayerKind) string {
	return fmt.Sprintf("%s_%s.png", a.Name, LayerKindName(kind))
}

// LayerKindName returns the name of the given layer kind.
func LayerKindName(kind LayerKind) string {
	m := map[LayerKind]string{
		LayerKindBackground:      "background",
		LayerKindBackgroundSmall: "thumb",
		LayerKindNormal:          "normal",
		LayerKindHeight:          "height",
		LayerKindAS:              "as", // TODO: rename AS.
	}
	if s, ok := m[kind]; ok {
		return s
	}
	panic(fmt.Errorf("support for layer kind %v not yet implemented", kind))
}

// AssetsDir specifies the game assets directory.
const AssetsDir = "_assets_"

// FullPath returns the full path to the specified game asset.
func FullPath(relPath string) string {
	path := filepath.Join(AssetsDir, relPath)
	return path
}
