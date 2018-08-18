package platform

import (
	"encoding/gob"
	"io"
)

type state struct {
	State
	TelemetryState
	HUDState
	LogViewState
}

func (p *Platform) readState(r io.Reader) error {
	var st state
	dec := gob.NewDecoder(r)
	for _, obj := range []interface{}{&st, p.client} {
		if err := dec.Decode(obj); err != nil {
			return err
		}
	}
	p.State = st.State
	p.Telemetry.TelemetryState = st.TelemetryState
	p.HUD.HUDState = st.HUDState
	p.HUD.logs.LogViewState = st.LogViewState
	return nil
}

func (p *Platform) writeState(w io.Writer) error {
	st := state{
		p.State,
		p.Telemetry.TelemetryState,
		p.HUD.HUDState,
		p.HUD.logs.LogViewState,
	}
	enc := gob.NewEncoder(w)
	for _, obj := range []interface{}{st, p.client} {
		if err := enc.Encode(obj); err != nil {
			return err
		}
	}
	return nil
}
