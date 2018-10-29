package platform

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"syscall"
	"time"
)

// Telemetry contains platform runtime performance data.
type Telemetry struct {
	TelemetryState

	LastTick    TicksMetric
	FPSEstimate FPSEstimate
	Timing      TimingData
	Stalls      StallsData

	coll telemetryCollector
}

// TelemetryState contains serializable Telemetry state.
type TelemetryState struct {
	TimingEnabled    bool
	StallDataEnabled bool
	LogTicks         bool
	LogTiming        bool
	LogStallData     bool
}

// FPSEstimate keep a running Frames Per Second estimate based on a windowed
// average.
type FPSEstimate struct {
	i     int
	data  []float64
	Value float64
}

type rusageData struct {
	i    int
	data []rusageEntry
}

type rusageEntry struct {
	t time.Time
	syscall.Rusage
}

type rusageDelta struct {
	a, b *rusageEntry
}

func (dr rusageDelta) dt() time.Duration    { return dr.a.t.Sub(dr.b.t) }
func (dr rusageDelta) utime() time.Duration { return tvDuration(dr.a.Utime, dr.b.Utime) }
func (dr rusageDelta) stime() time.Duration { return tvDuration(dr.a.Stime, dr.b.Stime) }
func (dr rusageDelta) maxrss() int64        { return dr.a.Maxrss - dr.b.Maxrss }
func (dr rusageDelta) ixrss() int64         { return dr.a.Ixrss - dr.b.Ixrss }
func (dr rusageDelta) idrss() int64         { return dr.a.Idrss - dr.b.Idrss }
func (dr rusageDelta) isrss() int64         { return dr.a.Isrss - dr.b.Isrss }
func (dr rusageDelta) minflt() int64        { return dr.a.Minflt - dr.b.Minflt }
func (dr rusageDelta) majflt() int64        { return dr.a.Majflt - dr.b.Majflt }
func (dr rusageDelta) nswap() int64         { return dr.a.Nswap - dr.b.Nswap }
func (dr rusageDelta) inblock() int64       { return dr.a.Inblock - dr.b.Inblock }
func (dr rusageDelta) oublock() int64       { return dr.a.Oublock - dr.b.Oublock }
func (dr rusageDelta) msgsnd() int64        { return dr.a.Msgsnd - dr.b.Msgsnd }
func (dr rusageDelta) msgrcv() int64        { return dr.a.Msgrcv - dr.b.Msgrcv }
func (dr rusageDelta) nsignals() int64      { return dr.a.Nsignals - dr.b.Nsignals }
func (dr rusageDelta) nvcsw() int64         { return dr.a.Nvcsw - dr.b.Nvcsw }
func (dr rusageDelta) nivcsw() int64        { return dr.a.Nivcsw - dr.b.Nivcsw }

func tvDuration(a, b syscall.Timeval) time.Duration {
	sec := a.Sec - b.Sec
	usec := a.Usec - b.Usec
	return time.Duration(sec)*time.Second + time.Duration(usec)*time.Microsecond
}

func (ru *rusageData) computeDeltas(drs []rusageDelta, over ...time.Duration) []rusageDelta {
	if len(drs) > 0 {
		drs = drs[:0]
	}
	n := len(ru.data)
	j := ru.i - 1
	if j < 0 {
		j += n
	}
	latest := &ru.data[j]
	if latest.t.IsZero() {
		return drs
	}
	overi := 0
	var oldest *rusageEntry
	for k := 1; overi < len(over) && k < n; k++ {
		var ent *rusageEntry
		if i := j - k; i < 0 {
			ent = &ru.data[n+i]
		} else {
			ent = &ru.data[i]
		}
		if ent.t.IsZero() || latest.t.Sub(ent.t) > over[overi] {
			if oldest != nil && !oldest.t.IsZero() {
				drs = append(drs, rusageDelta{latest, oldest})
			}
			overi++
		}
		oldest = ent
	}
	return drs
}

func (ru *rusageData) update(p *Platform) error {
	if len(ru.data) == 0 {
		return nil
	}
	ent := &ru.data[ru.i]
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &ent.Rusage)
	if err == nil {
		ent.t = p.Time
		ru.i = (ru.i + 1) % len(ru.data)
	}
	return err
}

// TimingData stores inter-frame timing data.
type TimingData struct {
	// state
	Stats TimingStats

	// collection
	i  int
	ts []time.Time
	ds []time.Duration
}

// TimingStats stores stats computed from the last round of frame TimingData.
type TimingStats struct {
	Time time.Time

	FPS        float64
	Elapsed    time.Duration
	Min, Max   time.Duration
	Q1, Q2, Q3 time.Duration
}

// StallsData stores output stall data.
type StallsData struct {
	Stats StallsStats
}

// StallsStats stores stats computed from the last round of output StallsData.
type StallsStats struct {
	N    int
	Time time.Time
	Min  time.Duration
	Max  time.Duration
	Sum  time.Duration
	Pct  float64
}

// FPS returns the measured FPS rate if timing collection is enabled, or the
// current FPSEstimate value otherwise.
func (tel *Telemetry) FPS() float64 {
	if tel.TimingEnabled || tel.LogTiming {
		return tel.Timing.Stats.FPS
	}
	return tel.FPSEstimate.Value
}

// SetTimingEnabled sets whether frame timing data collection is enabled.
func (p *Platform) SetTimingEnabled(enabled bool) {
	if p.TimingEnabled != enabled {
		p.TimingEnabled = enabled
		if !p.TimingEnabled && !p.LogTiming {
			p.Timing.reset()
		}
	}
}

// SetStallTracking sets whether output stall tracking is enabled.
func (p *Platform) SetStallTracking(enabled bool) {
	p.StallDataEnabled = enabled
	if p.StallDataEnabled {
		p.output.TrackStalls(len(p.Timing.ts))
	} else {
		p.output.TrackStalls(0)
		p.Stalls.reset()
	}
}

func (td *TimingData) reset() {
	td.i = 0
	for i := range td.ts {
		td.ts[i] = time.Time{}
	}
	for i := range td.ds {
		td.ds[i] = 0
	}
	td.Stats = TimingStats{}
}

func (sd *StallsData) reset() {
	sd.Stats = StallsStats{}
}

func (tel *Telemetry) update(p *Platform) error {
	tel.coll.Lock()
	defer tel.coll.Unlock()
	tel.coll.t = p.Time

	if err := tel.coll.rusage.update(p); err != nil {
		return err
	}

	tel.LastTick = p.ticks.Metric
	if tel.LogTicks {
		tel.coll.tick = &tel.LastTick
	}

	tel.FPSEstimate.update(p)
	if tel.TimingEnabled || tel.LogTiming {
		timingFrame := tel.Timing.update(p)
		consumeStalls := timingFrame
		if tel.LogTiming && timingFrame {
			tel.coll.timing = tel.Timing.ds
		}
		if stalls := p.output.Stalls(consumeStalls); stalls != nil {
			if tel.LogStallData {
				tel.coll.stalls = stalls
			}
			if tel.StallDataEnabled {
				tel.Stalls.update(p, stalls)
			}
		}
	}
	return nil
}

func (td *TimingData) update(p *Platform) bool {
	td.ts[td.i] = p.Time
	td.i = (td.i + 1) % len(td.ts)
	if td.i > 0 {
		return false
	}

	// first pass: deltas and simple stats
	stats := TimingStats{
		Time: p.Time,
	}
	for i := 0; i < len(td.ts); i++ {
		t := td.ts[i]
		d := t.Sub(td.Stats.Time)
		if td.Stats.Time.IsZero() {
			d = 0
		}
		td.ds[i] = d
		td.Stats.Time = t

		if stats.Elapsed == 0 {
			stats.Min, stats.Max, stats.Elapsed = d, d, d
		} else {
			stats.Elapsed += d
			if stats.Min > d {
				stats.Min = d
			}
			if stats.Max < d {
				stats.Max = d
			}
		}
	}

	stats.FPS = float64(len(td.ds)) / float64(stats.Elapsed) * float64(time.Second)

	// second pass: quantiles (TODO worth it to use something "better"?)
	sort.Sort(durations(td.ds))
	q := len(td.ds) / 4
	stats.Q1, stats.Q2, stats.Q3 = td.ds[1*q], td.ds[2*q], td.ds[3*q]

	td.Stats = stats

	return true
}

func (sd *StallsData) update(p *Platform, stalls []time.Duration) {
	stats := StallsStats{
		N:    len(stalls),
		Time: p.Time,
	}
	if len(stalls) > 0 {
		stats.Min = stalls[0]
		stats.Max = stalls[0]
		stats.Sum = stalls[0]
		for _, stall := range stalls[1:] {
			if stats.Min > stall {
				stats.Min = stall
			}
			if stats.Max < stall {
				stats.Max = stall
			}
			stats.Sum += stall
		}
	}
	if !sd.Stats.Time.IsZero() {
		stats.Pct = float64(stats.Sum) / float64(stats.Time.Sub(sd.Stats.Time))
	}
	sd.Stats = stats
}

func (fe *FPSEstimate) update(p *Platform) {
	delta := p.Time.Sub(p.LastTime)
	fe.data[fe.i] = float64(time.Second) / float64(delta)
	fe.i = (fe.i + 1) % len(fe.data)
	var est float64
	for _, d := range fe.data {
		est += d / float64(len(fe.data))
	}
	fe.Value = est
}

type durations []time.Duration

func (ds durations) Len() int           { return len(ds) }
func (ds durations) Less(i, j int) bool { return ds[i] < ds[j] }
func (ds durations) Swap(i int, j int)  { ds[i], ds[j] = ds[j], ds[i] }

type telemetryCollector struct {
	bgWorkerCore
	t      time.Time
	tick   *TicksMetric
	timing []time.Duration
	stalls []time.Duration

	lastRusageLog time.Time
	rusage        rusageData
	drs           []rusageDelta

	buf bytes.Buffer
	f   *os.File
}

func (coll *telemetryCollector) name() string            { return "Telemetry Log" }
func (coll *telemetryCollector) defaultFileName() string { return "telemetry.log" }
func (coll *telemetryCollector) file() *os.File          { return coll.f }
func (coll *telemetryCollector) create(fn string) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	coll.Lock()
	defer coll.Unlock()
	if coll.f != nil {
		if err := coll.f.Close(); err != nil {
			_ = f.Close()
			return err
		}
	}
	coll.f = f
	log.Printf("logging telemetry to %q", f.Name())
	return nil
}

func (coll *telemetryCollector) Start() error {
	if err := coll.bgWorkerCore.Start(); err != nil {
		return err
	}
	go coll.worker()
	coll.buf.Grow(64 * 1024)
	return nil
}

func (coll *telemetryCollector) worker() {
	defer close(coll.e)
	for open := true; open; {
		_, open = <-coll.w
		coll.logData()
	}
}

func (coll *telemetryCollector) logData() {
	coll.Lock()
	defer coll.Unlock()

	if coll.f != nil {
		coll.buf.Reset()
		t := coll.t.UnixNano()

		if coll.t.Sub(coll.lastRusageLog) >= time.Second {
			coll.lastRusageLog = coll.t
			if coll.drs = coll.rusage.computeDeltas(coll.drs,
				1*time.Second,
				2*time.Second,
				3*time.Second,
				4*time.Second,
				5*time.Second,
				6*time.Second,
				7*time.Second,
				8*time.Second,
			); len(coll.drs) > 0 {
				t := coll.t.UnixNano()
				fmt.Fprintf(&coll.buf, `{"t":%d,"name":"rusage_delta","data":[`, t)
				for i, dr := range coll.drs {
					if i > 0 {
						coll.buf.WriteByte(',')
					}
					fmt.Fprintf(&coll.buf,
						`{"dt":%d,"utime":%d,"stime":%d}`,
						dr.dt(),
						dr.utime(),
						dr.stime(),
						// TODO more?
						// dr.minflt()
						// dr.majflt()
						// dr.nvcsw()
						// dr.nivcsw()
					)
				}
				coll.buf.WriteString("]}\n")
			}
		}

		if coll.tick != nil {
			fmt.Fprintf(&coll.buf, `{"t":%d,"name":"tick_metric","data":`, t)
			coll.tick.WriteToBuffer(&coll.buf)
			coll.buf.WriteString("}\n")
			coll.tick = nil
		}

		if coll.timing != nil {
			fmt.Fprintf(&coll.buf, `{"t":%d,"name":"timing","data":`, t)
			appendDurationsTo(&coll.buf, coll.timing)
			coll.buf.WriteString("}\n")
			coll.timing = nil
		}

		if coll.stalls != nil {
			fmt.Fprintf(&coll.buf, `{"t":%d,"name":"stalls","stalls":`, t)
			appendDurationsTo(&coll.buf, coll.stalls)
			coll.buf.WriteString("}\n")
			coll.stalls = nil
		}

		if _, err := coll.buf.WriteTo(coll.f); err != nil {
			coll.e <- err
		}
		return
	}

	if coll.tick != nil {
		log.Printf("tick metric: %v", coll.tick)
		coll.tick = nil
	}
	if coll.timing != nil {
		log.Printf("timing data: %v", coll.timing)
		coll.timing = nil
	}
	if coll.stalls != nil {
		log.Printf("output stalls %v", coll.stalls)
		coll.stalls = nil
	}
}

func appendDurationsTo(buf *bytes.Buffer, ds []time.Duration) {
	fmt.Fprintf(buf, `[`)
	if len(ds) > 0 {
		fmt.Fprintf(buf, `%d`, ds[0])
		for i := 1; i < len(ds); i++ {
			fmt.Fprintf(buf, `,%d`, ds[i])
		}
	}
	buf.WriteString("]")
}
