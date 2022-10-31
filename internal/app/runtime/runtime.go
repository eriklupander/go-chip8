package runtime

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"image/color"
	"sync"
)

var (
	colorWhite = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	colorBlack = color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xFF}
)

func NewUIRuntime() *Runtime {

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, CHIP-8!")

	initializedImage := ebiten.NewImage(64, 32)
	initializedImage.Fill(color.RGBA{}) // initialize all pixels to black, 0 alpha.

	alphaColorM := ebiten.ColorM{}
	alphaColorM.Translate(1.0, 1.0, 1.0, -0.25)

	game := &Runtime{
		image:       image.NewRGBA(image.Rect(0, 0, 64, 32)),
		ghostImage:  initializedImage,
		tmpImage:    ebiten.NewImage(64, 32),
		lock:        sync.Mutex{},
		alphaColorM: alphaColorM,
	}
	game.ClearScreen() // always clear screen to init all pixels to black
	return game
}

// Runtime provides an ebitengine "game" implementation with fields for the images the drawing "canvas" consists of:
// image: The main screen "memory"
// ghostImage: Provides a small "ghost" aka CRT effect where previously lit pixels are faded out, overlaid on the image.
// tmpImage: Used as intermediate image on each frame update where the ghostImage is multiplied by alphaColorM to accomplish the fade effect.
type Runtime struct {
	keys        []ebiten.Key
	image       *image.RGBA
	ghostImage  *ebiten.Image
	tmpImage    *ebiten.Image
	alphaColorM ebiten.ColorM
	lock        sync.Mutex
}

func (g *Runtime) WaitForKeypress(keypresses chan byte) {

	for _, key := range g.keys {
		out := keyToByte(key)
		keypresses <- out
	}
}

func (g *Runtime) ClearScreen() {
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			g.image.Set(x, y, color.Black)
		}
	}
}

func (g *Runtime) Update() error {

	g.lock.Lock()
	g.tmpImage.Clear()
	g.tmpImage.DrawImage(g.ghostImage, &ebiten.DrawImageOptions{ColorM: g.alphaColorM})
	g.ghostImage.Clear()
	g.ghostImage.DrawImage(g.tmpImage, &ebiten.DrawImageOptions{})

	g.lock.Unlock()

	g.keys = inpututil.AppendPressedKeys(g.keys[:0])

	return nil
}

func (g *Runtime) Draw(screen *ebiten.Image) {
	g.lock.Lock()
	screen.WritePixels(g.image.Pix)
	screen.DrawImage(g.ghostImage, &ebiten.DrawImageOptions{})
	g.lock.Unlock()
}

func (g *Runtime) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 64, 32
}

func (g *Runtime) IsPixelSet(col int, row int) bool {
	g.lock.Lock()
	isSet := g.image.RGBAAt(col, row) == colorWhite
	g.lock.Unlock()
	// if pixel is already "on", we turn off the pixel.
	return isSet
}

func (g *Runtime) Set(col int, row int, on bool) {
	g.lock.Lock()
	if on {
		g.image.Set(col, row, colorWhite)
	} else {
		g.image.Set(col, row, colorBlack)

		// draw a "shadow" pixel where the previously lit pixel was
		g.ghostImage.Set(col, row, colorWhite)
	}
	g.lock.Unlock()
}

func keyToByte(key ebiten.Key) byte {
	var out byte
	switch key {
	case ebiten.KeyDigit0:
		out = 0x0
	case ebiten.KeyDigit1:
		out = 0x1
	case ebiten.KeyDigit2:
		out = 0x2
	case ebiten.KeyDigit3:
		out = 0x3
	case ebiten.KeyDigit4:
		out = 0x4
	case ebiten.KeyDigit5:
		out = 0x5
	case ebiten.KeyDigit6:
		out = 0x6
	case ebiten.KeyDigit7:
		out = 0x7
	case ebiten.KeyDigit8:
		out = 0x8
	case ebiten.KeyDigit9:
		out = 0x9
	case ebiten.KeyA:
		out = 0xA
	case ebiten.KeyB:
		out = 0xB
	case ebiten.KeyC:
		out = 0xC
	case ebiten.KeyD:
		out = 0xD
	case ebiten.KeyE:
		out = 0xE
	case ebiten.KeyF:
		out = 0xF
	}
	return out
}

func ByteToKey(b byte) ebiten.Key {
	var out ebiten.Key
	switch b {
	case 0x0:
		return ebiten.KeyDigit0
	case 0x1:
		return ebiten.KeyDigit1
	case 0x2:
		return ebiten.KeyDigit2
	case 0x3:
		return ebiten.KeyDigit3
	case 0x4:
		return ebiten.KeyDigit4
	case 0x5:
		return ebiten.KeyDigit5
	case 0x6:
		return ebiten.KeyDigit6
	case 0x7:
		return ebiten.KeyDigit7
	case 0x8:
		return ebiten.KeyDigit8
	case 0x9:
		return ebiten.KeyDigit9
	case 0xA:
		return ebiten.KeyA
	case 0xB:
		return ebiten.KeyB
	case 0xC:
		return ebiten.KeyC
	case 0xD:
		return ebiten.KeyD
	case 0xE:
		return ebiten.KeyE
	case 0xF:
		return ebiten.KeyF
	}
	return out
}
