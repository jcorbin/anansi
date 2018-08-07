package terminal

import (
	copsDisp "github.com/jcorbin/execs/internal/cops/display"
)

// TODO fully steal or revert to properly borrowing cops's cursor

// StartCursor is the initial cursor state.
var StartCursor = copsDisp.Start

// Cursor represents the terminal cursor.
type Cursor = copsDisp.Cursor
