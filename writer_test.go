package log

import (
	"fmt"
	"testing"
)

func TestWriter(t *testing.T) {
	w := &Writer{}
	fmt.Fprintf(w, "hello writer!\n")
	w.Rotate()
	w.Close()
}
