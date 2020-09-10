package log

import (
	"io"
	"runtime"
	"sync/atomic"
)

type RingWriter struct {
	w io.Writer
	r *ring
}

// NewRingWriter return a RingWriter
func NewRingWriter(w io.Writer, size uint32, batch uint32) (rw *RingWriter) {
	rw = new(RingWriter)
	rw.w = w
	rw.r = newRing(size, batch)
	go func(w io.Writer, r *ring) {
		r.Consume(func(it *iter, b *[]byte) {
			w.Write(*b)
		})
	}(w, rw.r)
	return
}

// Close implements io.Closer, and closes the underlying Writer.
func (rw *RingWriter) Close() (err error) {
	rw.r.Close()
	if closer, ok := rw.w.(io.Closer); ok {
		err = closer.Close()
	}
	return
}

// Write implements io.Writer.
func (rw *RingWriter) Write(p []byte) (n int, err error) {
	rw.r.Put(p)
	return len(p), nil
}

type ring struct {
	_        int64
	_        [7]int64
	wp       int64
	_        [7]int64
	rp       int64
	_        [7]int64
	rc       int64 // reader cache
	_        [7]int64
	data     [][]byte
	mask     int64
	size     int64
	maxbatch int64
	done     int32
	_        [42]byte
	seq      []int64
}

func newRing(size uint32, batch uint32) (r *ring) {
	r = new(ring)

	r.data = make([][]byte, roundUp2(size))
	r.mask = int64(len(r.data) - 1)

	if batch == 0 {
		batch = 255
	}
	r.maxbatch = int64(roundUp2(batch) - 1)

	r.size = int64(len(r.data))
	r.seq = make([]int64, len(r.data))
	for i := range r.seq {
		r.seq[i] = int64(i)
	}
	r.wp = 1 // just to avoid 0-awkwardness with seq
	r.rp = 1
	r.rc = r.rp

	return r
}

func (r *ring) Close() {
	atomic.AddInt32(&r.done, 1)
}

func (r *ring) Done() bool {
	return atomic.LoadInt64(&r.wp) <= atomic.LoadInt64(&r.rp) && atomic.LoadInt32(&r.done) > 0
}

func (r *ring) Get(i *[]byte) bool {
	rc := r.rc
	pos := r.mask & rc
	data, seq := &r.data[pos], &r.seq[pos]

	if sv := atomic.LoadInt64(seq); rc > sv {
		if rc > r.rp {
			atomic.StoreInt64(&r.rp, rc)
		}
		for ; rc > sv; sv = atomic.LoadInt64(seq) {
			if r.Done() {
				return false
			}
			runtime.Gosched()
		}
	}

	*i = *data

	*seq = -rc
	rc++
	r.rc = rc
	if r.rc-r.rp > r.maxbatch {
		atomic.StoreInt64(&r.rp, rc)
	}
	return true
}

func (r *ring) Consume(fn func(it *iter, b *[]byte)) {
	var maxbatch = int(r.maxbatch)
	var it iter
	for keep := true; keep; {
		var rc, wp = r.rc, atomic.LoadInt64(&r.wp)
		for ; rc >= wp; wp = atomic.LoadInt64(&r.wp) {
			if atomic.LoadInt32(&r.done) > 0 {
				return
			}
			runtime.Gosched()
		}

		for i := 0; rc < wp && keep; it.inc() {
			pos := r.mask & rc
			data, seq := &r.data[pos], &r.seq[pos]
			if i++; atomic.LoadInt64(seq) <= 0 || i&maxbatch == 0 {
				r.rc = rc
				atomic.StoreInt64(&r.rp, rc)
				for atomic.LoadInt64(seq) <= 0 {
					runtime.Gosched()
				}
			}
			fn(&it, data)
			*seq = -rc
			keep = !it.stop
			rc++
		}
		r.rc = rc
		atomic.StoreInt64(&r.rp, r.rc)
	}
}

func (r *ring) Put(b []byte) {
	var wp = atomic.AddInt64(&r.wp, 1) - 1
	for diff := wp - r.mask; diff >= atomic.LoadInt64(&r.rp); {
		runtime.Gosched()
	}
	var pos = wp & r.mask
	r.data[pos] = b
	atomic.StoreInt64(&r.seq[pos], wp)
}

type iter struct {
	count int
	stop  bool
}

func (i *iter) Stop() {
	i.stop = true
}

func (i *iter) Count() int {
	return i.count
}

func (i *iter) inc() {
	i.count++
}

func roundUp2(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	return v + 1
}
