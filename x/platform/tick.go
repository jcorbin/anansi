package platform

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/jcorbin/anansi"
)

// PID controller constants; TODO tune better
const (
	ticksFactorP = 0.5
	ticksFactorI = 0.1
	ticksFactorD = 0.1
)

// TicksMetric contains metric data from the last round of a Tics.Wait() loop.
type TicksMetric struct {
	Wakeup  time.Time
	Delay   time.Duration
	Elapsed time.Duration
	Return  time.Time
	TicksState
}

// WriteToBuffer writes the metric into the given bytes buffer.
func (m *TicksMetric) WriteToBuffer(buf *bytes.Buffer) {
	fmt.Fprintf(buf,
		`{"wakeup":%d,"delay":%d,"elapsed":%d,"return":%d,`+
			`"goal":%d,"goal_delta":%d,"set_delta":%d,`+
			`"last_time":%d,"last_delta":%d,`+
			`"err":%d,"cum_err":%f,"delta_diff":%d,"adjust":%d}`,
		m.Wakeup.UnixNano(),
		m.Delay,
		m.Elapsed,
		m.Return.UnixNano(),
		m.Goal,
		m.GoalDelta,
		m.SetDelta,
		m.LastTime.UnixNano(),
		m.LastDelta,
		m.Err,
		m.CumErr,
		m.DeltaDiff,
		m.Adjust,
	)
}

// NewTicks creates a new animation tick controller; see Ticks for a detailed
// description.
func NewTicks(goalTPS int) *Ticks {
	// compute an exponential decay parameter for the integral term with a
	// half-life of 1 second (goalTPS rounds of control):
	// 1. cumerr * decay^periods = rem*cumerr
	// 2. decay^periods = rem
	// 3. periods*log(decay) = log(rem)
	// 4. log(decay) = log(rem)/periods
	// 5. decay = exp(log(rem)/periods)
	return &Ticks{
		decay: math.Exp(math.Log(0.5) / float64(goalTPS)),
		state: TicksState{
			Goal: goalTPS,
		},
	}
}

// SetGoal changes the TPS goal.
func (ticks *Ticks) SetGoal(tps int) {
	ticks.state.Goal = tps
}

// TicksState represents the control state of Ticks.
type TicksState struct {
	// Goal ticks per second.
	Goal int

	// Goal interval between ticks.
	GoalDelta time.Duration

	// Current sleep setting for waiting between ticks.
	SetDelta time.Duration

	// Time of last tick period start; last call to Wait(), or time of
	// activation before first Wait().
	LastTime time.Time

	// Last observed delta between ticks (calls to Wait()).
	LastDelta time.Duration

	// Error control term (P).
	Err time.Duration

	// Integral control term (I).
	CumErr float64

	// Differential control term (D).
	DeltaDiff time.Duration

	// Last adjustment made.
	Adjust time.Duration
}

func (state TicksState) control(decay float64, now time.Time) TicksState {
	if state.LastTime.IsZero() {
		state.SetDelta = state.GoalDelta // reset directly to goal
	} else {
		prior := state.LastDelta                  // D term reference
		state.LastDelta = now.Sub(state.LastTime) // measure new delta

		state.Err = state.LastDelta - state.GoalDelta // P term
		adj := ticksFactorP * float64(state.Err)      // P adjust

		state.CumErr *= decay                             // I term decay
		state.CumErr += float64(state.Err)                // I term accumulate
		cumerr := state.CumErr                            //
		if math.Abs(cumerr) >= float64(state.GoalDelta) { // I term windup check
			cumerr = 0 // I term knockout; wait for it to decay back to usable
		}
		adj += ticksFactorI * cumerr // I adjust

		state.DeltaDiff = state.LastDelta - prior      // D term
		adj += ticksFactorD * float64(state.DeltaDiff) // D adjust

		state.Adjust = time.Duration(adj) // apply
		state.SetDelta -= state.Adjust    // adjustment

		if state.SetDelta < time.Millisecond { // clip
			state.SetDelta = time.Millisecond // set point
		}
	}
	state.LastTime = now // store time for next control round
	return state
}

// Ticks is essentially a dynamic time.Ticker using a PID control loop to
// ensure consistent timing between ticks.
type Ticks struct {
	timer *time.Timer
	state TicksState
	decay float64

	Metric TicksMetric
}

// Wait updates controller state, and then sleeps until the next tick
// time, which it then returns. Returns zero time once stopped. Users
// MUST store the returned time value, and pass it back to Wait() in
// the next round, for example:
//
//	ticks := anansi.NewTicks(60)
//	for now := time.Now(); !now.IsZero(); now = ticks.Wait(now) {
//		// TODO draw you a world
//	}
func (ticks *Ticks) Wait(lastWakeup time.Time) (now time.Time) {
	ticks.Metric.Wakeup = lastWakeup
	// TODO zero rest

	defer func() {
		ticks.Metric.Return = now
	}()

	if ticks.state.GoalDelta == 0 {
		ticks.reset()
		ticks.state.LastTime = lastWakeup
	} else {
		ticks.control(lastWakeup)
	}
	ticks.Metric.TicksState = ticks.state

	if ticks.state.SetDelta == 0 {
		return ticks.Metric.Return
	}

	delay := ticks.state.SetDelta
	ticks.Metric.Delay = delay

	now = time.Now()
	elapsed := now.Sub(lastWakeup)
	ticks.Metric.Elapsed = elapsed
	if elapsed >= delay {
		return now
	}
	delay -= elapsed
	// TODO try pivoting the set/goal-points to be correction terms added to
	// remaining time, rather than adjusting the set-point like this.

	if delay > 0 {
		if ticks.timer == nil {
			ticks.timer = time.NewTimer(delay)
		} else {
			ticks.timer.Reset(delay)
		}
	}

	now = <-ticks.timer.C
	if ticks.timer == nil {
		ticks.Metric.Return = time.Time{}
		return time.Time{}
	}
	return now
}

// Stop causes any current pending Wait() call to return zero time,
// which should break out of correctly written render loops. Any
// future call to Wait will reset controller state, allowing ticks to
// be reused. Also resets controller state.
func (ticks *Ticks) Stop() {
	ticks.timer = nil
	ticks.state = TicksState{
		Goal: ticks.state.Goal,
	}
}

// Enter resets controller state.
func (ticks *Ticks) Enter(term *anansi.Term) error {
	ticks.state = TicksState{
		Goal: ticks.state.Goal,
	}
	return nil
}

// Exit resets controller state.
func (ticks *Ticks) Exit(term *anansi.Term) error {
	ticks.state = TicksState{
		Goal: ticks.state.Goal,
	}
	return nil
}

func (ticks *Ticks) reset() {
	ticks.state = TicksState{
		Goal:      ticks.state.Goal,
		GoalDelta: time.Second / time.Duration(ticks.state.Goal),
		SetDelta:  time.Second / time.Duration(ticks.state.Goal),
		LastTime:  time.Now(),
	}
}

func (ticks *Ticks) control(t time.Time) {
	ticks.state = ticks.state.control(ticks.decay, t)
}
