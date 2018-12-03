package main

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/anansi/x/platform"
)

var errInt = errors.New("interrupt")

func main() {
	platform.MustRun(os.Stdin, os.Stdout, Run, platform.FrameRate(60))
}

// Run the demo under an active terminal platform.
func Run(p *platform.Platform) error {
	for {
		var d demo
		if err := p.Run(&d); platform.IsReplayDone(err) {
			continue // loop replay
		} else if err == io.EOF || err == errInt {
			return nil
		} else if err != nil {
			log.Printf("exiting due to %v", err)
			return err
		}
	}
}

type demo struct {
	Grid anansi.Grid
}

func (d *demo) Update(ctx *platform.Context) (err error) {
	// Ctrl-C interrupts
	if ctx.Input.HasTerminal('\x03') {
		// ... AFTER any other available input has been processed
		err = errInt
		// ... NOTE err != nil will prevent wasting any time flushing the final
		//          lame-duck frame
	}

	// Ctrl-Z suspends
	if ctx.Input.CountRune('\x1a') > 0 {
		defer func() {
			if err == nil {
				err = ctx.Suspend()
			} // else NOTE don't bother suspending, e.g. if Ctrl-C was also present
		}()
	}

	// log any unhandled escape sequences
	defer func() {
		if err == nil {
			for id, kind := range ctx.Input.Type {
				if kind == platform.EventEscape {
					log.Printf("unhandled escape sequence %v", ctx.Input.Escape(id))
				}
			}
		}
	}()

	ctx.Output.Clear()

	// TODO a readline utility

	d.Grid.Resize(ctx.Output.Bounds().Size())

	sweep := 10*ctx.Time.Second() + ctx.Time.Nanosecond()/int(time.Millisecond)/100
	r := ctx.Output.Bounds()
	for p := r.Min; p.Y < r.Max.Y; p.Y++ {
		line := p.Y%16 == 0
		i, _ := d.Grid.CellOffset(p)
		for p.X = r.Min.X; p.X < r.Max.X; p.X++ {
			a := ansi.RGB(0, uint8(p.X), uint8(p.Y)).BG()
			var r rune
			if line {
				a |= ansi.RGB(uint8(sweep), 0, 0).BG()
				r = runeSweep[sweep%len(runeSweep)]
				sweep++
			}
			d.Grid.Attr[i] = a
			d.Grid.Rune[i] = r
			i++
		}
	}

	// TODO Grid method
	// model := ansi.ColorModelID
	// at := image.Pt(1, 1)
	def := ' '
	r = d.Grid.Bounds()
	p := r.Min
	for i, ch := range d.Grid.Rune {
		if ch == 0 {
			ch = def
		}
		if j, ok := ctx.Output.CellOffset(p); ok {
			ctx.Output.Grid.Rune[j] = ch
			ctx.Output.Grid.Attr[j] = d.Grid.Attr[i]
		}
		if p.X++; p.X >= r.Max.X {
			p.X = r.Min.X
			p.Y++
		}
	}

	return err
}

var runeSweep = []rune{
	'0', '1', '2', '3', '4',
	'5', '6', '7', '8', '9',

	// '⠋',
	// '⠙',
	// '⠹',
	// '⠸',
	// '⠼',
	// '⠴',
	// '⠦',
	// '⠧',
	// '⠇',
	// '⠏',
	// '⣾',
	// '⣽',
	// '⣻',
	// '⢿',
	// '⡿',
	// '⣟',
	// '⣯',
	// '⣷',
	// '⠋',
	// '⠙',
	// '⠚',
	// '⠞',
	// '⠖',
	// '⠦',
	// '⠴',
	// '⠲',
	// '⠳',
	// '⠓',
	// '⠄',
	// '⠆',
	// '⠇',
	// '⠋',
	// '⠙',
	// '⠸',
	// '⠰',
	// '⠠',
	// '⠰',
	// '⠸',
	// '⠙',
	// '⠋',
	// '⠇',
	// '⠆',
	// '⠋',
	// '⠙',
	// '⠚',
	// '⠒',
	// '⠂',
	// '⠂',
	// '⠒',
	// '⠲',
	// '⠴',
	// '⠦',
	// '⠖',
	// '⠒',
	// '⠐',
	// '⠐',
	// '⠒',
	// '⠓',
	// '⠋',
	// '⠁',
	// '⠉',
	// '⠙',
	// '⠚',
	// '⠒',
	// '⠂',
	// '⠂',
	// '⠒',
	// '⠲',
	// '⠴',
	// '⠤',
	// '⠄',
	// '⠄',
	// '⠤',
	// '⠴',
	// '⠲',
	// '⠒',
	// '⠂',
	// '⠂',
	// '⠒',
	// '⠚',
	// '⠙',
	// '⠉',
	// '⠁',
	// '⠈',
	// '⠉',
	// '⠋',
	// '⠓',
	// '⠒',
	// '⠐',
	// '⠐',
	// '⠒',
	// '⠖',
	// '⠦',
	// '⠤',
	// '⠠',
	// '⠠',
	// '⠤',
	// '⠦',
	// '⠖',
	// '⠒',
	// '⠐',
	// '⠐',
	// '⠒',
	// '⠓',
	// '⠋',
	// '⠉',
	// '⠈',
	// '⠁',
	// '⠁',
	// '⠉',
	// '⠙',
	// '⠚',
	// '⠒',
	// '⠂',
	// '⠂',
	// '⠒',
	// '⠲',
	// '⠴',
	// '⠤',
	// '⠄',
	// '⠄',
	// '⠤',
	// '⠠',
	// '⠠',
	// '⠤',
	// '⠦',
	// '⠖',
	// '⠒',
	// '⠐',
	// '⠐',
	// '⠒',
	// '⠓',
	// '⠋',
	// '⠉',
	// '⠈',
	// '⠈',
	// '⢹',
	// '⢺',
	// '⢼',
	// '⣸',
	// '⣇',
	// '⡧',
	// '⡗',
	// '⡏',
	// '⢄',
	// '⢂',
	// '⢁',
	// '⡁',
	// '⡈',
	// '⡐',
	// '⡠',
	// '⠁',
	// '⠂',
	// '⠄',
	// '⡀',
	// '⢀',
	// '⠠',
	// '⠐',
	// '⠈',
	// '-',
	// '\\',
	// '|',
	// '/',
	// '⠂',
	// '-',
	// '–',
	// '—',
	// '–',
	// '-',
	// '┤',
	// '┘',
	// '┴',
	// '└',
	// '├',
	// '┌',
	// '┬',
	// '┐',
	// '✶',
	// '✸',
	// '✹',
	// '✺',
	// '✹',
	// '✷',
	// '+',
	// 'x',
	// '*',
	// '_',
	// '_',
	// '_',
	// '-',
	// '`',
	// '`',
	// '\'',
	// '´',
	// '-',
	// '_',
	// '_',
	// '_',
	// '☱',
	// '☲',
	// '☴',
	// '▁',
	// '▃',
	// '▄',
	// '▅',
	// '▆',
	// '▇',
	// '▆',
	// '▅',
	// '▄',
	// '▃',
	// '▏',
	// '▎',
	// '▍',
	// '▌',
	// '▋',
	// '▊',
	// '▉',
	// '▊',
	// '▋',
	// '▌',
	// '▍',
	// '▎',
	// ' ',
	// '.',
	// 'o',
	// 'O',
	// '@',
	// '*',
	// ' ',
	// '.',
	// 'o',
	// 'O',
	// '°',
	// 'O',
	// 'o',
	// '.',
	// '▓',
	// '▒',
	// '░',
	// '⠁',
	// '⠂',
	// '⠄',
	// '⠂',
	// '▖',
	// '▘',
	// '▝',
	// '▗',
	// '▌',
	// '▀',
	// '▐',
	// '▄',
	// '◢',
	// '◣',
	// '◤',
	// '◥',
	// '◜',
	// '◠',
	// '◝',
	// '◞',
	// '◡',
	// '◟',
	// '◡',
	// '⊙',
	// '◠',
	// '◰',
	// '◳',
	// '◲',
	// '◱',
	// '◴',
	// '◷',
	// '◶',
	// '◵',
	// '◐',
	// '◓',
	// '◑',
	// '◒',
	// '╫',
	// '╪',
	// '⊶',
	// '⊷',
	// '▫',
	// '▪',
	// '□',
	// '■',
	// '■',
	// '□',
	// '▪',
	// '▫',
	// '▮',
	// '▯',
	// 'ဝ',
	// '၀',
	// '⦾',
	// '⦿',
	// '◍',
	// '◌',
	// '◉',
	// '◎',
	// '㊂',
	// '㊀',
	// '㊁',
	// '⧇',
	// '⧆',
	// '☗',
	// '☖',
	// '=',
	// '*',
	// '-',
	// '←',
	// '↖',
	// '↑',
	// '↗',
	// '→',
	// '↘',
	// '↓',
	// '↙',
	// 'd',
	// 'q',
	// 'p',
	// 'b',
	// '-',
	// '=',
	// '≡',

}

/*
var runeSpinners = map[string]runeSpinner{
	{
		name:     "dots",
		interval: 80,
		frames: {
			'⠋',
			'⠙',
			'⠹',
			'⠸',
			'⠼',
			'⠴',
			'⠦',
			'⠧',
			'⠇',
			'⠏',
		},
	},

	{
		name:     "dots2",
		interval: 80,
		frames: {
			'⣾',
			'⣽',
			'⣻',
			'⢿',
			'⡿',
			'⣟',
			'⣯',
			'⣷',
		},
	},

	{
		name:     "dots3",
		interval: 80,
		frames: {
			'⠋',
			'⠙',
			'⠚',
			'⠞',
			'⠖',
			'⠦',
			'⠴',
			'⠲',
			'⠳',
			'⠓',
		},
	},

	{
		name:     "dots4",
		interval: 80,
		frames: {
			'⠄',
			'⠆',
			'⠇',
			'⠋',
			'⠙',
			'⠸',
			'⠰',
			'⠠',
			'⠰',
			'⠸',
			'⠙',
			'⠋',
			'⠇',
			'⠆',
		},
	},

	{
		name:     "dots5",
		interval: 80,
		frames: {
			'⠋',
			'⠙',
			'⠚',
			'⠒',
			'⠂',
			'⠂',
			'⠒',
			'⠲',
			'⠴',
			'⠦',
			'⠖',
			'⠒',
			'⠐',
			'⠐',
			'⠒',
			'⠓',
			'⠋',
		},
	},

	{
		name:     "dots6",
		interval: 80,
		frames: {
			'⠁',
			'⠉',
			'⠙',
			'⠚',
			'⠒',
			'⠂',
			'⠂',
			'⠒',
			'⠲',
			'⠴',
			'⠤',
			'⠄',
			'⠄',
			'⠤',
			'⠴',
			'⠲',
			'⠒',
			'⠂',
			'⠂',
			'⠒',
			'⠚',
			'⠙',
			'⠉',
			'⠁',
		},
	},

	{
		name:     "dots7",
		interval: 80,
		frames: {
			'⠈',
			'⠉',
			'⠋',
			'⠓',
			'⠒',
			'⠐',
			'⠐',
			'⠒',
			'⠖',
			'⠦',
			'⠤',
			'⠠',
			'⠠',
			'⠤',
			'⠦',
			'⠖',
			'⠒',
			'⠐',
			'⠐',
			'⠒',
			'⠓',
			'⠋',
			'⠉',
			'⠈',
		},
	},

	{
		name:     "dots8",
		interval: 80,
		frames: {
			'⠁',
			'⠁',
			'⠉',
			'⠙',
			'⠚',
			'⠒',
			'⠂',
			'⠂',
			'⠒',
			'⠲',
			'⠴',
			'⠤',
			'⠄',
			'⠄',
			'⠤',
			'⠠',
			'⠠',
			'⠤',
			'⠦',
			'⠖',
			'⠒',
			'⠐',
			'⠐',
			'⠒',
			'⠓',
			'⠋',
			'⠉',
			'⠈',
			'⠈',
		},
	},

	{
		name:     "dots9",
		interval: 80,
		frames: {
			'⢹',
			'⢺',
			'⢼',
			'⣸',
			'⣇',
			'⡧',
			'⡗',
			'⡏',
		},
	},

	{
		name:     "dots10",
		interval: 80,
		frames: {
			'⢄',
			'⢂',
			'⢁',
			'⡁',
			'⡈',
			'⡐',
			'⡠',
		},
	},

	{
		name:     "dots11",
		interval: 100,
		frames: {
			'⠁',
			'⠂',
			'⠄',
			'⡀',
			'⢀',
			'⠠',
			'⠐',
			'⠈',
		},
	},

	{
		name:     "line",
		interval: 130,
		frames: {
			'-',
			'\\',
			'|',
			'/',
		},
	},

	{
		name:     "line2",
		interval: 100,
		frames: {
			'⠂',
			'-',
			'–',
			'—',
			'–',
			'-',
		},
	},

	{
		name:     "pipe",
		interval: 100,
		frames: {
			'┤',
			'┘',
			'┴',
			'└',
			'├',
			'┌',
			'┬',
			'┐',
		},
	},

	{
		name:     "star",
		interval: 70,
		frames: {
			'✶',
			'✸',
			'✹',
			'✺',
			'✹',
			'✷',
		},
	},

	{
		name:     "star2",
		interval: 80,
		frames: {
			'+',
			'x',
			'*',
		},
	},

	{
		name:     "flip",
		interval: 70,
		frames: {
			'_',
			'_',
			'_',
			'-',
			'`',
			'`',
			'\'',
			'´',
			'-',
			'_',
			'_',
			'_',
		},
	},

	{
		name:     "hamburger",
		interval: 100,
		frames: {
			'☱',
			'☲',
			'☴',
		},
	},

	{
		name:     "growVertical",
		interval: 120,
		frames: {
			'▁',
			'▃',
			'▄',
			'▅',
			'▆',
			'▇',
			'▆',
			'▅',
			'▄',
			'▃',
		},
	},

	{
		name:     "growHorizontal",
		interval: 120,
		frames: {
			'▏',
			'▎',
			'▍',
			'▌',
			'▋',
			'▊',
			'▉',
			'▊',
			'▋',
			'▌',
			'▍',
			'▎',
		},
	},

	{
		name:     "balloon",
		interval: 140,
		frames: {
			' ',
			'.',
			'o',
			'O',
			'@',
			'*',
			' ',
		},
	},

	{
		name:     "balloon2",
		interval: 120,
		frames: {
			'.',
			'o',
			'O',
			'°',
			'O',
			'o',
			'.',
		},
	},

	{
		name:     "noise",
		interval: 100,
		frames: {
			'▓',
			'▒',
			'░',
		},
	},

	{
		name:     "bounce",
		interval: 120,
		frames: {
			'⠁',
			'⠂',
			'⠄',
			'⠂',
		},
	},

	{
		name:     "boxBounce",
		interval: 120,
		frames: {
			'▖',
			'▘',
			'▝',
			'▗',
		},
	},

	{
		name:     "boxBounce2",
		interval: 100,
		frames: {
			'▌',
			'▀',
			'▐',
			'▄',
		},
	},

	{
		name:     "triangle",
		interval: 50,
		frames: {
			'◢',
			'◣',
			'◤',
			'◥',
		},
	},

	{
		name:     "arc",
		interval: 100,
		frames: {
			'◜',
			'◠',
			'◝',
			'◞',
			'◡',
			'◟',
		},
	},

	{
		name:     "circle",
		interval: 120,
		frames: {
			'◡',
			'⊙',
			'◠',
		},
	},

	{
		name:     "squareCorners",
		interval: 180,
		frames: {
			'◰',
			'◳',
			'◲',
			'◱',
		},
	},

	{
		name:     "circleQuarters",
		interval: 120,
		frames: {
			'◴',
			'◷',
			'◶',
			'◵',
		},
	},

	{
		name:     "circleHalves",
		interval: 50,
		frames: {
			'◐',
			'◓',
			'◑',
			'◒',
		},
	},

	{
		name:     "squish",
		interval: 100,
		frames: {
			'╫',
			'╪',
		},
	},

	{
		name:     "toggle",
		interval: 250,
		frames: {
			'⊶',
			'⊷',
		},
	},

	{
		name:     "toggle2",
		interval: 80,
		frames: {
			'▫',
			'▪',
		},
	},

	{
		name:     "toggle3",
		interval: 120,
		frames: {
			'□',
			'■',
		},
	},

	{
		name:     "toggle4",
		interval: 100,
		frames: {
			'■',
			'□',
			'▪',
			'▫',
		},
	},

	{
		name:     "toggle5",
		interval: 100,
		frames: {
			'▮',
			'▯',
		},
	},

	{
		name:     "toggle6",
		interval: 300,
		frames: {
			'ဝ',
			'၀',
		},
	},

	{
		name:     "toggle7",
		interval: 80,
		frames: {
			'⦾',
			'⦿',
		},
	},

	{
		name:     "toggle8",
		interval: 100,
		frames: {
			'◍',
			'◌',
		},
	},

	{
		name:     "toggle9",
		interval: 100,
		frames: {
			'◉',
			'◎',
		},
	},

	{
		name:     "toggle10",
		interval: 100,
		frames: {
			'㊂',
			'㊀',
			'㊁',
		},
	},

	{
		name:     "toggle11",
		interval: 50,
		frames: {
			'⧇',
			'⧆',
		},
	},

	{
		name:     "toggle12",
		interval: 120,
		frames: {
			'☗',
			'☖',
		},
	},

	{
		name:     "toggle13",
		interval: 80,
		frames: {
			'=',
			'*',
			'-',
		},
	},

	{
		name:     "arrow",
		interval: 100,
		frames: {
			'←',
			'↖',
			'↑',
			'↗',
			'→',
			'↘',
			'↓',
			'↙',
		},
	},

	{
		name:     "dqpb",
		interval: 100,
		frames: {
			'd',
			'q',
			'p',
			'b',
		},
	},

	{
		name:     "layer",
		interval: 150,
		frames: {
			'-',
			'=',
			'≡',
		},
	},
}

var stringSpinners = map[string]struct {
	name     string
	interval int
	frames   []rune
}{
	{
		name:     "dots12",
		interval: 80,
		frames: {
			"⢀⠀",
			"⡀⠀",
			"⠄⠀",
			"⢂⠀",
			"⡂⠀",
			"⠅⠀",
			"⢃⠀",
			"⡃⠀",
			"⠍⠀",
			"⢋⠀",
			"⡋⠀",
			"⠍⠁",
			"⢋⠁",
			"⡋⠁",
			"⠍⠉",
			"⠋⠉",
			"⠋⠉",
			"⠉⠙",
			"⠉⠙",
			"⠉⠩",
			"⠈⢙",
			"⠈⡙",
			"⢈⠩",
			"⡀⢙",
			"⠄⡙",
			"⢂⠩",
			"⡂⢘",
			"⠅⡘",
			"⢃⠨",
			"⡃⢐",
			"⠍⡐",
			"⢋⠠",
			"⡋⢀",
			"⠍⡁",
			"⢋⠁",
			"⡋⠁",
			"⠍⠉",
			"⠋⠉",
			"⠋⠉",
			"⠉⠙",
			"⠉⠙",
			"⠉⠩",
			"⠈⢙",
			"⠈⡙",
			"⠈⠩",
			"⠀⢙",
			"⠀⡙",
			"⠀⠩",
			"⠀⢘",
			"⠀⡘",
			"⠀⠨",
			"⠀⢐",
			"⠀⡐",
			"⠀⠠",
			"⠀⢀",
			"⠀⡀",
		},
	},

	{
		name:     "simpleDots",
		interval: 400,
		frames: {
			".  ",
			".. ",
			"...",
			"   ",
		},
	},

	{
		name:     "simpleDotsScrolling",
		interval: 200,
		frames: {
			".  ",
			".. ",
			"...",
			" ..",
			"  .",
			"   ",
		},
	},

	{
		name:     "arrow2",
		interval: 80,
		frames: {
			"⬆️ ",
			"↗️ ",
			"➡️ ",
			"↘️ ",
			"⬇️ ",
			"↙️ ",
			"⬅️ ",
			"↖️ ",
		},
	},

	{
		name:     "arrow3",
		interval: 120,
		frames: {
			"▹▹▹▹▹",
			"▸▹▹▹▹",
			"▹▸▹▹▹",
			"▹▹▸▹▹",
			"▹▹▹▸▹",
			"▹▹▹▹▸",
		},
	},

	{
		name:     "bouncingBar",
		interval: 80,
		frames: {
			"[    ]",
			"[=   ]",
			"[==  ]",
			"[=== ]",
			"[ ===]",
			"[  ==]",
			"[   =]",
			"[    ]",
			"[   =]",
			"[  ==]",
			"[ ===]",
			"[====]",
			"[=== ]",
			"[==  ]",
			"[=   ]",
		},
	},

	{
		name:     "bouncingBall",
		interval: 80,
		frames: {
			"( ●    )",
			"(  ●   )",
			"(   ●  )",
			"(    ● )",
			"(     ●)",
			"(    ● )",
			"(   ●  )",
			"(  ●   )",
			"( ●    )",
			"(●     )",
		},
	},

	{
		name:     "smiley",
		interval: 200,
		frames: {
			"😄 ",
			"😝 ",
		},
	},

	{
		name:     "monkey",
		interval: 300,
		frames: {
			"🙈 ",
			"🙈 ",
			"🙉 ",
			"🙊 ",
		},
	},

	{
		name:     "hearts",
		interval: 100,
		frames: {
			"💛 ",
			"💙 ",
			"💜 ",
			"💚 ",
			"❤️ ",
		},
	},

	{
		name:     "clock",
		interval: 100,
		frames: {
			"🕛 ",
			"🕐 ",
			"🕑 ",
			"🕒 ",
			"🕓 ",
			"🕔 ",
			"🕕 ",
			"🕖 ",
			"🕗 ",
			"🕘 ",
			"🕙 ",
			"🕚 ",
		},
	},

	{
		name:     "earth",
		interval: 180,
		frames: {
			"🌍 ",
			"🌎 ",
			"🌏 ",
		},
	},

	{
		name:     "moon",
		interval: 80,
		frames: {
			"🌑 ",
			"🌒 ",
			"🌓 ",
			"🌔 ",
			"🌕 ",
			"🌖 ",
			"🌗 ",
			"🌘 ",
		},
	},

	{
		name:     "runner",
		interval: 140,
		frames: {
			"🚶 ",
			"🏃 ",
		},
	},

	{
		name:     "pong",
		interval: 80,
		frames: {
			"▐⠂       ▌",
			"▐⠈       ▌",
			"▐ ⠂      ▌",
			"▐ ⠠      ▌",
			"▐  ⡀     ▌",
			"▐  ⠠     ▌",
			"▐   ⠂    ▌",
			"▐   ⠈    ▌",
			"▐    ⠂   ▌",
			"▐    ⠠   ▌",
			"▐     ⡀  ▌",
			"▐     ⠠  ▌",
			"▐      ⠂ ▌",
			"▐      ⠈ ▌",
			"▐       ⠂▌",
			"▐       ⠠▌",
			"▐       ⡀▌",
			"▐      ⠠ ▌",
			"▐      ⠂ ▌",
			"▐     ⠈  ▌",
			"▐     ⠂  ▌",
			"▐    ⠠   ▌",
			"▐    ⡀   ▌",
			"▐   ⠠    ▌",
			"▐   ⠂    ▌",
			"▐  ⠈     ▌",
			"▐  ⠂     ▌",
			"▐ ⠠      ▌",
			"▐ ⡀      ▌",
			"▐⠠       ▌",
		},
	},

	{
		name:     "shark",
		interval: 120,
		frames: {
			"▐|\\____________▌",
			"▐_|\\___________▌",
			"▐__|\\__________▌",
			"▐___|\\_________▌",
			"▐____|\\________▌",
			"▐_____|\\_______▌",
			"▐______|\\______▌",
			"▐_______|\\_____▌",
			"▐________|\\____▌",
			"▐_________|\\___▌",
			"▐__________|\\__▌",
			"▐___________|\\_▌",
			"▐____________|\\▌",
			"▐____________/|▌",
			"▐___________/|_▌",
			"▐__________/|__▌",
			"▐_________/|___▌",
			"▐________/|____▌",
			"▐_______/|_____▌",
			"▐______/|______▌",
			"▐_____/|_______▌",
			"▐____/|________▌",
			"▐___/|_________▌",
			"▐__/|__________▌",
			"▐_/|___________▌",
			"▐/|____________▌",
		},
	},

	{
		name:     "weather",
		interval: 100,
		frames: {
			"☀️ ",
			"☀️ ",
			"☀️ ",
			"🌤 ",
			"⛅️ ",
			"🌥 ",
			"☁️ ",
			"🌧 ",
			"🌨 ",
			"🌧 ",
			"🌨 ",
			"🌧 ",
			"🌨 ",
			"⛈ ",
			"🌨 ",
			"🌧 ",
			"🌨 ",
			"☁️ ",
			"🌥 ",
			"⛅️ ",
			"🌤 ",
			"☀️ ",
			"☀️ ",
		},
	},

	{
		name:     "christmas",
		interval: 400,
		frames: {
			"🌲",
			"🎄",
		},
	},

	{
		name:     "grenade",
		interval: 80,
		frames: {
			"،   ",
			"′   ",
			" ´ ",
			" ‾ ",
			"  ⸌",
			"  ⸊",
			"  |",
			"  ⁎",
			"  ⁕",
			" ෴ ",
			"  ⁓",
			"   ",
			"   ",
			"   ",
		},
	},

	{
		name:     "point",
		interval: 125,
		frames: {
			"∙∙∙",
			"●∙∙",
			"∙●∙",
			"∙∙●",
			"∙∙∙",
		},
	},
}
*/
