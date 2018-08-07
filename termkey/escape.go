package termkey

import (
	"bytes"

	"github.com/jcorbin/anansi/terminfo"
)

type escapeAutomaton struct {
	term [256]terminfo.KeyCode
	next [256]*escapeAutomaton
}

func (ea *escapeAutomaton) addChain(bs []byte, kc terminfo.KeyCode) {
	for len(bs) > 1 {
		b := bs[0]
		next := ea.next[b]
		if next == nil {
			next = &escapeAutomaton{}
			ea.next[b] = next
		}
		ea = next
		bs = bs[1:]
	}
	b := bs[0]
	ea.term[b] = kc
}

func (ea *escapeAutomaton) lookup(bs []byte) (_ terminfo.KeyCode, n int) {
	for ea != nil && len(bs) > 1 {
		b := bs[0]
		if kc := ea.term[b]; kc != 0 {
			return kc, n + 1
		}
		ea = ea.next[b]
	}
	return 0, 0
}

func (ea *escapeAutomaton) decode(buf []byte) (ev Event, n int) {
	if kc, m := ea.lookup(buf); kc != 0 {
		ev.Key, n = Key(0x80|kc), m
	} else if bytes.HasPrefix(buf, []byte("\x1b[")) {
		ev, n = decodeMouseEvent(buf)
	}
	if n == 0 && buf[0] == 0x1b {
		ev.Key, n = KeyEsc, 1
	}
	return ev, n
}
