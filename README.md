# AnANSI - a bag of collected wisdom for manipulating terminals

... or yet ANother ANSI terminal library.

## Why?

- Designed to be a loosely coupled set of principled layers, rather than (just)
  one unified convenient interface
- Be more Go-idiomatic / natural: e.g.  [ansi.DecodeEscape][ansi_decode_escape]
  following [utf8.DecodeRune][decode_rune] convention, rather than heavier
  weight event parsing/handling
- Supporting use cases other than fullscreen raw mode
- Allow applications to choose input modality, rather than lock-in to one
  paradigm like non-blocking/SIGIO
- Support implementing terminal emulators, e.g. to build a multiplexer or debug
  wrapper

## Status

**Experimental**: AnANSI is currently in initial exploration mode, and while
things on master are reasonably stable, there's no guarantees yet.

### Demos

The [Decode Demo Command][decode_demo] that demonstrates ansi Decoding,
optionally with mouse reporting and terminal state manipulation (for raw and
alternate screen mode).

There's also a [lolwut][lolwut] demo, which is a port of antirez's,
demonstrating braille-bitmap rendering capability. It has an optional
interactive animated mode of which demonstrates the experimental `x/platform`
layer.

There's another `x/platform` [demo][demo] that draws a colorful test pattern.

### Done

Experimental cohesive [`x/platform`][platform_pkg] layer:
- provides a `platform.Events` queue layered on top of `anansi.input`, which
  contains parsed `rune`, `ansi.Escape`, and `ansi.MouseState` data
- synthesizes all of the below `anansi` pieces (`Term`, `Input`, `Output`, etc)
  into one cohesive `platform.Context` which supports a single combined round
  of non-blocking input processing and output generation
- provides signal handling for typical things like `SIGINT`, `SIGERM`,
  `SIGHUP`, and `SIGWINCH`
- drives a `platform.Client` in a `platform.Tick` loop at a desired
  Frames-Per-Second (FPS) rate
- provides input record and replay on top of (de)serialized client and platform
  state
- supports inter-frame background work
- provides a diagnostic HUD overlay that displays things like Go's `log`
  output, FPS, time, mouse state, screen size, etc

Toplevel [`anansi`][anansi_pkg] package:
- [`anansi.Term`][anansi_term], [`anansi.Context`][anansi_context], and
  [`anansi.Attr`][anansi_attr] provide cohesive management of terminal state
  such as raw mode, ANSI escape sequenced modes, and SGR attribute state
- [`anansi.Input`][anansi_input] supports reading input from a file handle,
  implementing both blocking `.ReadMore()` and non-blocking `.ReadAny()` modes
- [`anansi.Output`][anansi_output] mediates flushing output from any
  `io.WriterTo` (implemented by both `anansi.Cursor` and `anansi.Screen`) into
  a file handle.  It properly handles non-blocking IO (by temporarily doing a
  blocking write if necessary) to coexist with `anansi.Input` (since `stdin`
  and `stdout` share the same underlying file descriptor)
- [`anansi.Cursor`][anansi_cursor] represents cursor state including position,
  visibility, and SGR attribute(s); it supports processing under an
  [`ansi.Buffer`][ansi_buffer]
- [`anansi/ansi.Point`][anansi_point] and
  [`anansi/ansi.Rectangle`][anansi_rectangle] support sane handling of
  1,1-originated screen geometry
- [`anansi.Grid`][anansi_grid] provides a 2d array of `rune` and`ansi.SGRAttr`
  data; it supports processing under an [`ansi.Buffer`][ansi_buffer].
- [`anansi.Screen`][anansi_screen] combines an `anansi.Cursor` with
  `anansi.Grid`, supporting differential screen updates and final post-update
  cursor display
- [`anansi.Bitmap`][anansi_bitmap] provides a 2d bitmap that can be rendered or
  drawn into braille runes.
- Both `anansi.Grid` and `anansi.Bitmap` support `anansi.Style`d
  [render][anansi_render_grid]ing into an `ansi.Buffer`, or
  [draw][anansi_draw_grid]ing into an (other) `anansi.Grid`.

Core [`anansi/ansi`][ansi_pkg] package:
- [`ansi.DecodeEscape`][ansi_decode_escape] provides escape sequence decoding
  as similarly to [`utf8.DecodeRune`][decode_rune] as possible. Additional
  support for decoding escape arguments is provided (`DecodeNumber`,
  `DecodeSGR`, `DecodeMode`, and `DecodeCursorCardinal`)
- [`ansi.SGRAttr`][ansi_sgr] supports dealing with terminal colors and text
  attributes
- [`ansi.MouseState`][ansi_mousestate] supports handling xterm extended mouse
  reporting
- function definitions like [`ansi.CUP`][ansi_cup] and [`ansi.SM`][ansi_sm] for
  building [`control sequences`][ansi_seq] terminal state management
- [`ansi.Mode`][ansi_mode] supports setting and clearing various modes such as
  mouse reporting (and its optional extra levels like motion and full button
  reporting)
- [`ansi.Buffer`][ansi_buffer] supports deferred writing to a terminal; the
  primary trick that it adds beyond a basic `bytes.Buffer` convenience, is
  allowing the users to process escape sequences, no matter how they're
  written. This enables keeping virtual state (such as cursor position or a
  cell grid) up to date without locking downstream users into specific APIs for
  writing

### Errata

- differential screen update is still not perfect, although the glitches that
  were previously present are now lessened due to the functional test; however
  this was done by removing a (perhaps premature) cursor movement optimization
  to simplify diffing
- Works For Me ™ in tmux-under-iTerm2: should also work in other modern
  xterm-descended terminals, such as the libvte family; however terminfo
  detection not yet used by the platform layer, so basic things like
  smcup/rmcup inversion may by broken
- `anansi.Screen` doesn't (yet) implement full vt100 emulation, notably lacking
  is scrolling region support
- there's something glitchy with trying to write into the final cell (last
  column of last row), sometimes it seems to trigger a scroll (as when used by
  hud log view) sometimes not (as when background filled by demo)

### WIP

- an [interact command demo](../../tree/dev/cmd/interact/main.go) which
  allows you to interactively manipulate arguments passed to a dynamically
  executed command

### TODO

- platform "middleware", i.e. for re-usable Ctrl-C and Ctrl-Z behavior (ideally
  making current builtins like Ctrl-L and record/replay pluggable)
- fancier image rendition (e.g. leveraging iTerm2's image support)
- special decoding for CSI M, whose arg follows AFTER
- provide `DecodeEscapeInString(s string)` for completeness
- support bracketed paste mode (and decoding pastes from it)
- consider compacting the record file format; maybe also compression it
- terminfo layer:
  - automated codegen (for builtins)
  - full load rather than the termbox-inherited cherry picking
- terminal interrogation:
  - where's the cursor?
  - CSI DA
  - CSI DSR

### Branches

AnANSI uses a triple branch (`master`, `rc`, and `dev`) pattern that I've found
useful:
- the [master branch][master] has relatively stable code but is
  still pre `v1.0.0`, and so is not *actually* stable; tests must pass on all
  commits. NOTE any package under `anansi/x/` doesn't even have the tacit
  attempt made at stability that the rest of `anansi/` on master does.
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

[platform_pkg]: https://godoc.org/github.com/jcorbin/anansi/x/platform
[anansi_pkg]: https://godoc.org/github.com/jcorbin/anansi
[ansi_pkg]: https://godoc.org/github.com/jcorbin/anansi/ansi

[anansi_attr]: https://godoc.org/github.com/jcorbin/anansi#Attr
[anansi_bitmap]: https://godoc.org/github.com/jcorbin/anansi#Bitmap
[anansi_context]: https://godoc.org/github.com/jcorbin/anansi#Context
[anansi_cursor]: https://godoc.org/github.com/jcorbin/anansi#Cursor
[anansi_draw_grid]: https://godoc.org/github.com/jcorbin/anansi#DrawGrid
[anansi_grid]: https://godoc.org/github.com/jcorbin/anansi#Grid
[anansi_input]: https://godoc.org/github.com/jcorbin/anansi#Input
[anansi_output]: https://godoc.org/github.com/jcorbin/anansi#Output
[anansi_point]: https://godoc.org/github.com/jcorbin/anansi#Point
[anansi_rectangle]: https://godoc.org/github.com/jcorbin/anansi#Rectangle
[anansi_render_grid]: https://godoc.org/github.com/jcorbin/anansi#RenderGrid
[anansi_screen]: https://godoc.org/github.com/jcorbin/anansi#Screen
[anansi_term]: https://godoc.org/github.com/jcorbin/anansi#Term
[ansi_buffer]: https://godoc.org/github.com/jcorbin/anansi/ansi#Buffer
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

[demo]: ../../tree/master/cmd/demo
[lolwut]: ../../tree/master/cmd/lolwut/main.go
[decode_demo]: ../../tree/master/cmd/decode/main.go
[master]: ../../tree/master
[rc]: ../../tree/rc
[dev]: ../../tree/dev
