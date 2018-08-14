# AnANSI - a bag of collected wisdom for manipulating terminals

... or yet ANother ANSI terminal library.

## Why?

- Designed to be a loosely coupled set of principled layers, rather than (just)
  one unified convenient interface.
- Be more Go-idiomatic / natural.
- Supporting use cases other than fullscreen raw mode.
- Allow applications to choose input modality, rather than lock-in to one
  paradigm like non-blocking/SIGIO.
- Support implementing terminal emulators, e.g. to build a multiplexer or debug
  wrapper.

## Status

**Prototyping/Experimental**: AnANSI is currently in initial exploration mode,
using a triple branch (`master`, `rc`, and `dev`) pattern that I've found
useful:
- the [master branch](../../tree/master) has relatively stable code but is
  still pre `v1.0.0`, and so is not *actually* stable; tests must pass on all
  commits
- the [rc branch](../../tree/rc) contains code that is stable-ish: tests should
  pass on all commits
- the [dev branch](../../tree/dev) contains the sum of all hopes/fears, tests
  may not pass

## TODO

- unwritten:
  - a termios layer
  - an ANSI layer
  - an input buffer
  - an output buffer
  - a signal processing layer
  - a cursor state piece (e.g. to support immediate mode API)
  - a screen grid box-of-state (e.g. to support things like back/front buffer
    diffing and other tricks)
  - maybe event synthesis from signals and input
  - maybe a high level client api that gets events and an output context
- terminfo layer:
  - automated codegen (for builtins)
  - full load rather than the termbox-inherited cherry picking

## Resources

- [xterm control sequences](http://invisible-island.net/xterm/ctlseqs/ctlseqs.html)
- [vt100.net](https://www.vt100.net)
  - especially its [dec ansi parser](https://www.vt100.net/emu/dec_ansi_parser) state diagram
- [UCS history][ucs] and the [unicode BMP][unicode_bmp] of course
- ansicode.txt [source1](https://github.com/tmux/tmux/blob/master/tools/ansicode.txt) [source2](http://www.inwap.com/pdp10/ansicode.txt)
- more history collation:
  - https://www.cl.cam.ac.uk/~mgk25/unicode.html
  - https://www.dabsoft.ch/dicom/3/C.12.1.1.2/
- various related Go libraries like:
  - the ill-fated [x/term](https://github.com/golang/go/issues/13104) package
  - [termbox](https://github.com/nsf/termbox-go)
  - [tcell](https://github.com/gdamore/tcell)
  - [cops](https://github.com/kriskowal/cops)
  - [go-ansiterm](https://github.com/Azure/go-ansiterm)
  - [terminfo](https://github.com/xo/terminfo)

[ucs]: https://en.wikipedia.org/wiki/Universal_Coded_Character_Set
[unicode_bmp]: https://en.wikipedia.org/wiki/Plane_(Unicode)#Basic_Multilingual_Plane
