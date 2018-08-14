package ansi_test

import (
	"sort"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/anansi/terminfo"
)

func Test_Terminfo_Integration(t *testing.T) {
	for _, term := range []string{"xterm", "screen", "linux"} {
		t.Run(term, func(t *testing.T) {
			info, err := terminfo.GetBuiltin(term)
			require.NoError(t, err, "unable to get builtin terminfo")
			m := info.FuncMap()
			for _, key := range sortedKeys(m) {
				t.Run(key, func(t *testing.T) {
					p := []byte(m[key])
					offset := 0
					for len(p) > 0 {
						e, a, n := ansi.DecodeEscape(p)
						if n > 0 {
							t.Logf("%v %q %v", e, a, n)
							p = p[n:]
							offset += n
							continue
						}
						r, n := utf8.DecodeRune(p)
						if ansi.Escape(r).IsCharacterSetControl() {
							t.Logf("character set control %U", r)
							p = p[n:]
							offset += n
							continue
						}
						t.Logf("failed to decode escape @%v in %q (next rune: %U)", offset, m[key], r)
						t.FailNow()
					}
				})
			}
		})
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
