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

// FuncMap builds and returns a string-string map of the control sequence
// functions defined for callers that are more interested in flexibility than
// performance.
func (info Terminfo) FuncMap() map[string]string {
	r := make(map[string]string, maxFuncs)
	r["EnterCA"] = info.Funcs[FuncEnterCA]
	r["ExitCA"] = info.Funcs[FuncExitCA]
	r["ShowCursor"] = info.Funcs[FuncShowCursor]
	r["HideCursor"] = info.Funcs[FuncHideCursor]
	r["ClearScreen"] = info.Funcs[FuncClearScreen]
	r["SGR0"] = info.Funcs[FuncSGR0]
	r["Underline"] = info.Funcs[FuncUnderline]
	r["Bold"] = info.Funcs[FuncBold]
	r["Blink"] = info.Funcs[FuncBlink]
	r["Reverse"] = info.Funcs[FuncReverse]
	r["EnterKeypad"] = info.Funcs[FuncEnterKeypad]
	r["ExitKeypad"] = info.Funcs[FuncExitKeypad]
	r["EnterMouse"] = info.Funcs[FuncEnterMouse]
	r["ExitMouse"] = info.Funcs[FuncExitMouse]
	return r
}

// KeyMap builds and returns a string-string map of the key control sequences
// defined for callers that are more interested in flexibility than
// performance.
func (info Terminfo) KeyMap() map[string]string {
	r := make(map[string]string, maxKeys)
	r["F1"] = info.Keys[KeyF1]
	r["F2"] = info.Keys[KeyF2]
	r["F3"] = info.Keys[KeyF3]
	r["F4"] = info.Keys[KeyF4]
	r["F5"] = info.Keys[KeyF5]
	r["F6"] = info.Keys[KeyF6]
	r["F7"] = info.Keys[KeyF7]
	r["F8"] = info.Keys[KeyF8]
	r["F9"] = info.Keys[KeyF9]
	r["F10"] = info.Keys[KeyF10]
	r["F11"] = info.Keys[KeyF11]
	r["F12"] = info.Keys[KeyF12]
	r["Insert"] = info.Keys[KeyInsert]
	r["Delete"] = info.Keys[KeyDelete]
	r["Home"] = info.Keys[KeyHome]
	r["End"] = info.Keys[KeyEnd]
	r["PageUp"] = info.Keys[KeyPageUp]
	r["PageDown"] = info.Keys[KeyPageDown]
	r["Up"] = info.Keys[KeyUp]
	r["Down"] = info.Keys[KeyDown]
	r["Left"] = info.Keys[KeyLeft]
	r["Right"] = info.Keys[KeyRight]
	return r
}
