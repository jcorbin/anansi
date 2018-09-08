package platform

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/jcorbin/anansi"
)

// Config uration for a Platform; populated from -platform.* flags; it
// implements Option, so applications may unmarshal it from some file and pass
// it to New().
//
// TODO hud manipulable / dynamic
type Config struct {
	LogFileName    string
	CPUProfileName string
	TraceFileName  string
	MemProfileName string
	// TODO config for arbitrary pprof profiles

	StartTiming bool // Whether to start and
	LogTiming   bool // log timing right away

	logFile       *os.File
	cpuProfile    cpuProfileContext
	traceProfile  traceProfileContext
	pprofProfiles []pprofProfileContext
}

// AddFlags adds flag variables to the given flag set that will allow the user
// to set various log and profiling files from the outset. If a non-empty
// prefix is given, then timing variables are also wired up.
func (cfg *Config) AddFlags(f *flag.FlagSet, prefix string) {
	f.StringVar(&cfg.LogFileName, prefix+"logfile", cfg.LogFileName,
		"write logs to a file")
	f.StringVar(&cfg.CPUProfileName, prefix+"cpuprofile", cfg.CPUProfileName,
		"enables platform cpu profiling")
	f.StringVar(&cfg.MemProfileName, prefix+"memprofile", cfg.MemProfileName,
		"enables platform memory profiling")
	f.StringVar(&cfg.TraceFileName, prefix+"tracefile", cfg.TraceFileName,
		"enables platform execution tracing")
	if prefix != "" {
		f.BoolVar(&cfg.StartTiming, prefix+"timing", false,
			"measure timing from the beginning")
		f.BoolVar(&cfg.LogTiming, prefix+"timing.log", false,
			"measure and log timing from the beginning")
	}
}

func (cfg *Config) apply(p *Platform) error {
	p.Config.Merge(*cfg)
	return nil
}

// Merge an other config value into the receiver.
func (cfg *Config) Merge(other Config) {
	if other.LogFileName != "" && cfg.LogFileName == "" {
		cfg.LogFileName = other.LogFileName
	}
	if other.CPUProfileName != "" && cfg.CPUProfileName == "" {
		cfg.CPUProfileName = other.CPUProfileName
	}
	if other.TraceFileName != "" && cfg.TraceFileName == "" {
		cfg.TraceFileName = other.TraceFileName
	}
	if other.MemProfileName != "" && cfg.MemProfileName == "" {
		cfg.MemProfileName = other.MemProfileName
	}
	if other.LogTiming {
		cfg.LogTiming = true
	}
}

func (cfg *Config) setup(p *Platform) error {
	p.SetTimingEnabled(cfg.LogTiming)

	if cfg.LogFileName != "" && p.logFile == nil {
		f, err := os.Create(cfg.LogFileName)
		if err != nil {
			return fmt.Errorf("failed to open log file %q: %v", cfg.LogFileName, err)
		}
		log.SetOutput(io.MultiWriter(&Logs, f))
		log.Printf("logging to %q", f.Name())
	}

	if cfg.CPUProfileName != "" {
		f, err := os.Create(cfg.CPUProfileName)
		if err != nil {
			return fmt.Errorf("failed to create %q: %v", cfg.CPUProfileName, err)
		}
		p.cpuProfile.f = f
	}

	if cfg.TraceFileName != "" {
		f, err := os.Create(cfg.TraceFileName)
		if err != nil {
			return fmt.Errorf("failed to create %q: %v", cfg.TraceFileName, err)
		}
		p.traceProfile.f = f
	}

	if cfg.MemProfileName != "" {
		f, err := os.Create(cfg.MemProfileName)
		if err != nil {
			return fmt.Errorf("failed to create %q: %v", cfg.MemProfileName, err)
		}
		p.pprofProfiles = append(p.pprofProfiles, pprofProfileContext{
			profile: pprof.Lookup("heap"),
			f:       f,
			debug:   1,
		})
	}

	return nil
}

// Enter starts any CPU or Trace profiling.
func (cfg *Config) Enter(term *anansi.Term) error {
	if err := cfg.cpuProfile.Enter(term); err != nil {
		return err
	}
	if err := cfg.traceProfile.Enter(term); err != nil {
		return err
	}
	for i := 0; i < len(cfg.pprofProfiles); i++ {
		if err := cfg.pprofProfiles[i].Enter(term); err != nil {
			return err
		}
	}
	return nil
}

// Exit stops any CPU or Trace profiling, and writes any configured pprof
// profiles (like memory).
func (cfg *Config) Exit(term *anansi.Term) error {
	var err error
	for i := len(cfg.pprofProfiles) - 1; i >= 0; i-- {
		err = errOr(err, cfg.pprofProfiles[i].Exit(term))
	}
	err = errOr(err, cfg.traceProfile.Exit(term))
	err = errOr(err, cfg.cpuProfile.Exit(term))
	return err
}

type cpuProfileContext struct {
	active bool
	f      *os.File
}

type traceProfileContext struct {
	active bool
	f      *os.File
}

type pprofProfileContext struct {
	active  bool
	profile *pprof.Profile
	f       *os.File
	debug   int
}

func (cpu *cpuProfileContext) defaultFileName() string    { return "prof.cpu" }
func (trc *traceProfileContext) defaultFileName() string  { return "prof.trace" }
func (prof *pprofProfileContext) defaultFileName() string { return "prof." + prof.profile.Name() }

func (cpu *cpuProfileContext) name() string    { return "CPU profiler" }
func (trc *traceProfileContext) name() string  { return "Trace profiler" }
func (prof *pprofProfileContext) name() string { return prof.profile.Name() + " profiler" }

func (cpu *cpuProfileContext) file() *os.File    { return cpu.f }
func (trc *traceProfileContext) file() *os.File  { return trc.f }
func (prof *pprofProfileContext) file() *os.File { return prof.f }

func (cpu *cpuProfileContext) isActive() bool    { return cpu.active }
func (trc *traceProfileContext) isActive() bool  { return trc.active }
func (prof *pprofProfileContext) isActive() bool { return prof.active }

func (cpu *cpuProfileContext) create(name string) error {
	if cpu.active {
		if err := cpu.stop(); err != nil {
			return err
		}
	}
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if cpu.f != nil {
		_ = cpu.f.Close()
	}
	cpu.f = f
	return cpu.start()
}

func (trc *traceProfileContext) create(name string) error {
	if trc.active {
		if err := trc.stop(); err != nil {
			return err
		}
	}
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if trc.f != nil {
		_ = trc.f.Close()
	}
	trc.f = f
	return trc.start()
}

func (prof *pprofProfileContext) create(name string) error {
	if prof.active {
		if err := prof.stop(); err != nil {
			return err
		}
	}
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	if prof.f != nil {
		_ = prof.f.Close()
	}
	prof.f = f
	return prof.start()
}

func (cpu *cpuProfileContext) Enter(_ *anansi.Term) error {
	if cpu.f == nil {
		return nil
	}
	return cpu.start()
}

func (cpu *cpuProfileContext) Exit(_ *anansi.Term) error {
	if cpu.f == nil {
		return nil
	}
	return cpu.stop()
}

func (trc *traceProfileContext) Enter(_ *anansi.Term) error {
	if trc.f == nil {
		return nil
	}
	return trc.start()
}

func (trc *traceProfileContext) Exit(_ *anansi.Term) (err error) {
	if trc.f == nil {
		return nil
	}
	return trc.stop()
}

func (prof *pprofProfileContext) Enter(_ *anansi.Term) error {
	if prof.f == nil {
		return nil
	}
	return prof.start()
}

func (prof *pprofProfileContext) Exit(_ *anansi.Term) error {
	if prof.f == nil {
		return nil
	}
	return prof.stop()
}

func (cpu *cpuProfileContext) start() error {
	if cpu.f == nil {
		f, err := os.Create(cpu.defaultFileName())
		if err != nil {
			return err
		}
		cpu.f = f
	}
	if cpu.active {
		return nil
	}
	_, err := cpu.f.Seek(0, io.SeekStart)
	if err == nil {
		err = cpu.f.Truncate(0)
	}
	if err != nil {
		return err
	}
	if err = pprof.StartCPUProfile(cpu.f); err != nil {
		return fmt.Errorf("failed to start CPU profiling: %v", err)
	}
	cpu.active = true
	log.Printf("CPU profiling to %q", cpu.f.Name())
	return nil
}

func (cpu *cpuProfileContext) stop() error {
	if !cpu.active {
		return nil
	}
	pprof.StopCPUProfile()
	if err := cpu.f.Sync(); err != nil {
		return err
	}
	cpu.active = false
	log.Printf("Flushed CPU profiling to %q", cpu.f.Name())
	return nil
}

func (trc *traceProfileContext) start() error {
	if trc.f == nil {
		f, err := os.Create(trc.defaultFileName())
		if err != nil {
			return err
		}
		trc.f = f
	}
	if trc.active {
		return nil
	}
	_, err := trc.f.Seek(0, io.SeekStart)
	if err == nil {
		err = trc.f.Truncate(0)
	}
	if err != nil {
		return err
	}
	if err := trace.Start(trc.f); err != nil {
		return fmt.Errorf("failed to start execution tracing: %v", err)
	}
	trc.active = true
	log.Printf("Tracing execution to %q", trc.f.Name())
	return nil
}

func (trc *traceProfileContext) stop() (err error) {
	if !trc.active {
		return nil
	}
	trace.Stop()
	if err := trc.f.Sync(); err != nil {
		return err
	}
	trc.active = false
	log.Printf("Flushed execution trace execution to %q", trc.f.Name())
	return nil
}

func (prof *pprofProfileContext) start() error {
	if prof.f == nil {
		f, err := os.Create(prof.defaultFileName())
		if err != nil {
			return err
		}
		prof.f = f
	}
	prof.active = true
	return nil
}

func (prof *pprofProfileContext) stop() error {
	if !prof.active {
		return nil
	}
	err := prof.take()
	prof.active = false
	return err
}

func (prof *pprofProfileContext) take() error {
	_, err := prof.f.Seek(0, io.SeekStart)
	if err == nil {
		err = prof.f.Truncate(0)
		if err == nil {
			err = prof.profile.WriteTo(prof.f, prof.debug)
			if err == nil {
				err = prof.f.Sync()
				if err == nil {
					log.Printf("Wrote %v profile to %q", prof.profile.Name(), prof.f.Name())
				}
			}
		}
	}
	return err
}
