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

- more contextual links
- a terminfo layer
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
