package platform

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	errReplayDone = errors.New("replay done")
	errReplayStop = errors.New("replay stop")
)

type replay struct {
	cereal []byte
	input  anansi.InputReplay
	cur    anansi.InputReplay
	frame  anansi.InputFrame
	pause  time.Time
	size   image.Point
	mouse  struct {
		Mouse
		decay int
	}
}

func (rep *replay) update(ctx *Context) {
	if len(rep.cur) == 0 {
		ctx.Err = errReplayDone
		return
	}

	// Ctrl-C stops replay
	if ctx.Input.HasTerminal('\x03') {
		log.Printf("stopping replay")
		ctx.Platform.replay = nil
		return
	}

	defer func() {
		if ctx.Err == nil {
			rep.drawOverlay(ctx)
		}
	}()

	// <Space> (un)pauses replay
	if ctx.Input.CountRune(' ')%2 == 1 {
		if rep.pause.IsZero() {
			rep.pause = rep.frame.T
		} else {
			rep.pause = time.Time{}
		}
	}

	switch {
	// When paused '.' steps
	case !rep.pause.IsZero() && ctx.Input.CountRune('.')%2 == 1:
		rep.pause = rep.frame.T
		fallthrough

	case rep.pause.IsZero():
		// next non-message frame
		for rep.next() {
			switch m := rep.frame.M; {
			case bytes.HasPrefix(m, []byte("resize:")):
				if sz, err := parseSize(m[7:]); err != nil {
					log.Printf("invalid resize message %q in replay", m)
				} else {
					rep.size = sz
				}
			case len(m) > 0:
				log.Printf("unrecognized replay message %q", m)
			}
			if !rep.frame.T.IsZero() {
				break
			}
		}
		if rep.frame.T.IsZero() {
			ctx.Err = errReplayDone
			return
		}

		// load replay frame input
		ctx.events.Load(rep.frame.B)

		// update mouse cursor state
		if m, have := ctx.events.LastMouse(false); have {
			rep.mouse.Mouse = m
			rep.mouse.decay = 15
		} else if rep.mouse.decay > 0 {
			if rep.mouse.decay--; rep.mouse.decay == 0 {
				rep.mouse.State = ansi.MouseNoButton
			}
		}
	}

	ctx.Err = rep.runClient(ctx)
}

func (rep *replay) runClient(ctx *Context) error {
	ctx.Time = rep.frame.T
	ctx.Output.Resize(rep.size) // TODO better afford size difference
	err := rep.frame.E          // TODO wrap so that it can be non-fatal to looper?
	return errOr(err, ctx.runClient())
}

var buttonAttrs = []ansi.SGRAttr{
	ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRWhite.BG() | ansi.SGRWhite.FG(),   // none
	ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRWhite.FG() | ansi.SGRRed.BG(),     // left
	ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRWhite.FG() | ansi.SGRGreen.BG(),   // middle
	ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRWhite.FG() | ansi.SGRBlue.BG(),    // right
	ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRWhite.FG() | ansi.SGRMagenta.BG(), // wheel up
	ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRWhite.FG() | ansi.SGRCyan.BG(),    // wheel down
	ansi.SGRAttrClear | ansi.SGRAttrBold | ansi.SGRWhite.FG() | ansi.SGRYellow.BG(),  // inconceivable
}

func (rep *replay) drawOverlay(ctx *Context) {
	// TODO better status, maybe integrate with more general debug overlay

	ctx.Output.WriteESC(ansi.DECSC)

	if rep.mouse.Mouse != ZM {
		// TODO better mouse cursor drawing
		ctx.Output.Cell(rep.mouse.Point).Set('X', buttonAttrs[rep.mouse.State.ButtonID()])
	}

	// TODO OSD for keyboard events?

	ctx.Output.WriteESC(ansi.DECRC)
}

func (rep *replay) next() bool {
	if len(rep.cur) == 0 {
		return false
	}
	rep.frame = rep.cur[0]
	rep.cur = rep.cur[1:]
	return true
}

func (p *Platform) setRecording(f *os.File, err error) {
	p.events.input.SetRecording(nil)
	if p.recording != nil {
		if err := p.recording.Close(); err != nil {
			log.Printf("failed to close record file %q: %v", p.recording.Name(), err)
		}
		p.recording = nil
	}
	if f != nil {
		sw := sizedWriter{ws: f}
		if err == nil {
			err = p.writeState(&sw)
			if err == nil {
				err = sw.Finish()
				p.recording = f
				if err == nil {
					err = p.recordSize()
				}
			}
		}

		if err != nil {
			p.recording = nil
			_ = os.Remove(f.Name())
			_ = f.Close()
			log.Printf("failed to encode platform state (aborting recording): %v", err)
			return
		}

		p.events.input.SetRecording(f)
		log.Printf("recording input to %q", f.Name())
	}
}

func (p *Platform) loadReplay(f *os.File) error {
	p.setRecording(nil, nil)
	rep, err := readReplay(f)
	if err != nil {
		return err
	}
	p.replay = rep
	log.Printf("replaying %v frames over %v from %q",
		len(p.replay.input), p.replay.input.Duration(), readerName(f))
	return nil
}

func (p *Platform) toggleRecRep() error {
	if p.recording == nil {
		// TODO better filename selection
		p.setRecording(os.Create("auto.rec"))
		return nil
	}
	name := p.recording.Name()
	p.setRecording(nil, nil)

	f, err := os.Open(name)
	if err == nil {
		err = p.loadReplay(f)
		err = errOr(err, f.Close())
	}
	return errOr(err, errReplayDone)
}

func (p *Platform) recordSize() error {
	// APC "resize:" width "," height ST
	if p.recording != nil {
		if _, err := fmt.Fprintf(p.recording,
			"\x1b_resize:%d,%d\x1b\\",
			p.screen.Size.X, p.screen.Size.Y); err != nil {
			return fmt.Errorf("failed to record size: %v", err)
		}
	}
	return nil
}

func parseSize(b []byte) (pt image.Point, err error) {
	i := bytes.IndexByte(b, ',')
	if i < 0 {
		return image.ZP, errors.New("no ',' separator")
	}
	pt.X, _, err = ansi.DecodeNumber(b[:i])
	if err != nil {
		return image.ZP, err
	}
	pt.Y, _, err = ansi.DecodeNumber(b[i+1:])
	if err != nil {
		return image.ZP, err
	}
	return pt, err
}

func readReplay(f *os.File) (_ *replay, err error) {
	var rep replay
	if rep.cereal, err = readSized(f); err != nil {
		return nil, err
	}
	if rep.input, err = anansi.ReadInputReplay(f); err != nil {
		return nil, err
	}
	return &rep, nil
}

func readSized(r io.Reader) ([]byte, error) {
	var tmp [8]byte
	if _, err := r.Read(tmp[:]); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint64(tmp[:])
	return ioutil.ReadAll(io.LimitReader(r, int64(size)))
}

func readerName(r io.Reader) string {
	type named interface{ Name() string }
	if nd, ok := r.(named); ok {
		return nd.Name()
	}
	return "<unknown>"
}

type sizedWriter struct {
	started bool
	off     int64
	size    uint64
	ws      io.WriteSeeker
}

func (sw *sizedWriter) Write(p []byte) (n int, err error) {
	if !sw.started {
		off, err := sw.ws.Seek(0, io.SeekCurrent)
		if err == nil {
			_, err = sw.ws.Seek(8, io.SeekCurrent)
		}
		if err != nil {
			return 0, err
		}
		sw.off = off
		sw.started = true
	}
	n, err = sw.ws.Write(p)
	sw.size += uint64(n)
	return n, err
}

func (sw *sizedWriter) Finish() error {
	if !sw.started {
		return nil
	}
	off, err := sw.ws.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if _, err := sw.ws.Seek(sw.off, io.SeekStart); err != nil {
		return err
	}
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], sw.size)
	if _, err = sw.ws.Write(tmp[:]); err != nil {
		return err
	}
	if _, err := sw.ws.Seek(off, io.SeekStart); err != nil {
		return err
	}
	sw.off += int64(sw.size)
	sw.size = 0
	sw.started = false
	return nil
}
