// +build !windows
// +build !linux

package log

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()

	file := w.File()

	if file == nil {
		err = w.rotate(false)
		if err != nil {
			w.mu.Unlock()
			return
		}
	}

	n, err = file.Write(p)

	w.size += int64(n)
	if w.MaxSize > 0 && w.size > w.MaxSize {
		w.rotate(true)
	}

	w.mu.Unlock()
	return
}
