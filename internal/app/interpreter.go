package app

import (
	"fmt"
	"github.com/eriklupander/go-chip8/internal/app/runtime"
	"github.com/hajimehoshi/ebiten/v2"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	memOffset  = 0x200
	fontOffset = 0x50
)

type emulator struct {
	stack          [32]uint16 // The stack offers a max depth of 32 with 2 bytes per stack frame
	stackFrame     int        // current stack frame. Starts at -1 and is set to 0 on first use
	memory         [4096]byte // 4kb of internal memory
	I              uint16     // represents Index register
	registers      [16]byte   // represents the 16 1-byte registers
	pc             uint16     // Program counter, set it to the initial memory offset
	delayTimer     byte       // represents the delay timer that's decremented at 60hz if > 0
	soundTimer     byte       // represents the sound timer that's decremented at 60hz and plays a beep if > 0.
	delayTimerLock sync.Mutex // lock for incrementing/setting/accessing the delay timer
	soundTimerLock sync.Mutex // lock for incrementing/setting/accessing the sound timer
}

func NewEmulator() *emulator {
	return &emulator{
		stack:          [32]uint16{},
		stackFrame:     -1,
		memory:         [4096]byte{},
		I:              0x0,
		registers:      [16]byte{},
		pc:             memOffset,
		delayTimer:     0x0,
		soundTimer:     0x0,
		delayTimerLock: sync.Mutex{},
		soundTimerLock: sync.Mutex{},
	}
}

func (e *emulator) startDelayTimer() {
	var tick = 1000 / 60
	for {
		time.Sleep(time.Millisecond * time.Duration(tick))
		e.delayTimerLock.Lock()
		if e.delayTimer > 0 {
			e.delayTimer--
		}
		e.delayTimerLock.Unlock()
	}
}
func (e *emulator) startSoundTimer(r *runtime.Runtime) {
	var tick = 1000 / 60
	for {
		time.Sleep(time.Millisecond * time.Duration(tick))
		e.soundTimerLock.Lock()
		if e.soundTimer > 0 {
			e.soundTimer--
			r.PlayAudio()
		} else {
			r.StopAudio()
		}
		e.soundTimerLock.Unlock()
	}
}
func (e *emulator) Run(game *runtime.Runtime, romFile string) {

	// launch timer loops
	go e.startDelayTimer()
	go e.startSoundTimer(game)

	// 1. load program into memory.
	romData, err := os.ReadFile(romFile)
	if err != nil {
		panic("error reading ROM: " + romFile)
	}

	// 2. Stuff fonts into internal memory as defined by fontOffset.
	for i := range font {
		e.memory[fontOffset+i] = font[i]
	}
	// 3. Stuff program into memory starting at memOffset
	for i := range romData {
		e.memory[memOffset+i] = romData[i]
	}

	// 4. Run main interpretator loop
	for {
		// FETCH
		// Read 2 bytes from memory, designated by the value of program counter (PC).
		b0 := e.memory[e.pc]
		b1 := e.memory[e.pc+1]
		e.pc += 2

		//instruction := fmt.Sprintf("%02x%02x", b0, b1)
		// use for outputting instructions to stdout.
		//fmt.Printf("%02X%02X\n", b0, b1)

		// DECODE
		instr := (b0 & 0xF0) >> 4        // first nibble, the instruction
		X := b0 & 0x0F                   // second nibble, register lookup!
		Y := (b1 & 0xF0) >> 4            // third nibble, register lookup!
		N := b1 & 0x0F                   // fourth nibble, 4 bit number
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
					e.pc = e.stack[e.stackFrame] // remember - this is actually the "parent" stack frame
					e.stackFrame--
				default:
					panic("Unknown instruction: " + fmt.Sprintf("%02x", instr))
				}
			default:
				panic("Unknown instruction: " + fmt.Sprintf("%02x", instr))
			}
		case 0x1: // Jump PC to NNN
			e.pc = NNN
		case 0x2: // Subroutine: Push to stack, then set PC to NNN
			e.stackFrame++
			e.stack[e.stackFrame] = e.pc // store _current_ program counter in the NEXT stack frame.
			e.pc = NNN
		case 0x3: // Skip if value in register X equals NN
			if e.registers[X] == NN {
				e.pc += 2
			}
		case 0x4: // Skip if value in register X not equals NN
			if e.registers[X] != NN {
				e.pc += 2
			}
		case 0x5: // Skip if values in registers X and Y are equal
			if N == 0x0 && e.registers[X] == e.registers[Y] {
				e.pc += 2
			}
		case 0x6: // Set register X to NN
			e.registers[X] = NN
		case 0x7: // Add NN to register X
			//fmt.Printf("%s: ADD: val before add in register %d: %d: Add by: %d ", instruction, X, e.registers[X], NN)
			e.registers[X] = e.registers[X] + NN
			//fmt.Printf(": resulting in new value %d\n", e.registers[X])
		case 0x8:
			switch N {
			case 0x0: // Set register X to value of register Y
				b := e.registers[Y]
				e.registers[X] = b
			case 0x1: // Set register X to OR of registers X and Y
				e.registers[X] = e.registers[X] | e.registers[Y]
			case 0x2: // Set register X to AND of registers X and Y
				e.registers[X] = e.registers[X] & e.registers[Y]
			case 0x3: // Set register X to XOR of registers X and Y
				e.registers[X] = e.registers[X] ^ e.registers[Y]
			case 0x4: // Set register X to X + Y, set register F (15) to 1 or 0 depending on overflow
				//fmt.Printf("%s: Set X to X+Y: registers: X: %d, Y: %d. Result %d+%d = %d\n", instruction, X, Y, e.registers[X], e.registers[Y], (e.registers[X] + e.registers[Y]))
				vx := e.registers[X]
				result := vx + e.registers[Y]
				e.registers[X] = result
				if result < vx { // if result is less than original, we've had an overflow
					e.registers[0xF] = 0x1
				} else {
					e.registers[0xF] = 0x0
				}
			case 0x5: // Subtract: set register X to the result of registers X - Y.
				//fmt.Printf("%s: Subtract register X-Y reg-%d - reg-%d with values %d - %d\n", instruction, X, Y, e.registers[X], e.registers[Y])

				if e.registers[X] > e.registers[Y] {
					e.registers[0xF] = 0x1
				} else {
					e.registers[0xF] = 0x0
				}
				e.registers[X] = e.registers[X] - e.registers[Y]
			case 0x6: // Shift register X one step to the right
				// check if rightmost bit is set (and shifted out)
				//fmt.Printf("%s: Shift right. Was: %08b", instruction, e.registers[Y])
				e.registers[X] = e.registers[Y]
				if e.registers[X]&(1<<0) > 0 {
					//fmt.Printf(" 0xF bit was set!\n")
					e.registers[0xF] = 0x1
				} else {
					e.registers[0xF] = 0x0
				}
				e.registers[X] = e.registers[X] >> 1
			case 0x7: // Subtract: set register X to the result of registers Y - X.
				//fmt.Printf("%s: Subtract register Y-X reg-%d - reg-%d with values %d - %d\n", instruction, Y, X, e.registers[Y], e.registers[X])
				e.registers[X] = e.registers[Y] - e.registers[X]
				if e.registers[Y] > e.registers[X] {
					e.registers[0xF] = 0x1
				} else {
					e.registers[0xF] = 0x0
				}
			case 0xE: // Shift register X one step to the left
				e.registers[X] = e.registers[Y]
				// check if leftmost bit is set (and shifted out)
				//fmt.Printf("%s: Shift left. Was: %08b", instruction, e.registers[Y])
				if e.registers[X]&(1<<7) > 0 {
					//fmt.Printf(" 0xF bit was set!\n")
					e.registers[0xF] = 0x1
				} else {
					e.registers[0xF] = 0x0
				}
				e.registers[X] = e.registers[X] << 1
			default:
				panic("Unknown instruction: " + fmt.Sprintf("%02x", instr))
			}

		case 0x9: // Skip if values in registers X and Y are not equal
			if N == 0x0 && e.registers[X] != e.registers[Y] {
				e.pc += 2
			}
		case 0xA: // Set Index register to NNN
			e.I = NNN
		case 0xB: // May need to be made configurable!
			e.pc = NNN + uint16(e.registers[0x0]) // original behaviour, assume register 0x0.
		case 0xC: // Random number
			rnd := rand.Uint32()
			e.registers[X] = byte(rnd) & NN
		case 0xD: // draw sprite at I at screen x, y given by values in registers X and Y.
			xCoord := e.registers[X] % 64
			yCoord := e.registers[Y] % 32
			//fmt.Printf("%s: draw at x:%d y:%d\n", instruction, xCoord, yCoord)
			e.registers[0xF] = 0x0
			firstByteIndex := e.I
			numLines := int(N)
			for line := 0; line < numLines; line++ {
				spriteByte := e.memory[firstByteIndex]
				row := int(yCoord) + line
				if row > 31 {
					continue
				}

				for bit := 0; bit < 8; bit++ {

					col := int(xCoord) + bit
					// ignore if outside of screen
					if col > 63 {
						continue
					}

					// check if bit is set, moving from left-most bit to the right
					if spriteByte&(1<<(7-bit)) > 0 {
						if game.IsPixelSet(col, row) {
							game.Set(col, row, false)
							// set register F to 1
							e.registers[0xF] = 0x1
						} else {
							game.Set(col, row, true)
						}
					}
				}
				firstByteIndex++
			}
		case 0xE: // handle key presses EX9E and EXA1
			switch NN {
			case 0x9E: // pressed
				if ebiten.IsKeyPressed(runtime.ByteToKey(e.registers[X])) {
					e.pc += 2
				}
			case 0xA1: // not pressed
				if !ebiten.IsKeyPressed(runtime.ByteToKey(e.registers[X])) {
					e.pc += 2
				}
			default:
				panic("Unknown instruction: " + fmt.Sprintf("%02x", instr))
			}
		case 0xF: // timers
			switch NN {
			case 0x07: // Set register X to current value of delay timer
				e.delayTimerLock.Lock()
				e.registers[X] = e.delayTimer
				e.delayTimerLock.Unlock()
			case 0x15: // Set the delay timer to value of register X
				e.delayTimerLock.Lock()
				e.delayTimer = e.registers[X]
				e.delayTimerLock.Unlock()
			case 0x18: // Set the sound timer to value of register X
				e.soundTimerLock.Lock()
				e.soundTimer = e.registers[X]
				e.soundTimerLock.Unlock()
			case 0x1E: // Add to index: Add value of register X to I
				//fmt.Printf("%s: add register %d val %d to I\n", instruction, X, e.registers[X])
				i := e.I + uint16(e.registers[X])

				// old-school amiga behaviour
				if i > 0xFFF {
					e.registers[0xF] = 0x1
					i = i % 0x1000 // mod 4096 in case of overflow over original 4kb of RAM
				} else {
					e.registers[0xF] = 0x0
				}
				e.I = i
			case 0x0A: // Get key (blocks until input is received) TODO this one is not complete!!
				// input chan should be cleared beforehand?

				keypresses := make(chan byte)
				go game.WaitForKeypress(keypresses)
				key := <-keypresses

				e.registers[X] = key
				//pc = pc - 2 // decrease PC by 2... if the "wait for input" is cancelled by a timer...?
			case 0x29: // font character
				//fmt.Printf("%s: Set FONT char from register %d having value %d\n", instruction, X, e.registers[X])
				b := e.registers[X] & 0x0F // just use last nibble of value in register X
				e.I = uint16(fontOffsets[b])
			case 0x33: // binary-coded decimal conversion. Note that "10" is split into 0,1,0 and 4 into 0,0,4.
				e.memory[e.I+0] = (e.registers[X] / 100) % 10
				e.memory[e.I+1] = (e.registers[X] / 10) % 10
				e.memory[e.I+2] = (e.registers[X] / 1) % 10
			case 0x55: // Store register to memory
				for i := 0; i <= int(X); i++ {
					index := e.I + uint16(i)
					e.memory[index] = e.registers[i]
				}
				e.I = e.I + uint16(X+1)
			case 0x65: // Load value from memory into register
				//fmt.Printf("%s: Load value from memory into register 0-%d\n", instruction, X)
				for i := uint8(0); i <= X; i++ {
					e.registers[i] = e.memory[e.I]
					e.I = e.I + 1
				}
				//fmt.Println()
			default:
				panic("Unknown instruction: " + fmt.Sprintf("%02x", instr))
			}

		default:
			panic("Unknown instruction: " + fmt.Sprintf("%02x", instr))
		}
		time.Sleep(time.Microsecond * 1300) // corresponds to about 700 instructions per second...
	}
}

func (e *emulator) logger(instruction string, X, Y, N, NN byte, NNN uint16) {
	switch instruction {
	case "2NNN":
		fmt.Printf("%s: Push to stack. Old PC: %d, new stack frame: %d, NNN: %d\n", instruction, e.pc, e.stackFrame, NNN)
	case "6XNN":
		fmt.Printf("%s: SET: register %d to %d\n", instruction, X, NN)
	}
}
