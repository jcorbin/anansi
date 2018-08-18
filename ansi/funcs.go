package ansi

// Private two-character escape sequences (allowed by ANSI X3.41-1974)

var (
	// DECGON graphics on for VT105, DECHTS horiz tab set for LA34/LA120
	DECGON = ESC('1')

	// DECGOFF graphics off VT105, DECCAHT clear all horz tabs LA34/LA120
	DECGOFF = ESC('2')

	// DECVTS set vertical tab for LA34/LA120
	DECVTS = ESC('3')

	// DECCAVT clear all vertical tabs for LA34/LA120
	DECCAVT = ESC('4')

	// DECXMT Host requests that VT132 transmit as if ENTER were pressed
	DECXMT = ESC('5')

	// DECSC Save cursor position and character attributes
	DECSC = ESC('7')

	// DECRC Restore cursor and attributes to previously saved position
	DECRC = ESC('8')

	// DECANSI Switch from VT52 mode to VT100 mode
	DECANSI = ESC('<')

	// DECKPAM Set keypad to applications mode (ESCape instead of digits)
	DECKPAM = ESC('=')

	// DECKPNM Set keypad to numeric mode (digits intead of ESCape seq)
	DECKPNM = ESC('>')
)

// Control Sequences (defined by ANSI X3.64-1979)
var (
	/*ICH Insert CHaracter
	  [10@ = Make room for 10 characters at current position */
	ICH = CSI('@')

	/*CUU CUrsor Up
	  [A = Move up one line, stop at top of screen, [9A = move up 9 */
	CUU = CSI('A')

	/*CUD CUrsor Down
	  [B = Move down one line, stop at bottom of screen */
	CUD = CSI('B')

	/*CUF CUrsor Forward
	  [C = Move forward one position, stop at right edge of screen */
	CUF = CSI('C')

	/*CUB CUrsor Backward
	  [D = Same as BackSpace, stop at left edge of screen */
	CUB = CSI('D')

	/*CNL Cursor to Next Line
	  [5E = Move to first position of 5th line down */
	CNL = CSI('E')

	/*CPL Cursor to Previous Line
	  [5F = Move to first position of 5th line previous */
	CPL = CSI('F')

	/*CHA Cursor Horizontal position Absolute
	  [40G = Move to column 40 of current line */
	CHA = CSI('G')

	/*CUP CUrsor Position
	  [H = Home
	  [24;80H = Row 24, Column 80 */
	CUP = CSI('H')

	/*CHT Cursor Horizontal Tabulation
	  [I = Same as HT (Control-I), [3I = Go forward 3 tabs */
	CHT = CSI('I')

	/*ED Erase in Display (cursor does not move)
	  [J =
	  [0J = Erase from current position to end (inclusive)
	  [1J = Erase from beginning to current position (inclusive)
	  [2J = Erase entire display
	  [?0J = Selective erase in display ([?1J, [?2J similar) */
	ED = CSI('J')

	/*EL Erase in Line (cursor does not move)
	  [K = [0K = Erase from current position to end (inclusive)
	  [1K = Erase from beginning to current position
	  [2K = Erase entire current line
	  [?0K = Selective erase to end of line ([?1K, [?2K similar) */
	EL = CSI('K')

	/*IL Insert Line, current line moves down (VT102 series)
	  [3L = Insert 3 lines if currently in scrolling region */
	IL = CSI('L')

	/*DL Delete Line, lines below current move up (VT102 series)
	  [2M = Delete 2 lines if currently in scrolling region */
	DL = CSI('M')

	/*EF Erase in Field (as bounded by protected fields)
	  [0N, [1N, [2N act like [L but within current field */
	EF = CSI('N')

	/*EA Erase in qualified Area (defined by DAQ)
	  [0O, [1O, [2O act like [J but within current area */
	EA = CSI('O')

	/*DCH Delete Character, from current position to end of field
	  [4P = Delete 4 characters, VT102 series */
	DCH = CSI('P')

	/*SEM Set Editing extent Mode (limits ICH and DCH)
	  [0Q = [Q = Insert/delete character affects rest of display
	  [1Q = ICH/DCH affect the current line only
	  [2Q = ICH/DCH affect current field (between tab stops) only
	  [3Q = ICH/DCH affect qualified area (between protected fields) */
	SEM = CSI('Q')

	/*CPR Cursor Position Report (from terminal to host)
	  [24;80R = Cursor is positioned at line 24 column 80 */
	CPR = CSI('R')

	/*SU Scroll up, entire display is moved up, new lines at bottom
	  [3S = Move everything up 3 lines, bring in 3 new lines */
	SU = CSI('S')

	/*SD Scroll down, new lines inserted at top of screen
	  [4T = Scroll down 4, bring previous lines back into view */
	SD = CSI('T')

	/*NP Next Page (if terminal has more than 1 page of memory)
	  [2U = Scroll forward 2 pages */
	NP = CSI('U')

	/*PP Previous Page (if terminal remembers lines scrolled off top)
	  [1V = Scroll backward 1 page */
	PP = CSI('V')

	/*CTC Cursor Tabulation Control
	  [0W = Set horizontal tab for current line at current position
	  [1W = Set vertical tab stop for current line of current page
	  [2W = Clear horiz tab stop at current position of current line
	  [3W = Clear vert tab stop at current line of current page
	  [4W = Clear all horiz tab stops on current line only
	  [5W = Clear all horiz tab stops for the entire terminal
	  [6W = Clear all vert tabs stops for the entire terminal */
	CTC = CSI('W')

	/*ECH Erase CHaracter
	  [4X = Change next 4 characters to "erased" state */
	ECH = CSI('X')

	/*CVT Cursor Vertical Tab
	  [2Y = Move forward to 2nd following vertical tab stop */
	CVT = CSI('Y')

	/*CBT Cursor Back Tab
	  [3Z = Move backwards to 3rd previous horizontal tab stop */
	CBT = CSI('Z')

	/*HPA Horizontal Position Absolute (depends on PUM)
	  [720` = Move to 720 decipoints (1 inch) from left margin
	  [80` = Move to column 80 on LA120 */
	HPA = CSI('`')

	/*HPR Horizontal Position Relative (depends on PUM)
	  [360a = Move 360 decipoints (1/2 inch) from current position
	  [40a = Move 40 columns to right of current position on LA120 */
	HPR = CSI('a')

	/*REP REPeat previous displayable character
	  [80b = Repeat character 80 times */
	REP = CSI('b')

	/*DA Device Attributes
	  [c = Terminal will identify itself
	  [?1;2c = Terminal is saying it is a VT100 with AVO
	  [>0c = Secondary DA request (distinguishes VT240 from VT220) */
	DA = CSI('c')

	/*VPA Vertical Position Absolute (depends on PUM)
	  [90d = Move to 90 decipoints (1/8 inch) from top margin
	  [10d = Move to line 10 if before that else line 10 next page */
	VPA = CSI('d')

	/*VPR Vertical Position Relative (depends on PUM)
	  [720e = Move 720 decipoints (1 inch) down from current position
	  [6e = Advance 6 lines forward on LA120 */
	VPR = CSI('e')

	/*HVP Horizontal and Vertical Position (depends on PUM)
	  [720,1440f = Move to 1 inch down and 2 inches over (decipoints)
	  [24;80f = Move to row 24 column 80 if PUM is set to character */
	HVP = CSI('f')

	/*TBC Tabulation Clear
	  [0g = Clear horizontal tab stop at current position
	  [1g = Clear vertical tab stop at current line (LA120)
	  [2g = Clear all horizontal tab stops on current line only LA120
	  [3g = Clear all horizontal tab stops in the terminal */
	TBC = CSI('g')

	/*SM Set Standard Mode (. means permanently set on VT100)
	  TODO merge with constants defined in modes.go
	  [0h = Error, this command is ignored
	  [1h = GATM - Guarded Area Transmit Mode, send all (VT132)
	  [2h = KAM - Keyboard Action Mode, disable keyboard input
	  [3h = CRM - Control Representation Mode, show all control chars
	  [4h = IRM - Insertion/Replacement Mode, set insert mode (VT102)
	  [5h = SRTM - Status Report Transfer Mode, report after DCS
	  [6h = ERM - ERasure Mode, erase protected and unprotected
	  [7h = VEM - Vertical Editing Mode, IL/DL affect previous lines
	  [8h, [9h are reserved
	  [10h = HEM - Horizontal Editing mode, ICH/DCH/IRM go backwards
	  [11h = PUM - Positioning Unit Mode, use decipoints for HVP/etc
	  [12h = SRM - Send Receive Mode, transmit without local echo
	  [13h = FEAM - Format Effector Action Mode, FE's are stored
	  [14h = FETM - Format Effector Transfer Mode, send only if stored
	  [15h = MATM - Multiple Area Transfer Mode, send all areas
	  [16h = TTM - Transmit Termination Mode, send scrolling region
	  [17h = SATM - Send Area Transmit Mode, send entire buffer
	  [18h = TSM - Tabulation Stop Mode, lines are independent
	  [19h = EBM - Editing Boundry Mode, all of memory affected
	  [20h = LNM - Linefeed Newline Mode, LF interpreted as CR LF */
	SM = CSI('h')

	/*SMprivate Set Private Mode
	  TODO merge with constants defined in modes.go
	  [?1h = DECCKM - Cursor Keys Mode, send ESC O A for cursor up
	  [?2h = DECANM - ANSI Mode, use ESC < to switch VT52 to ANSI
	  [?3h = DECCOLM - COLumn mode, 132 characters per line
	  [?4h = DECSCLM - SCrolL Mode, smooth scrolling
	  [?5h = DECSCNM - SCreeN Mode, black on white background
	  [?6h = DECOM - Origin Mode, line 1 is relative to scroll region
	  [?7h = DECAWM - AutoWrap Mode, start newline after column 80
	  [?8h = DECARM - Auto Repeat Mode, key will autorepeat
	  [?9h = DECINLM - INterLace Mode, interlaced for taking photos
	  [?10h = DECEDM - EDit Mode, VT132 is in EDIT mode
	  [?11h = DECLTM - Line Transmit Mode, ignore TTM, send line
	  [?12h = ?
	  [?13h = DECSCFDM - Space Compression/Field Delimiting on,
	  [?14h = DECTEM - Transmit Execution Mode, transmit on ENTER
	  [?15h = ?
	  [?16h = DECEKEM - Edit Key Execution Mode, EDIT key is local
	  [?17h = ?
	  [?18h = DECPFF - Print FormFeed mode, send FF after printscreen
	  [?19h = DECPEXT - Print Extent mode, print entire screen
	  [?20h = OV1 - Overstrike, overlay characters on GIGI
	  [?21h = BA1 - Local BASIC, GIGI to keyboard and screen
	  [?22h = BA2 - Host BASIC, GIGI to host computer
	  [?23h = PK1 - GIGI numeric keypad sends reprogrammable sequences
	  [?24h = AH1 - Autohardcopy before erasing or rolling GIGI screen
	  [?29h =     - Use only the proper pitch for the LA100 font
	  [?38h = DECTEK - TEKtronix mode graphics */
	SMprivate = SM.With('?')

	/*MC Media Copy (printer port on VT102)
	  [0i = Send contents of text screen to printer
	  [1i = Fill screen from auxiliary input (printer's keyboard)
	  [2i = Send screen to secondary output device
	  [3i = Fill screen from secondary input device
	  [4i = Turn on copying received data to primary output (VT125)
	  [4i = Received data goes to VT102 screen, not to its printer
	  [5i = Turn off copying received data to primary output (VT125)
	  [5i = Received data goes to VT102's printer, not its screen
	  [6i = Turn off copying received data to secondary output (VT125)
	  [7i = Turn on copying received data to secondary output (VT125)
	  [?0i = Graphics screen dump goes to graphics printer VT125,VT240
	  [?1i = Print cursor line, terminated by CR LF
	  [?2i = Graphics screen dump goes to host computer VT125,VT240
	  [?4i = Disable auto print
	  [?5i = Auto print, send a line at a time when linefeed received */
	MC = CSI('i')

	/*RM Reset Mode (. means permanently reset on VT100)
	  TODO merge with constants defined in modes.go
	  [1l = GATM - Transmit only unprotected characters (VT132)
	  [2l = KAM - Enable input from keyboard
	  [3l = CRM - Control characters are not displayable characters
	  [4l = IRM - Reset to replacement mode (VT102)
	  [5l = SRTM - Report only on command (DSR)
	  [6l = ERM - Erase only unprotected fields
	  [7l = VEM - IL/DL affect lines after current line
	  [8l reserved
	  [9l reserved
	  [10l = HEM - ICH and IRM shove characters forward, DCH pulls
	  [11l = PUM - Use character positions for HPA/HPR/VPA/VPR/HVP
	  [12l = SRM - Local echo - input from keyboard sent to screen
	  [13l = FEAM - HPA/VPA/SGR/etc are acted upon when received
	  [14l = FETM - Format Effectors are sent to the printer
	  [15l = MATM - Send only current area if SATM is reset
	  [16l = TTM - Transmit partial page, up to cursor position
	  [17l = SATM - Transmit areas bounded by SSA/ESA/DAQ
	  [18l = TSM - Setting a tab stop on one line affects all lines
	  [19l = EBM - Insert does not overflow to next page
	  [20l = LNM - Linefeed does not change horizontal position */
	RM = CSI('l')

	/*RMprivate Reset Private Mode
	  [?1l = DECCKM - Cursor keys send ANSI cursor position commands
	  [?2l = DECANM - Use VT52 emulation instead of ANSI mode
	  [?3l = DECCOLM - 80 characters per line (erases screen)
	  [?4l = DECSCLM - Jump scrolling
	  [?5l = DECSCNM - Normal screen (white on black background)
	  [?6l = DECOM - Line numbers are independent of scrolling region
	  [?7l = DECAWM - Cursor remains at end of line after column 80
	  [?8l = DECARM - Keys do not repeat when held down
	  [?9l = DECINLM - Display is not interlaced to avoid flicker
	  [?10l = DECEDM - VT132 transmits all key presses
	  [?11l = DECLTM - Send page or partial page depending on TTM
	  [?12l = ?
	  [?13l = DECSCFDM - Don't suppress trailing spaces on transmit
	  [?14l = DECTEM - ENTER sends ESC S (STS) a request to send
	  [?15l = ?
	  [?16l = DECEKEM - EDIT key transmits either $[10h or $[10l
	  [?17l = ?
	  [?18l = DECPFF - Don't send a formfeed after printing screen
	  [?19l = DECPEXT - Print only the lines within the scroll region
	  [?20l = OV0 - Space is destructive, replace not overstrike, GIGI
	  [?21l = BA0 - No BASIC, GIGI is On-Line or Local
	  [?22l = BA0 - No BASIC, GIGI is On-Line or Local
	  [?23l = PK0 - Ignore reprogramming on GIGI keypad and cursors
	  [?24l = AH0 - No auto-hardcopy when GIGI screen erased
	  [?29l = Allow all character pitches on the LA100
	  [?38l = DECTEK - Ignore TEKtronix graphics commands */
	RMprivate = RM.With('?')

	/*SGR Set Graphics Rendition (affects character attributes)
	  [0m = Clear all special attributes
	  [1m = Bold or increased intensity
	  [2m = Dim or secondary color on GIGI  (superscript on XXXXXX)
	  [3m = Italic                          (subscript on XXXXXX)
	  [4m = Underscore
	  [0;4m = Clear, then set underline only
	  [5m = Slow blink
	  [6m = Fast blink                      (overscore on XXXXXX)
	  [7m = Negative image
	  [0;1;7m = Bold + Inverse
	  [8m = Concealed (do not display character echoed locally)
	  [9m = Reserved for future standardization
	  [10m = Select primary font (LA100)
	  [11m -
	  [19m = Select alternate font (LA100 has 11 thru 14)
	  [20m = FRAKTUR (whatever that means)
	  [22m = Cancel bold or dim attribute only (VT220)
	  [24m = Cancel underline attribute only (VT220)
	  [25m = Cancel fast or slow blink attribute only (VT220)
	  [27m = Cancel negative image attribute only (VT220)
	  [30m = Write with black
	  [31m = Write with red
	  [32m = Write with green
	  [33m = Write with yellow
	  [34m = Write with blue
	  [35m = Write with magenta
	  [36m = Write with cyan
	  [37m = Write with white
	  [38m reserved
	  [39m reserved
	  [40m = Set background to black (GIGI)
	  [41m = Set background to red
	  [42m = Set background to green
	  [43m = Set background to yellow
	  [44m = Set background to blue
	  [45m = Set background to magenta
	  [46m = Set background to cyan
	  [47m = Set background to white
	  [48m reserved
	  [49m reserved
	*/
	SGR = CSI('m')

	/*DSR Device Status Report
	  [0n = Terminal is ready, no malfunctions detected
	  [1n = Terminal is busy, retry later
	  [2n = Terminal is busy, it will send DSR when ready
	  [3n = Malfunction, please try again
	  [4n = Malfunction, terminal will send DSR when ready
	  [5n = Command to terminal to report its status
	  [6n = Command to terminal requesting cursor position (CPR)
	  [?15n = Command to terminal requesting printer status, returns
	          [?10n = OK
	          [?11n = not OK
	          [?13n = no printer.
	  [?25n = "Are User Defined Keys Locked?" (VT220) */
	DSR = CSI('n')

	/*DAQ Define Area Qualification starting at current position
	  [0o = Accept all input, transmit on request
	  [1o = Protected and guarded, accept no input, do not transmit
	  [2o = Accept any printing character in this field
	  [3o = Numeric only field
	  [4o = Alphabetic (A-Z and a-z) only
	  [5o = Right justify in area
	  [3;6o = Zero fill in area
	  [7o = Set horizontal tab stop, this is the start of the field
	  [8o = Protected and unguarded, accept no input, do transmit
	  [9o = Space fill in area */
	DAQ = CSI('o')
)

// Private Control Sequences (allowed by ANSI X3.41-1974)
var (
	// DECSTR Soft Terminal Reset
	// [!p = Soft Terminal Reset
	DECSTR = CSI('p')

	// SoftReset is a control sequence that causes a soft terminal reset.
	// NOTE it does not erase the screen or home the cursor; for that also
	// send ED.With('2') and CUP.
	SoftReset = DECSTR.With('!')

	/*DECLL Load LEDs
	  [0q           = Turn off all
	  [?1;4q        = turns on L1 and L4, etc
	  [154;155;157q = VT100 goes bonkers
	  [2;23!q       = Partial screen dump from GIGI to graphics printer
	  [0"q          = DECSCA Select Character Attributes off
	  [1"q          = DECSCA - designate set as non-erasable
	  [2"q          = DECSCA - designate set as erasable */
	DECLL = CSI('q')

	/*DECSTBM Set top and bottom margins (scroll region on VT100)
	  [4;20r = Set top margin at line 4 and bottom at line 20 */
	DECSTBM = CSI('r')

	/*DECSTRM Set left and right margins on LA100,LA120
	  [5;130s = Set left margin at column 5 and right at column 130 */
	DECSTRM = CSI('s')

	/*DECSLPP Set physical lines per page
	  [66t = Paper has 66 lines (11 inches at 6 per inch) */
	DECSLPP = CSI('t')

	/*DECSHTS Set many horizontal tab stops at once on LA100
	  [9;17;25;33;41;49;57;65;73;81u = Set standard tab stops */
	DECSHTS = CSI('u')

	/*DECSVTS Set many vertical tab stops at once on LA100
	  [1;16;31;45v = Set vert tabs every 15 lines */
	DECSVTS = CSI('v')

	/*DECSHORP Set horizontal pitch on LAxxx printers
	  [1w = 10 characters per inch
	  [2w = 12 characters per inch
	  [0w = 10
	  [3w = 13.2
	  [4w = 16.5
	  [5w = 5
	  [6w = 6
	  [7w = 6.6
	  [8w = 8.25 */
	DECSHORP = CSI('w')

	/*DECREQTPARM Request terminal parameters
	  [3;5;2;64;64;1;0x = Report, 7 bit Even, 1200 baud, 1200 baud */
	DECREQTPARM = CSI('x')

	/*DECTST Invoke confidence test
	  [2;1y = Power-up test on VT100 series (and VT100 part of VT125)
	  [3;1y = Power-up test on GIGI (VK100)
	  [4;1y = Power-up test on graphics portion of VT125 */
	DECTST = CSI('y')

	/*DECVERP Set vertical pitch on LA100
	  [1z = 6 lines per inch
	  [2z = 8 lines per inch
	  [0z = 6
	  [3z = 12
	  [4z = 3
	  [5z = 3
	  [6z = 4 */
	DECVERP = CSI('z')

	/*DECTTC Transmit Termination Character
	                [0| = No extra characters
					[1| = terminate with FF */
	DECTTC = CSI('|')

	/*DECPRO Define protected field on VT132
	  [0}       = No protection
	  [1;4;5;7} = Any attribute is protected
	  [254}     = Characters with no attributes are protected */
	DECPRO = CSI('}')

	/*DECKEYS Sent by special function keys
	  [1~  = FIND
	  [2~  = INSERT
	  [3~  = REMOVE
	  [4~  = SELECT
	  [5~  = PREV
	  [6~  = NEXT
	  [17~ = F6...
	  [34~ = F20
	  [23~ = ESC
	  [24~ = BS
	  [25~ = LF
	  [28~ = HELP
	  [29~ = DO */
	DECKEYS = CSI('~')
)
