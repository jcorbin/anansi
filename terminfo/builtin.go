package terminfo

import (
	"errors"
	"fmt"
	"strings"
)

func getBuiltin(term string) (*Terminfo, error) {
	if term == "" {
		if defaultTerm == "" {
			return nil, errors.New("no term name given, and no default defined")
		}
		term = defaultTerm
	}
	if ti, def := builtins[term]; def {
		return ti, nil
	}
	for _, compat := range compatTable {
		if strings.Contains(term, compat.partial) {
			return compat.Terminfo, nil
		}
	}
	return nil, fmt.Errorf("unsupported TERM=%q", term)
}
