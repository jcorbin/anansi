package terminfo

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	tiMouseEnter = "\x1b[?1000h\x1b[?1002h\x1b[?1015h\x1b[?1006h"
	tiMouseLeave = "\x1b[?1006l\x1b[?1015l\x1b[?1002l\x1b[?1000l"

	// "Maps" the Key* constants from terminfo.go to the number of the
	// respective string capability in the terminfo file.
	tiKeys = [maxKeys]uint16{
		0,   // XXX invalid
		66,  // KeyF1
		68,  // KeyF2 NOTE not a typo; 67 is F10
		69,  // KeyF3
		70,  // KeyF4
		71,  // KeyF5
		72,  // KeyF6
		73,  // KeyF7
		74,  // KeyF8
		75,  // KeyF9
		67,  // KeyF10
		216, // KeyF11
		217, // KeyF12
		77,  // KeyInsert
		59,  // KeyDelete
		76,  // KeyHome
		164, // KeyEnd
		82,  // KeyPgup
		81,  // KeyPgdn
		87,  // KeyArrowUp
		61,  // KeyArrowDown
		79,  // KeyArrowLeft
		83,  // KeyArrowRight
	}

	// "Maps" the Func* constants from terminfo.go to the number of the
	// respective string capability in the terminfo file.
	tiFuncs = [maxFuncs - 2]uint16{
		0,  // XXX invalid
		28, // FuncEnterCA
		40, // FuncExitCA
		16, // FuncShowCursor
		13, // FuncHideCursor
		5,  // FuncClearScreen
		39, // FuncSGR0
		36, // FuncUnderline
		27, // FuncBold
		26, // FuncBlink
		34, // FuncReverse
		89, // FuncEnterKeypad
		88, // FuncExitKeypad
	}
)

// ReadFrom reads terminfo data from the given io.ReadSeeker, returning any
// read error.
//
// TODO should we return an `nRead int` too so that TermInfo implements io.ReaderFrom?
func (ti *Terminfo) ReadFrom(rs io.ReadSeeker) error {
	const (
		magic        = 0432
		headerLength = 12
	)

	// 0: magic number
	// 1: size of names section
	// 2: size of boolean section
	// 3: size of numbers section (in integers)
	// 4: size of the strings section (in integers)
	// 5: size of the string table

	var header [6]int16

	if err := binary.Read(rs, binary.LittleEndian, header[:]); err != nil {
		return err
	}

	if header[0] != magic {
		return fmt.Errorf("invalid magic number %07o", header[0])
	}

	if (header[1]+header[2])%2 != 0 {
		// old quirk to align everything on word boundaries
		header[2]++
	}

	strOffset := headerLength + uint16(header[1]+header[2]+2*header[3])
	tableOffset := strOffset + 2*uint16(header[4])

	for i := 1; i < len(tiKeys); i++ {
		key, err := readTableString(rs, strOffset+2*tiKeys[i], tableOffset)
		if err != nil {
			return err
		}
		ti.Keys[i] = key
	}

	for i := 1; i < len(tiFuncs); i++ {
		fnc, err := readTableString(rs, strOffset+2*tiFuncs[i], tableOffset)
		if err != nil {
			return err
		}
		ti.Funcs[i] = fnc
	}

	ti.Funcs[FuncEnterMouse] = tiMouseEnter
	ti.Funcs[FuncExitMouse] = tiMouseLeave

	return nil
}

func readTableString(rs io.ReadSeeker, off, table uint16) (s string, err error) {
	index, err := readUint16(rs, off)
	if err != nil {
		return "", err
	}
	if _, err := rs.Seek(int64(table+index), 0); err != nil {
		return "", err
	}
	return readNullString(rs)
}

func readUint16(rs io.ReadSeeker, off uint16) (uint16, error) {
	if _, err := rs.Seek(int64(off), 0); err != nil {
		return 0, err
	}
	if err := binary.Read(rs, binary.LittleEndian, &off); err != nil {
		return 0, err
	}
	return off, nil
}

func readNullString(r io.Reader) (s string, err error) {
	var bs []byte
	var buf [8]byte
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return "", err
		}
		for i, b := range buf[:n] {
			if b == 0 {
				bs = append(bs, buf[:i]...)
				return string(bs), nil
			}
		}
		bs = append(bs, buf[:n]...)
	}
}
