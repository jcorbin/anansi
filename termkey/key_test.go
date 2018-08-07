package termkey_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/jcorbin/anansi/termkey"
)

func TestKey_String(t *testing.T) {
	var specials = []string{
		"<F1>",
		"<F2>",
		"<F3>",
		"<F4>",
		"<F5>",
		"<F6>",
		"<F7>",
		"<F8>",
		"<F9>",
		"<F10>",
		"<F11>",
		"<F12>",
		"<Insert>",
		"<Delete>",
		"<Home>",
		"<End>",
		"<PageUp>",
		"<PageDown>",
		"<Up>",
		"<Down>",
		"<Left>",
		"<Right>",
		"<MouseLeft>",
		"<MouseMiddle>",
		"<MouseRight>",
		"<MouseRelease>",
		"<MouseWheelUp>",
		"<MouseWheelDown>",
	}

	for k := Key(0); k < 0xff; k++ {
		t.Run(fmt.Sprintf("Key(0x%02x)", uint8(k)), func(t *testing.T) {
			switch {
			case k < 0x20 || k == 0x7f:
				assert.Equal(t, string([]byte{'^', byte(k) ^ 0x40}), k.String())
			case k < 0x80:
				assert.Equal(t, string(byte(k)), k.String())
			case k.IsSpecial():
				assert.Equal(t, specials[k&0x7f], k.String())
			default:
				assert.Equal(t, fmt.Sprintf("Key<%02x>", uint8(k)), k.String())
			}
		})
	}
}
