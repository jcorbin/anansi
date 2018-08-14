package terminfo

import (
	"encoding/hex"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type compatEntry struct {
	partial string
	*Terminfo
}

var (
	cache       = make(map[string]*Terminfo, 64)
	builtins    = make(map[string]*Terminfo, 64) // TODO more coverage
	compatTable = make([]compatEntry, 0, 64)
)

// KeyCode indexes Terminfo.Keys
type KeyCode uint8

// FuncCode indexes Terminfo.Funcs
type FuncCode uint8

// These constants provide convenient aliases for accessing function strings.
const (
	FuncEnterCA FuncCode = iota + 1
	FuncExitCA
	FuncShowCursor
	FuncHideCursor
	FuncClearScreen
	FuncSGR0
	FuncUnderline
	FuncBold
	FuncBlink
	FuncReverse
	FuncEnterKeypad
	FuncExitKeypad
	FuncEnterMouse
	FuncExitMouse

	maxFuncs
)

// These constants provide convenient aliases for accessing key strings.
const (
	KeyF1 KeyCode = iota + 1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyInsert
	KeyDelete
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyUp
	KeyDown
	KeyLeft
	KeyRight

	maxKeys
)

// Terminfo describes how to interact with a terminal.
type Terminfo struct {
	Name  string
	Keys  [maxKeys]string
	Funcs [maxFuncs]string
}

const (
	defaultTerm = "" // TODO provide a vt100 default?
	defaultPath = "/usr/share/terminfo"
)

// SearchPath returns candidate file paths to try to load from; this follows
// the behaviour described in terminfo(5) as distributed by ncurses.
func SearchPath() []string {
	if terminfo := os.Getenv("TERMINFO"); terminfo != "" {
		// if TERMINFO is set, no other directory should be searched
		return []string{terminfo}
	}

	dirs := os.Getenv("TERMINFO_DIRS")

	n := 2
	if dirs != "" {
		n += strings.Count(dirs, ":")
	}
	candidates := make([]string, 0, n)

	// ~/.terminfo comes first
	if u, err := user.Current(); err == nil {
		candidates = append(candidates, filepath.Join(u.HomeDir, ".terminfo"))
	}

	defaultAdded := false

	// then any TERMINFO_DIRS
	if dirs != "" {
		for _, dir := range strings.Split(dirs, ":") {
			if dir == "" { // "" -> "/usr/share/terminfo"
				candidates = append(candidates, defaultPath)
				defaultAdded = true
			} else {
				candidates = append(candidates, dir)
			}
		}
	}

	// fall back to default if not attepmted by an empty component in
	// TERMINFO_DIRS
	if !defaultAdded {
		candidates = append(candidates, defaultPath)
	}

	return candidates
}

// Load Terminfo for the given terminal name. Unless you know better, you want
// to pass the value of os.Getenv("TERM")`. No default is currently provided,
// so an error will be returned for the empty string. Loaded Terminfo is cached
// on success, and the same value returned next time. Some builtins are
// provided for common terminals, and will be used if reading from the terminfo
// database fails.
func Load(term string) (*Terminfo, error) {
	if ti, def := cache[term]; def {
		return ti, nil
	}
	paths := SearchPath()
	for _, path := range paths {
		for _, fp := range []string{
			filepath.Join(path, term[0:1], term),                            // the typical *nix path
			filepath.Join(path, hex.EncodeToString([]byte(term[:1])), term), // darwin specific dirs structure
		} {
			if f, err := os.Open(fp); err == nil {
				var ti Terminfo
				if err == nil {
					err = ti.ReadFrom(f)
					if cerr := f.Close(); err == nil {
						err = cerr
					}
				}
				if err == nil {
					cache[term] = &ti
				}
				return &ti, err
			}
		}
	}
	return GetBuiltin(term)
}
