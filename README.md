# AnANSI - a bag of collected wisdom for manipulating terminals

... or yet ANother ANSI terminal library.

## Why?

- Designed to be a loosely coupled set of principled layers, rather than (just)
  one unified convenient interface.
- Be more Go-idiomatic / natural: e.g.  [ansi.DecodeEscape][ansi_decode_escape]
  following [utf8.DecodeRune][decode_rune] convention, rather than heavier
  weight event parsing/handling.
- Supporting use cases other than fullscreen raw mode.
- Allow applications to choose input modality, rather than lock-in to one
  paradigm like non-blocking/SIGIO.
- Support implementing terminal emulators, e.g. to build a multiplexer or debug
  wrapper.

## Status

**Prototyping/Experimental**: AnANSI is currently in initial exploration mode,
and while things on master are reasonably stable, there's no guarantees yet.
That said, there is a working demo command on the [dev][dev] branch.

### Done

Toplevel `anansi` package:
- [anansi.Term][anansi_term], [anansi.Context][anansi_context], and
  [anansi.Attr][anansi_attr] provide cohesive management of terminal state such
  as raw mode, ANSI escape sequenced modes, and SGR attribute state.

Core `anansi/ansi` package:
- [ansi.DecodeEscape][ansi_decode_escape] provides escape sequence decoding
  as similarly to [utf8.DecodeRune][decode_rune] as possible. Additional
  support for decoding escape arguments is provided (`DecodeNumber`,
  `DecodeSGR`, `DecodeMode`, and `DecodeCursorCardinal`).
- [ansi.SGRAttr][ansi_sgr] supports dealing with terminal colors and text
  attributes.
- [ansi.MouseState][ansi_mousestate] supports handling xterm extended mouse
  reporting.
- function definitions like [ansi.CUP][ansi_cup] and [ansi.SM][ansi_sm] for
  building [control sequences][ansi_seq]
  terminal state management
- [ansi.Mode][ansi_mode] supports setting and clearing various modes such as
  mouse reporting (and its optional extra levels like motion and full button
  reporting).

### WIP

- buffered ansi processing with cursor state tracking ([rc][rc])
- input buffer ([rc][rc])
- output buffer ([rc][rc])
- cursor state tracking ([rc][rc])
- screen grid ([rc][rc])
- screen state tracking and differential update ([rc][rc])
- animation (tick) control loop ([dev][dev])
- a 60fps [demo][demo] with things like:
  - experimenting with the immediate mode user concept
  - an input event processing queue
  - a cursor state construct
  - a diagnostic HUD that displays things like Go's log output, frame timing
    data, and mouse state
- special decoding for CSI M, whose arg follows AFTER

### TODO

- a signal processing layer
- a cursor state piece (e.g. to support immediate mode API)
- a screen grid box-of-state (e.g. to support things like back/front buffer
  diffing and other tricks)
- maybe event synthesis from signals and input
- maybe a high level client api that gets events and an output context
- provide `DecodeEscapeInString(s string)` for completeness
- terminfo layer:
  - automated codegen (for builtins)
  - full load rather than the termbox-inherited cherry picking

### Branches

AnANSI uses a triple branch (`master`, `rc`, and `dev`) pattern that I've found
useful:
- the [master branch][master] has relatively stable code but is
  still pre `v1.0.0`, and so is not *actually* stable; tests must pass on all
  commits
- the [rc branch][rc] contains code that is stable-ish: tests should
  pass on all commits
- the [dev branch][dev] contains the sum of all hopes/fears, tests
  may not pass

## Resources

- [xterm control sequences][xterm_ctl]
- [vt100.net][vt100],
  - especially its [dec ansi parser][ansi_parser_sm] state diagram
- [UCS history][ucs] and the [unicode BMP][unicode_bmp] of course
- ansicode.txt [source1][tmux_ansicode] [source2][pdp10_ansicode]
- antirez did a great [raw mode teardown][kilo_rawmode] for kilo [kilo][kilo]
- more history collation:
  - https://www.cl.cam.ac.uk/~mgk25/unicode.html
  - https://www.dabsoft.ch/dicom/3/C.12.1.1.2/
- various related Go libraries like:
  - the ill-fated [x/term](https://github.com/golang/go/issues/13104) package
  - [termbox][termbox]
  - [tcell][tcell]
  - [cops][cops]
  - [go-ansiterm][go-ansiterm]
  - [terminfo][terminfo]

[anansi_attr]: https://godoc.org/github.com/jcorbin/anansi#Attr
[anansi_context]: https://godoc.org/github.com/jcorbin/anansi#Context
[anansi_term]: https://godoc.org/github.com/jcorbin/anansi#Term
[ansi_cup]: https://godoc.org/github.com/jcorbin/anansi/ansi#CUP
[ansi_decode_escape]: https://godoc.org/github.com/jcorbin/anansi/ansi#DecodeEscape
[ansi_mode]: https://godoc.org/github.com/jcorbin/anansi/ansi#Mode
[ansi_mousestate]: https://godoc.org/github.com/jcorbin/anansi/ansi#MouseState
[ansi_parser_sm]: https://www.vt100.net/emu/dec_ansi_parser
[ansi_seq]: https://godoc.org/github.com/jcorbin/anansi/ansi#Seq
[ansi_sgr]: https://godoc.org/github.com/jcorbin/anansi/ansi#SGRAttr
[ansi_sm]: https://godoc.org/github.com/jcorbin/anansi/ansi#SM

[cops]: https://github.com/kriskowal/cops
[decode_rune]: https://golang.org/pkg/unicode/utf8/#DecodeRune
[go-ansiterm]: https://github.com/Azure/go-ansiterm
[kilo]: https://github.com/antirez/kilo
[kilo_rawmode]: https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html
[pdp10_ansicode]: http://www.inwap.com/pdp10/ansicode.txt
[tcell]: https://github.com/gdamore/tcell
[termbox]: https://github.com/nsf/termbox-go
[terminfo]: https://github.com/xo/terminfo
[tmux_ansicode]: https://github.com/tmux/tmux/blob/master/tools/ansicode.txt
[ucs]: https://en.wikipedia.org/wiki/Universal_Coded_Character_Set
[unicode_bmp]: https://en.wikipedia.org/wiki/Plane_(Unicode)#Basic_Multilingual_Plane
[vt100]: https://www.vt100.net
[xterm_ctl]: http://invisible-island.net/xterm/ctlseqs/ctlseqs.html

[master]: ../../tree/master
[rc]: ../../tree/rc
[dev]: ../../tree/dev
[demo]: ../../tree/dev/cmd/demo
