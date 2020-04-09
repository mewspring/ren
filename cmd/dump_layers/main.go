package main

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"os"
	"path/filepath"

	"github.com/mewkiz/pkg/imgutil"
	"github.com/mewspring/ren/pkg/assets"
	"github.com/pkg/errors"
)

func main() {
	if err := dumpLayers(); err != nil {
		log.Fatalf("%+v", err)
	}
}

// dumpLayers dumps the layers of all maps in Pillars of Eternity.
func dumpLayers() error {
	for _, area := range areas {
		if err := area.Load(); err != nil {
			return errors.WithStack(err)
		}
		if err := area.Render(); err != nil {
			return errors.WithStack(err)
		}
		if err := area.Dump(); err != nil {
			return errors.WithStack(err)
		}
		// unload chunks.
		area.Imgs = nil
	}
	return nil
}

// Render renders the layers of the given map area.
func (area *Area) Render() error {
	width, height := area.getWidth(), area.getHeight()
	// Render background layer.
	area.BackgroundLayer = area.layer(assets.LayerKindBackground, width, height)
	// Render normal layer.
	area.NormalLayer = area.layer(assets.LayerKindNormal, width/2, height/2)
	// Render height layer.
	area.HeightLayer = area.layer(assets.LayerKindHeight, width, height)
	// Render AS layer.
	area.ASLayer = area.layer(assets.LayerKindAS, width/2, height/2)
	return nil
}

// Dump stores the layers of the given area to the output directory.
func (area *Area) Dump() error {
	// Create output directory.
	if err := os.MkdirAll(assets.AssetsDir, 0755); err != nil {
		return errors.WithStack(err)
	}
	// TODO: add support for background small (thumbnails)?
	// Output background layer.
	backgroundPath := area.layerPath(assets.LayerKindBackground)
	fmt.Printf("creating %q\n", backgroundPath)
	if err := imgutil.WriteFile(backgroundPath, area.BackgroundLayer); err != nil {
		return errors.WithStack(err)
	}
	// Output normal layer.
	normalPath := area.layerPath(assets.LayerKindNormal)
	fmt.Printf("creating %q\n", normalPath)
	if err := imgutil.WriteFile(normalPath, area.NormalLayer); err != nil {
		return errors.WithStack(err)
	}
	// Output height layer.
	heightPath := area.layerPath(assets.LayerKindHeight)
	fmt.Printf("creating %q\n", heightPath)
	if err := imgutil.WriteFile(heightPath, area.HeightLayer); err != nil {
		return errors.WithStack(err)
	}
	// Output AS layer.
	asPath := area.layerPath(assets.LayerKindAS)
	fmt.Printf("creating %q\n", asPath)
	if err := imgutil.WriteFile(asPath, area.ASLayer); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// layerPath returns the full path to the specified layer asset of the given
// area.
func (area *Area) layerPath(kind assets.LayerKind) string {
	layerName := fmt.Sprintf("%s_%s.png", area.Name, assets.LayerKindName(kind))
	return assets.FullPath(layerName)
}

// layer returns the specified layer of the given map area.
func (area *Area) layer(kind assets.LayerKind, width, height int) image.Image {
	bounds := image.Rect(0, 0, width, height)
	dst := image.NewRGBA(bounds)
	y := 0
	for row := area.NRows - 1; row >= 0; row-- {
		x := 0
		var srcHeight int
		for col := 0; col < area.NCols; col++ {
			src := area.chunk(kind, row, col)
			srcWidth := src.Bounds().Dx()
			srcHeight = src.Bounds().Dy()
			dr := image.Rect(x, y, x+srcWidth, y+srcHeight)
			sp := image.Pt(0, 0)
			draw.Draw(dst, dr, src, sp, draw.Src)
			x += srcWidth
		}
		y += srcHeight
	}
	return dst
}

// getWidth computes the width of the given area.
func (area *Area) getWidth() int {
	width := -1
	for row := 0; row < area.NRows; row++ {
		w := 0
		for col := 0; col < area.NCols; col++ {
			imgChunk := area.chunk(assets.LayerKindBackground, row, col)
			bounds := imgChunk.Bounds()
			w += bounds.Dx()
		}
		if width == -1 {
			width = w
		} else if w != width {
			panic(fmt.Errorf("mismatch between width of %q (prev=%d, new=%d)", area.Name, width, w))
		}
	}
	return width
}

// getHeight computes the height of the given area.
func (area *Area) getHeight() int {
	height := -1
	for col := 0; col < area.NCols; col++ {
		h := 0
		for row := 0; row < area.NRows; row++ {
			img := area.chunk(assets.LayerKindBackground, row, col)
			bounds := img.Bounds()
			h += bounds.Dy()
		}
		if height == -1 {
			height = h
		} else if h != height {
			panic(fmt.Errorf("mismatch between height of %q (prev=%d, new=%d)", area.Name, height, h))
		}
	}
	return height
}

// Layer file name format string.
const layerFormat = "%s_%s-R%03d_C%03d.png"

// chunk returns the chunk of the specified layer at row, col of the given map
// area.
func (area *Area) chunk(kind assets.LayerKind, row, col int) image.Image {
	// BKG_1501_yenwood-R000_C000.png
	imgName := fmt.Sprintf(layerFormat, kind, area.Name, row, col)
	img, ok := area.Imgs[imgName]
	if !ok {
		panic(fmt.Errorf("unable to locate %q (kind %q) of %q", imgName, kind, area.Name))
	}
	return img
}

// Pillars of Eternity assets directory.
const pillarsAssetsDir = "pillars_assets"

// Area is an area on the map.
type Area struct {
	// Area name.
	Name string
	// Number of rows.
	NRows int
	// Number of columns.
	NCols int
	// Maps from image file name to image contents.
	Imgs map[string]image.Image
	// Background layer of area.
	BackgroundLayer image.Image
	// Normal layer of area.
	NormalLayer image.Image
	// Height layer of area.
	HeightLayer image.Image
	// AS layer of area.
	ASLayer image.Image
}

// Load loads the game assets used by the given area.
func (area *Area) Load() error {
	fmt.Printf("loading graphics of %q\n", area.Name)
	area.Imgs = make(map[string]image.Image)
	for row := 0; row < area.NRows; row++ {
		for col := 0; col < area.NCols; col++ {
			// load background parts.
			if err := area.loadKind(assets.LayerKindBackground, row, col); err != nil {
				return errors.WithStack(err)
			}
			// load normal parts.
			if err := area.loadKind(assets.LayerKindNormal, row, col); err != nil {
				return errors.WithStack(err)
			}
			// load height parts.
			if err := area.loadKind(assets.LayerKindHeight, row, col); err != nil {
				return errors.WithStack(err)
			}
			// load AS parts.
			if err := area.loadKind(assets.LayerKindAS, row, col); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

// loadKind loads the game assets of the specified layer kind for the given area
// at row, col.
func (area *Area) loadKind(kind assets.LayerKind, row, col int) error {
	imgName := fmt.Sprintf(layerFormat, kind, area.Name, row, col)
	imgPath := filepath.Join(pillarsAssetsDir, imgName)
	img, err := imgutil.ReadFile(imgPath)
	if err != nil {
		return errors.WithStack(err)
	}
	area.Imgs[imgName] = img
	return nil
}

// Areas on the map.
var areas = []*Area{
	&Area{
		Name:  "1501_yenwood",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0009_Cliaban_Ruillaig_01",
		NRows: 8 + 1,
		NCols: 13 + 1,
	},
	&Area{
		Name:  "AR_0010_Cliaban_Ruillaig_02",
		NRows: 6 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0011_Dyrford_Tavern_02",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0012_Defiance_Bay_Copperlane",
		NRows: 7 + 1,
		NCols: 13 + 1,
	},
	&Area{
		Name:  "AR_0012_Dyrford_Temple",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0013_Dyrford_Mill",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0014_Hendynas_Interior",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0017_Def_Bay_Copperlane_House_01",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0018_Def_Bay_Copperlane_House_02",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0019_Def_Bay_Copperlane_House_03",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0020_Def_Bay_Copperlane_House_04",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0021_Defiance_Bay_Copperlane_House_05",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0103_Def_Bay_Ondras_Gift",
		NRows: 7 + 1,
		NCols: 13 + 1,
	},
	&Area{
		Name:  "AR_0107_DFB_Copperlane_Catacombs_01",
		NRows: 6 + 1,
		NCols: 11 + 1,
	},
	&Area{
		Name:  "AR_0108_Copperlane_Tavern_Lower",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0109_Copperlane_Tavern_Upper",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0110_Defiance_Bay_Copperlane_Expedition_Hall_01",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0111_Copperlane_Great_Library",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0112_Bridge_District",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0201_First_Fires_Exterior",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0207_First_Fires_Keep",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0208_First_Firers_Palace",
		NRows: 4 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0210_First_Fires_Embassy",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0302_DFB_ONG_Brothel_01",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0303_DFB_Trading_House",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0304_Ondras_Gift_House_01",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0305_Ondras_Gift_Lighthouse_01",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0306_Ondras_Gift_Lighthouse_02",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0307_Ondras_Gift_Lighthouse_03",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0311_Ondras_Gift_Lighthouse_01b",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0312_DFB_ONG_Brothel_02",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0312_Ondras_Gift_Oddas_House",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0312_Ondras_Gift_Poor_House_01",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0312_Ondras_Gift_Poor_House_02",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0314_Ondras_Gift_House_05",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0401_Brackenbury_Exterior",
		NRows: 4 + 1,
		NCols: 11 + 1,
	},
	&Area{
		Name:  "AR_0401_Brackenbury_Exterior_Riot",
		NRows: 3 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0402_Brackenbury_Hadret_House_Lower",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0402_Brackenbury_Hadret_House_Upper",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0404_Brackenbury_Sanitarium_01",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0405_Brackenbury_Sanitarium_02",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0407_Brackenbury_Reymont_Mannor_Upper",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0408_Brankenbury_House_Doemenel_Lower",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0408_Brankenbury_House_Doemenel_Upper",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0410_Brackenbury_Inn_Lower",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0411_Brackenbury_Inn_Upper",
		NRows: 6 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0412_Brackenbury_Inn_Cellar",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0501_Heritage_Hill_Ext",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0502_Heritage_Hill_Tower_01",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0503_Heritage_Hill_Tower_02",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0504_Heritage_Hill_Tower_03",
		NRows: 1 + 1,
		NCols: 2 + 1,
	},
	&Area{
		Name:  "AR_0505_Mausoleum_01",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0506_Mausoleum_02",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0508_Heritage_Hill_House_01",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0509_Heritage_Hill_Ground_01",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0601_StrongholdExterior",
		NRows: 10 + 1,
		NCols: 18 + 1,
	},
	&Area{
		Name:  "AR_0601_StrongholdExterior_Pristine",
		NRows: 10 + 1,
		NCols: 18 + 1,
	},
	&Area{
		Name:  "AR_0602_Brighthollow_Lower",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0602_Brighthollow_Lower_Destroyed",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0602_Brighthollow_Lower_Pristine",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0602_Damaged_Brighthollow_Lower",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0603_Brighthollow_Upper",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0603_Brighthollow_Upper_Pristine",
		NRows: 1 + 1,
		NCols: 2 + 1,
	},
	&Area{
		Name:  "AR_0604_Stronghold_Great_Hall",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0604_Stronghold_Great_Hall_Pristine",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0605_Damaged_Stronghold_Dungeon",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0605_Stronghold_Dungeon",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0605_Stronghold_Dungeon_Pristine",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0606_Damaged_Stronghold_Barracks",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0606_Stronghold_Barracks",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0606_Stronghold_Barracks_Pristine",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0607_Stronghold_Library",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0607_Stronghold_Library_Pristine",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0608_Warden_Lodge",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0609_Chapel",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0610_Craft_Shop",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0611_Artificer_Hall",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0612_Curio_Shop",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0701_Caravan_Encampment",
		NRows: 5 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0701_Gilded_Vale_Wilderness",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0702_Ruin_Int",
		NRows: 4 + 1,
		NCols: 0 + 1,
	},
	&Area{
		Name:  "AR_0703_Engwithan_Ruin_Ext",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0704_Valewood",
		NRows: 10 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0705_Gilded_Vale",
		NRows: 10 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0706_Esternwood",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0707_Raedrics_Hold_Ext",
		NRows: 7 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0708_Raedricks_Hold_Int_01",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0709_Raedricks_Hold_Int_02",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0710_Raedricks_Hold_Int_03",
		NRows: 5 + 1,
		NCols: 11 + 1,
	},
	&Area{
		Name:  "AR_0711_Temple_Eothas_Int_01",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0712_Temple_Eothas_Int_02",
		NRows: 0 + 1,
		NCols: 6 + 1,
	},
	&Area{
		Name:  "AR_0712_Temple_Eothas_Int_02_No_Bodys",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0713_The_Black_Hound_Inn_01",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0714_The_Black_Hound_Inn_02",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0715_Home_01",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0715_Home_02",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_0716_Blacksmith",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0717_Anslogs_Compass",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0718_Madhmr_Bridge",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0719_Windmill",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0801_Black_Meadow_Wilderness",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0802_Magrans_Fork",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0803_Elmshore",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0804_Stormwall_Gorge",
		NRows: 7 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0805_Pearlwood_Bluff",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0806_Searing_Falls",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_0807_Lle_a_Rhemen",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_0808_Lle_a_Rhemen_2",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0809_Pearlwood_Bluff_Cave",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0810_Generic_Cave_01",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0811_Woodend_Plains",
		NRows: 6 + 1,
		NCols: 11 + 1,
	},
	&Area{
		Name:  "AR_0812_TWE_Northweald",
		NRows: 6 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_0813_Searing_Falls_Drake",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_0816_Gilded_Vale_Hideout",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_1001_Od_Nua_Old_Watcher",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1002_Od_Nua_Xaurip_Base",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_1003_Od_Nua_Ogre_Lair",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1004_Od_Nua_Head",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_1005_Od_Nua_Drake",
		NRows: 3 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1006_Od_Nua_Catacombs",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "AR_1007_Od_Nua_Forge",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1008_Od_Nua_Vampires",
		NRows: 5 + 1,
		NCols: 13 + 1,
	},
	&Area{
		Name:  "AR_1009_Od_Nua_Ossuary",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1010_Od_Nua_Experiments",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1011_Od_Nua_Fungus",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1012_Od_Nua_Vithrak",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1013_Od_Nua_Banshees",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_1014_Od_Nua_Tomb",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1015_Od_Nua_Dragon",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_1101_Hearthsong_Ext",
		NRows: 5 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_1102_Hearthsong_Passage",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1103_Heathsong_Market",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1104_Hearthsong_Home_01",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_1105_Hearthsong_Home_02",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_1106_Hearthsong_Home_03",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_1107_Hearthsong_Inn",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1201_Oldsong_Ext",
		NRows: 10 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1202_Oldsong_The_Maw",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1203_Oldsong_Noonfrost",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1204_Oldsong_The_Nest",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1301_Twin_Elms_Exterior",
		NRows: 10 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_1303_Hall_Of_Stars",
		NRows: 3 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1304_Hall_Of_Warriors",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "AR_1305_Blood_Sands",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1307_Elms_Reach_Home_02",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_1308_Elms_Reach_Home_03",
		NRows: 2 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "AR_1401_Burial_Isle_Ext",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1402_Court_Of_The_Penitents",
		NRows: 6 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "AR_1404_Sun_In_Shadow_01",
		NRows: 5 + 1,
		NCols: 22 + 1,
	},
	&Area{
		Name:  "AR_1405_Sun_In_Shadow_02",
		NRows: 12 + 1,
		NCols: 14 + 1,
	},
	&Area{
		Name:  "AR_1405_Sun_In_Shadow_02_NoStatues",
		NRows: 7 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "AR_First_Fires_Reymont_Mannor",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "DFB_FirstFires_Ruins",
		NRows: 4 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "PRO_Inn_Int_01",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "PRO_OgreCave_A",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "PRO_Shop_Int_01",
		NRows: 1 + 1,
		NCols: 1 + 1,
	},
	&Area{
		Name:  "PRO_Tanner_Int_04",
		NRows: 2 + 1,
		NCols: 3 + 1,
	},
	&Area{
		Name:  "PRO_Village_A",
		NRows: 10 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "PRO_Wilderness_A",
		NRows: 10 + 1,
		NCols: 9 + 1,
	},
	&Area{
		Name:  "PX4_Cave01",
		NRows: 4 + 1,
		NCols: 7 + 1,
	},
	&Area{
		Name:  "Prototyoe_2_Dungeon_01",
		NRows: 6 + 1,
		NCols: 11 + 1,
	},
	&Area{
		Name:  "PrototypeExterior",
		NRows: 3 + 1,
		NCols: 5 + 1,
	},
	&Area{
		Name:  "Prototype_Interior_02",
		NRows: 6 + 1,
		NCols: 9 + 1,
	},
}
