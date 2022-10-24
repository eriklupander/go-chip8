package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image/color"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	memOffset  = 0x200
	fontOffset = 0x50
)

type Game struct {
	Screen *[64][32]bool
	Keys   []ebiten.Key
}

func (g *Game) WaitForKeypress(keypresses chan byte) {

	for _, key := range g.Keys {
		out := keyToByte(key)
		keypresses <- out
	}
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

func byteToKey(b byte) ebiten.Key {
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

func (g *Game) ClearScreen() {
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			g.Screen[x][y] = false
		}
	}
	//fmt.Print("\033[2J")
	//fmt.Printf("\033[%d;%dH", 1, 1)
}

func (g *Game) Update() error {

	g.Keys = inpututil.AppendPressedKeys(g.Keys[:0])
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// we should optimize this and re-use a texture and instead only update pixels that need updating from the
	// interpreter loop.
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if g.Screen[x][y] {
				screen.Set(x, y, color.White)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 64, 32
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")

	game := &Game{Screen: &[64][32]bool{}}
	go interpretor(game)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

var (
	stack      = [32]uint16{} // The stack offers a max depth of 32 with 2 bytes per stack frame
	stackFrame = -1
	memory     = [4096]byte{}      // 4kb of internal memory
	I          = uint16(0)         // represents Index register
	registers  = [16]byte{}        // represents the 16 1-byte registers
	pc         = uint16(memOffset) // Program counter, set it to the initial memory offset
	delayTimer = byte(60)
	soundTimer = byte(0x0)
	//display    = [64][32]bool{}    // repesents graphics memory. true == pixel is active
)

func startTimerLoop() {
	var tick = 1000 / 60
	for {
		time.Sleep(time.Millisecond * time.Duration(tick))
		if delayTimer > 0 {
			delayTimer--
		}
	}
}
func startSoundLoop() {
	var tick = 1000 / 60
	for {
		time.Sleep(time.Millisecond * time.Duration(tick))
		if soundTimer > 0 {
			soundTimer--
		}
	}
}
func interpretor(game *Game) {

	go startTimerLoop()
	go startSoundLoop()

	// 1. load program into memory.
	romData, err := os.ReadFile("roms/test_opcode.ch8")
	if err != nil {
		panic("error reading ROM")
	}

	// 2. Stuff fonts into internal memory as defined by fontOffset.
	for i := range font {
		memory[fontOffset+i] = font[i]
	}
	// 3. Stuff program into memory starting at memOffset
	for i := range romData {
		memory[memOffset+i] = romData[i]
	}

	// 4. Run main interpretator loop
	for {
		// FETCH
		// Read 2 bytes from memory, designated by the value of program counter (PC).
		b0 := memory[pc]
		b1 := memory[pc+1]
		pc += 2

		// use for outputting instructions to stdout.
		//fmt.Printf("%02X%02X\n", b0, b1)

		// DECODE
		instr := (b0 & 0xF0) >> 4        // first nibble, the instruction
		X := b0 & 0x0F                   // second nibble, register lookup!
		Y := (b1 & 0xF0) >> 4            // third nibble, register lookup!
		N := b1 & 0x0F                   // N = fourth nibble, 4 bit number
		NN := b1                         // NN = second byte
		NNN := uint16(X)<<8 | uint16(NN) // NNN = second, third and fourth nibbles

		// EXECUTE
		switch instr {
		case 0x0:
			switch Y {
			case 0xE:
				switch N {
				case 0x0: // clear screen
					game.ClearScreen()
				case 0xE: // pop stack
					pc = stack[stackFrame] // remember - this is actually the "parent" stack frame
					stackFrame--
				}

			}
		case 0x1: // Jump PC to NNN
			pc = NNN
		case 0x2: // Subroutine: Push to stack, then set PC to NNN
			stackFrame++
			stack[stackFrame] = pc // store _current_ program counter in the NEXT stack frame.
			pc = NNN
		case 0x3: // Skip if value in register X equals NN
			if registers[X] == NN {
				pc += 2
			}
		case 0x4: // Skip if value in register X not equals NN
			if registers[X] != NN {
				pc += 2
			}
		case 0x5: // Skip if values in registers X and Y are equal
			if registers[X] == registers[Y] {
				pc += 2
			}
		case 0x6: // Set register X to NN
			registers[X] = NN
		case 0x7: // Add NN to register X
			registers[X] = registers[X] + NN
		case 0x8:
			switch N {
			case 0x0: // Set register X to value of register Y
				b := registers[Y]
				registers[X] = b
			case 0x1: // Set register X to OR of registers X and Y
				registers[X] = registers[X] | registers[Y]
			case 0x2: // Set register X to AND of registers X and Y
				registers[X] = registers[X] & registers[Y]
			case 0x3: // Set register X to XOR of registers X and Y
				registers[X] = registers[X] ^ registers[Y]
			case 0x4: // Set register X to X + Y, set register F (15) to 1 or 0 depending on overflow
				registers[X] = registers[X] + registers[Y]
				if registers[X] > 0xFF {
					registers[0xF] = 0x1
				} else {
					registers[0xF] = 0x0
				}
			case 0x5: // Subtract: set register X to the result of registers X - Y.
				registers[X] = registers[X] - registers[Y]
			case 0x6: // Shift register X one step to the right
				if registers[X]&(1<<0) > 0 {
					registers[0xF] = 0x1
				} else {
					registers[0xF] = 0x0
				}
				registers[X] = registers[X] >> 1
			case 0x7: // Subtract: set register X to the result of registers Y - X.
				registers[X] = registers[Y] - registers[X]
			case 0xE: // Shift register X one step to the left
				if registers[X]&(1<<7) > 0 {
					registers[0xF] = 0x1
				} else {
					registers[0xF] = 0x0
				}
				registers[X] = registers[X] << 1
			}

		case 0x9: // Skip if values in registers X and Y are not equal
			if registers[X] != registers[Y] {
				pc += 2
			}
		case 0xA: // Set Index register to NNN
			I = NNN
		case 0xB: // May need to be made configurable!
			I = NNN + uint16(registers[0x0]) // original behaviour, assume register 0x0.
		case 0xC: // Random number
			rnd := rand.Uint32()
			registers[X] = byte(rnd) & NN
		case 0xD: // draw sprite at I at screen x, y given by values in registers X and Y.
			xCoord := registers[X] % 64
			yCoord := registers[Y] % 32
			registers[0xF] = 0x0
			firstByteIndex := I
			numLines := int(N)

			for line := 0; line < numLines; line++ {

				for bit := 0; bit < 8; bit++ {

					// check if bit is set
					if memory[firstByteIndex]&(1<<bit) > 0 {
						row := int(yCoord) + line
						col := int(xCoord) + 8 - bit // note endianess fix here, we draw from right to left.

						// temp: If either val is out of bounds, log and skip

						//if row >= 32 {
						//	fmt.Printf("got row out of bounds: %v\n", row)
						//	continue
						//}
						if col >= 64 {
							col = 63
						}
						//	fmt.Printf("got col out of bounds: %v\n", col)
						//	continue
						//}

						// if pixel is already "on", we turn off the pixel.
						if game.Screen[col][row] {
							// turn off pixel
							game.Screen[col][row] = false

							// clear pixel by writing a whitespace
							//fmt.Printf("\033[%d;%dH", row, col)
							//fmt.Printf(" ")

							// set register F to 1
							registers[0xF] = 0x1
						} else {
							// draw x+bit on line y
							//fmt.Printf("\033[%d;%dH", row, col)
							//fmt.Printf("O")
							game.Screen[col][row] = true
						}
					}
				}
				firstByteIndex++
			}
		case 0xE: // handle key presses EX9E and EXA1
			switch NN {
			case 0x9E: // pressed
				if ebiten.IsKeyPressed(byteToKey(registers[X])) {
					pc += 2
				}
			case 0xA1: // not pressed
				if !ebiten.IsKeyPressed(byteToKey(registers[X])) {
					pc += 2
				}
			}
		case 0xF: // timers
			switch NN {
			case 0x07: // Set register X to current value of delay timer
				registers[X] = delayTimer // should use lock...
			case 0x15: // Set the delay timer to value of register X
				delayTimer = registers[X]
			case 0x18: // Set the sound timer to value of register X
				soundTimer = registers[X]
			case 0xE1: // Add to index: Add value of register X to I
				lastI := I
				I = I + uint16(registers[X])

				// old-school amiga behaviour
				if lastI < 0x1000 && I >= 0x1000 {
					registers[0xF] = 0x1
				}
			case 0x0A: // Get key (blocks until input is received) TODO this one is not complete!!
				// input chan should be cleared beforehand?

				keypresses := make(chan byte)
				go game.WaitForKeypress(keypresses)
				key := <-keypresses

				registers[X] = key
				//pc = pc - 2 // decrease PC by 2... if the "wait for input" is cancelled by a timer...?
			case 0x29: // font character
				I = uint16(fontOffsets[registers[X]]) // possibly, just use last nibble of value in register X
			case 0x33: // binary-coded decimal conversion
				conv := binaryDecimalConversion(registers[X])
				for i := uint16(0); i < uint16(len(conv)); i++ {
					memory[I+i] = conv[i]
				}
			case 0x55: // Store register to memory
				for i := 0; i <= int(X); i++ {
					index := I + uint16(i)
					memory[index] = registers[i]
				}
			case 0x65: // Load value from memory into register
				for i := 0; i <= int(X); i++ {
					index := I + uint16(i)
					registers[i] = memory[index]
				}
			}
		default:
			panic("Unknown instruction: " + fmt.Sprintf("%X", instr))
		}
		time.Sleep(time.Microsecond * 1300) // corresponds to about 700 instructions per second...
	}
}

func binaryDecimalConversion(dec byte) []byte {

	out := make([]byte, 0)

	p3 := dec / 100
	p2 := dec % 100 / 10
	p1 := dec % 10
	if p3 != 0 {
		out = append(out, p3)
	}
	if p2 != 0 || p3 != 0 {
		out = append(out, p2)
	}
	out = append(out, p1)
	return out
}

// temp hack
var keys = make([]ebiten.Key, 0)

func keypress(check byte) bool {
	keys = inpututil.AppendPressedKeys(keys[:0])
	for i := range keys {
		if keyToByte(keys[i]) == check {
			return true
		}
	}
	return false
}

// Extras: The stuff below this point is just used for learning/debugging purposes.
var fontOffsets = map[byte]int{
	0x0: 0,
	0x1: 5,
	0x2: 10,
	0x3: 15,
	0x4: 20,
	0x5: 25,
	0x6: 30,
	0x7: 35,
	0x8: 40,
	0x9: 45,
	0xA: 50,
	0xB: 55,
	0xC: 60,
	0xD: 65,
	0xE: 70,
	0xF: 75,
}
var offsets = map[rune]int{
	'0': 0,
	'1': 5,
	'2': 10,
	'3': 15,
	'4': 20,
	'5': 25,
	'6': 30,
	'7': 35,
	'8': 40,
	'9': 45,
	'A': 50,
	'B': 55,
	'C': 60,
	'D': 65,
	'E': 70,
	'F': 75,
}
var font = []byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80} // F
