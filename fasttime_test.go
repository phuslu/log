//go:build gc && linux && (amd64 || arm64) && go1.25 && !go1.28

package log

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestFastClockValue(t *testing.T) {
	checkFastClockValue(t)
}

func checkFastClockValue(t *testing.T) {
	t.Helper()
	real := time.Now()
	sec, nsec := walltime()
	if nsec < 0 || nsec >= 1e9 {
		t.Fatalf("walltime returned invalid nanoseconds: %d", nsec)
	}
	got := time.Unix(sec, int64(nsec))
	if d := real.Sub(got); d < -time.Second || d > time.Second {
		t.Fatalf("walltime returned %v, %v off from %v", got, d, real)
	}
}

// checkFastClockFallback logs one line through the built-in header fast path
// and checks that the emitted time is the current wall clock. With walltime
// returning zero the header has to reach now() itself, so this is what keeps
// the fallback honest on platforms without a vDSO fast path.
func checkFastClockFallback(t *testing.T) {
	t.Helper()
	if sec, nsec := walltime(); sec != 0 || nsec != 0 {
		t.Fatalf("walltime returned %v, %v with the fast path disabled; want 0, 0", sec, nsec)
	}
	var buf bytes.Buffer
	logger := Logger{TimeFormat: time.RFC3339Nano, TimeLocation: time.UTC, Writer: IOWriter{&buf}}
	real := time.Now()
	logger.Info().Msg("x")
	var m map[string]string
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m); err != nil {
		t.Fatalf("bad json %q: %v", buf.String(), err)
	}
	got, err := time.Parse(time.RFC3339Nano, m[TimeKey])
	if err != nil {
		t.Fatalf("header time %q is not a valid timestamp: %v", m[TimeKey], err)
	}
	if d := real.Sub(got); d < -time.Second || d > time.Second {
		t.Fatalf("header time returned %v, %v off from %v", got, d, real)
	}
}

func TestFastClockG0Layout(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if vdsoClockgettimeSym == 0 {
		t.Skip("no vDSO clock_gettime on this machine")
	}
	if g0LayoutOK(runtimeGetm()) {
		if !vdsoReady {
			t.Fatal("g/m offsets match but the vDSO fast path is disabled")
		}
		return
	}
	if vdsoReady {
		t.Fatal("vDSO fast path enabled with incompatible g/m offsets")
	}
	checkFastClockFallback(t)
	t.Skip("g/m offsets do not match this toolchain; now() fallback verified")
}

func TestFastClockG0LayoutRejectsWrongM(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	m := runtimeGetm()
	for _, bogus := range []uintptr{0, 1, m + 8} {
		if g0LayoutOK(bogus) {
			t.Fatalf("g0LayoutOK(%#x) accepted a wrong m; got %#x", bogus, m)
		}
	}
}

func TestFastClockFallback(t *testing.T) {
	ready := vdsoReady
	vdsoReady = false
	t.Cleanup(func() { vdsoReady = ready })
	checkFastClockFallback(t)
}

// TestFastClockSignalStress hammers walltime from many goroutines while
// SIGPROF (cpu profiling) and a SIGWINCH storm are being delivered, so signals
// land inside the vDSO call while running on the g0 stack. This is the
// scenario the arm64 gsignal stash exists for (golang.org/issue/32912): a
// mishandled signal there crashes the process or corrupts the clock reads.
func TestFastClockSignalStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping signal stress test in short mode")
	}
	if !vdsoReady {
		t.Skip("vDSO fast path disabled")
	}

	// Errors only mean profiling is already running; the storm still applies.
	if err := pprof.StartCPUProfile(io.Discard); err == nil {
		defer pprof.StopCPUProfile()
	}

	ch := make(chan os.Signal, 128)
	signal.Notify(ch, syscall.SIGWINCH)
	defer signal.Stop(ch)
	go func() {
		for range ch {
		}
	}()

	stop := make(chan struct{})
	var storm sync.WaitGroup
	storm.Add(1)
	go func() {
		defer storm.Done()
		pid := syscall.Getpid()
		for {
			select {
			case <-stop:
				return
			default:
				syscall.Kill(pid, syscall.SIGWINCH)
			}
		}
	}()

	startSec := time.Now().Unix()
	var bad atomic.Int64
	var wg sync.WaitGroup
	deadline := time.Now().Add(2 * time.Second)
	for i := 0; i < 2*runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Now().Before(deadline) {
				for j := 0; j < 4096; j++ {
					sec, nsec := walltime()
					if nsec < 0 || nsec >= 1e9 || sec < startSec-60 || sec > startSec+60 {
						bad.Add(1)
					}
				}
			}
		}()
	}
	wg.Wait()
	close(stop)
	storm.Wait()

	if n := bad.Load(); n != 0 {
		t.Fatalf("walltime returned %d invalid values under signal stress", n)
	}
}

func TestFastClockCallUnavailable(t *testing.T) {
	want := timespec{sec: 123, nsec: 456}
	got := want
	if errno := vdsoCallG0(0, &got); errno == 0 {
		t.Fatal("vdsoCallG0 succeeded with no vDSO function")
	}
	if got != want {
		t.Fatalf("vdsoCallG0 changed timespec on failure: got %+v, want %+v", got, want)
	}
}

// walltimeSecSink, walltimeNsecSink and unixSecSink keep the compiler from
// discarding the clock reads the benchmark is measuring.
var (
	walltimeSecSink  int64
	walltimeNsecSink int32
	unixSecSink      int64
)

// BenchmarkWalltime compares walltime(), the single CLOCK_REALTIME read the
// built-in header uses, with a plain time.Now().Unix(), which reads
// CLOCK_REALTIME and CLOCK_MONOTONIC and then unpacks a time.Time. Platforms
// taking fasttime_zzz.go instead have no vDSO fast path to benchmark, and the
// header there falls back to now().
func BenchmarkWalltime(b *testing.B) {
	b.Run("walltime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			walltimeSecSink, walltimeNsecSink = walltime()
		}
	})

	b.Run("time.Now().Unix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			unixSecSink = time.Now().Unix()
		}
	})
}
