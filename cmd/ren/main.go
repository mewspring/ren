package main

import (
	"fmt"
	"image"
	"log"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/kr/pretty"
	"github.com/mewkiz/pkg/imgutil"
	"github.com/mewspring/ren/pkg/assets"
	"github.com/pkg/errors"
)

func main() {
	const (
		width  = 1280
		height = 768
		scale  = 1.0
		title  = "ren"
	)
	game := &Game{}
	if err := ebiten.Run(game.run, width, height, scale, title); err != nil {
		log.Fatalf("%+v", err)
	}
}

// Game holds game state.
type Game struct {
	// Specifies whether game assets have been loaded.
	assetsLoaded bool
	// Yenwood map area.
	yenwood                  *assets.Area
	yenwoodBackgroundLayer   *ebiten.Image
	balrogUnit               *ebiten.Image
	balrogStandAnimFromDir   [ndirs][]*ebiten.Image
	balrogWalkAnimFromDir    [ndirs][]*ebiten.Image
	balrogAttackAnimFromDir  [ndirs][]*ebiten.Image
	balrogHitAnimFromDir     [ndirs][]*ebiten.Image
	balrogDeathAnimFromDir   [ndirs][]*ebiten.Image
	balrogSpecialAnimFromDir [ndirs][]*ebiten.Image
	tx, ty                   float64
	balrogStandAnim          *Anim
	balrogWalkAnim           *Anim
	balrogAttackAnim         *Anim
}

// Number of directions.
const ndirs = 8

// run is invoked by Ebiten 60 times per second.
func (game *Game) run(screen *ebiten.Image) error {
	// Load game assets.
	if !game.assetsLoaded {
		if err := game.loadAssets(); err != nil {
			return errors.WithStack(err)
		}
		game.assetsLoaded = true
	}
	// Render to screen.
	opt := &ebiten.DrawImageOptions{}
	//fmt.Printf("translate with %v,%v\n", game.tx, game.ty)
	opt.GeoM.Scale(0.5, 0.5)
	game.tx += -0.5
	game.ty += -0.5
	if err := screen.DrawImage(game.yenwoodBackgroundLayer, opt); err != nil {
		return errors.WithStack(err)
	}

	const dir = 6
	attackAnim := game.balrogAttackAnimFromDir[dir]
	opt2 := &ebiten.DrawImageOptions{}
	opt2.GeoM.Translate(300, 300)
	game.balrogAttackAnim.Update()
	if curAttackFrame, ok := game.balrogAttackAnim.FrameNum(); ok {
		if err := screen.DrawImage(attackAnim[curAttackFrame], opt2); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// loadAssets loads game assets.
func (game *Game) loadAssets() error {
	fmt.Printf("loading assets\n")
	// Load Yenwood area.
	yenwoodArea, err := assets.LoadArea("yenwood")
	if err != nil {
		return errors.WithStack(err)
	}
	game.yenwood = yenwoodArea
	yenwoodBackgroundLayer, err := ebiten.NewImageFromImage(yenwoodArea.BackgroundLayer, ebiten.FilterDefault)
	if err != nil {
		return errors.WithStack(err)
	}
	game.yenwoodBackgroundLayer = yenwoodBackgroundLayer
	// Load Balrog unit.
	if err := game.loadBalrogAssets(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (game *Game) loadBalrogAssets() error {
	balrogUnitImg, err := imgutil.ReadFile(assets.FullPath(filepath.Join("monsters", "balrog.png")))
	if err != nil {
		return errors.WithStack(err)
	}
	balrogUnit, err := ebiten.NewImageFromImage(balrogUnitImg, ebiten.FilterDefault)
	if err != nil {
		return errors.WithStack(err)
	}
	game.balrogUnit = balrogUnit
	const (
		// render_size=160,160
		// render_offset=80,80
		frameWidth  = 160
		frameHeight = 160
		// [stance]
		// position=0
		// frames=6
		// duration=300ms
		// type=back_forth
		standFirstFrame = 0
		standNFrames    = 6
		// [run]
		// position=6
		// frames=7
		// duration=7ms
		// type=looped
		walkFirstFrame = 6
		walkNFrames    = 7
		// [swing]
		// position=13
		// frames=14
		// duration=14ms
		// type=play_once
		attackFirstFrame = 13
		attackNFrames    = 14
		// [hit]
		// position=27
		// frames=1
		// duration=1ms
		// type=play_once
		hitFirstFrame = 27
		hitNFrames    = 1
		// [die]
		// position=28
		// frames=24
		// duration=24ms
		// type=play_once
		deathFirstFrame = 28
		deathNFrames    = 24 - 1 // TODO: set to 24 when revering to orig graphics.
		// [shoot]
		// position=52
		// frames=5
		// duration=5ms
		// type=play_once
		specialFirstFrame = 52
		specialNFrames    = 5 - 5 // TODO: set to 5 when reverting to orig graphics.
	)
	balrogStandAnim := &Anim{
		FirstFrame: standFirstFrame,
		NFrames:    standNFrames,
		Dur:        standNFrames * 50 * time.Millisecond, // 50 * time.Millisecond * nframes
		AnimType:   AnimTypeBackForth,
		Inc:        1,
	}
	game.balrogStandAnim = balrogStandAnim
	balrogWalkAnim := &Anim{
		FirstFrame: walkFirstFrame,
		NFrames:    walkNFrames,
		Dur:        walkNFrames * 50 * time.Millisecond, // 50 * time.Millisecond * nframes
		AnimType:   AnimTypeLoop,
	}
	game.balrogWalkAnim = balrogWalkAnim
	balrogAttackAnim := &Anim{
		FirstFrame: attackFirstFrame,
		NFrames:    attackNFrames,
		Dur:        attackNFrames * 50 * time.Millisecond, // 50 * time.Millisecond * nframes
		AnimType:   AnimTypeLoop,
	}
	game.balrogAttackAnim = balrogAttackAnim
	for dir := 0; dir < 8; dir++ {
		// stand anims.
		var standFrames []*ebiten.Image
		for i := 0; i < standNFrames; i++ {
			frameNum := i + standFirstFrame
			x := frameNum * frameWidth
			y := dir * frameHeight
			r := image.Rect(x, y, x+frameWidth, y+frameHeight)
			pretty.Println("rect:", r)
			frame := balrogUnit.SubImage(r)
			if dir == 0 && i == 0 {
				if err := imgutil.WriteFile(fmt.Sprintf("balrog_frame_0.png"), frame); err != nil {
					return errors.WithStack(err)
				}
			}
			standFrames = append(standFrames, frame.(*ebiten.Image))
		}
		game.balrogStandAnimFromDir[dir] = standFrames

		// walk anims.
		var walkFrames []*ebiten.Image
		for i := 0; i < walkNFrames; i++ {
			frameNum := i + walkFirstFrame
			x := frameNum * frameWidth
			y := dir * frameHeight
			r := image.Rect(x, y, x+frameWidth, y+frameHeight)
			pretty.Println("rect:", r)
			frame := balrogUnit.SubImage(r)
			if dir == 0 && i == 0 {
				if err := imgutil.WriteFile(fmt.Sprintf("balrog_frame_0.png"), frame); err != nil {
					return errors.WithStack(err)
				}
			}
			walkFrames = append(walkFrames, frame.(*ebiten.Image))
		}
		game.balrogWalkAnimFromDir[dir] = walkFrames

		// attack anims.
		var attackFrames []*ebiten.Image
		for i := 0; i < attackNFrames; i++ {
			frameNum := i + attackFirstFrame
			x := frameNum * frameWidth
			y := dir * frameHeight
			r := image.Rect(x, y, x+frameWidth, y+frameHeight)
			pretty.Println("rect:", r)
			frame := balrogUnit.SubImage(r)
			if dir == 0 && i == 0 {
				if err := imgutil.WriteFile(fmt.Sprintf("balrog_frame_0.png"), frame); err != nil {
					return errors.WithStack(err)
				}
			}
			attackFrames = append(attackFrames, frame.(*ebiten.Image))
		}
		game.balrogAttackAnimFromDir[dir] = attackFrames

		// TODO: load other anims.
	}
	return nil
}

// Anim is a graphics animation.
type Anim struct {
	// First frame number of graphics animation in sprite sheet.
	FirstFrame int
	// Number of frames in graphics animation.
	NFrames int
	// Duration of graphics animation (e.g. 300ms).
	Dur time.Duration
	// Animation type (e.g. play once, loop, back-and-forth).
	AnimType AnimType
	// Anim frame number increment (+1 or -1). Used by back-and-forth animation
	// to determine direction of animation sequence.
	Inc int
	// Current frame.
	CurFrame int
	// Time of last frame update.
	LastUpdate time.Time
}

// AnimType specifies an animation type.
type AnimType uint8

// Animation types.
const (
	// Play once from first to last frame.
	//
	//    0, 1, 2, 3, 4, 5
	AnimTypeOnce AnimType = iota + 1
	// Play in a loop.
	//
	//    0, 1, 2, 3, 4, 5,
	//    0, 1, 2, 3, 4, 5,
	//    ...
	AnimTypeLoop
	// Play in a loop, back-and-forth.
	//
	//    0, 1, 2, 3, 4, 5,
	//       4, 3, 2, 1,
	//    0, 1, 2, 3, 4, 5,
	//    ...
	AnimTypeBackForth
	// Show still image.
	//
	//    0,
	//    0,
	//    ...
	AnimTypeStill
)

// Update updates the current frame number of enough time has passed since last
// frame update. The boolean return value indicates that a frame update took
// place.
func (anim *Anim) Update() bool {
	durPerFrame := anim.Dur / time.Duration(anim.NFrames)
	if time.Since(anim.LastUpdate) < durPerFrame {
		return false
	}
	anim.LastUpdate = time.Now() // TODO: handle skip of frames if too long since last update.
	switch anim.AnimType {
	case AnimTypeOnce:
		anim.CurFrame++
	case AnimTypeLoop:
		anim.CurFrame++
		if anim.CurFrame >= anim.NFrames {
			anim.CurFrame = 0
		}
	case AnimTypeBackForth:
		anim.CurFrame += anim.Inc
		switch {
		case anim.Inc < 0:
			// back
			if anim.CurFrame < 0 {
				anim.CurFrame = 1 // TODO: start at 1?
				anim.Inc = 1
			}
		case anim.Inc > 0:
			// forth
			if anim.CurFrame >= anim.NFrames {
				anim.CurFrame = anim.NFrames - 2 // TODO: start at anim.NFrames - 2?
				anim.Inc = -1
			}
		default:
			// inc == 0
			panic(fmt.Errorf("invalid increment for back-forth animation mode; expected +1 or -1, got %d", anim.Inc))
		}
	case AnimTypeStill:
		// keep current frame as is.
	default:
		panic(fmt.Errorf("support for animation type %v not yet implemented", anim.AnimType))
	}
	return true
}

// FrameNum returns the current frame number. The boolean return value indicates
// if the animation is still playing.
func (anim *Anim) FrameNum() (int, bool) {
	switch anim.AnimType {
	case AnimTypeOnce:
		if anim.CurFrame >= anim.NFrames {
			fmt.Println("false")
			return 0, false
		}
	case AnimTypeLoop:
	case AnimTypeBackForth:
	case AnimTypeStill:
		// keep current frame as is.
	default:
		panic(fmt.Errorf("support for animation type %v not yet implemented", anim.AnimType))
	}
	fmt.Println("true, curFrame:", anim.CurFrame)
	return anim.CurFrame, true
}
