//go:build go1.21 && !go1.22
// +build go1.21,!go1.22

// MIT license, copy and modify from https://github.com/tlog-dev/loc

//nolint:unused
package log

import (
	"unsafe"
)

// inlinedCall is the encoding of entries in the FUNCDATA_InlTree table.
type inlinedCall struct {
	funcID    uint8 // type of the called function
	_         [3]byte
	nameOff   int32 // offset into pclntab for name of called function
	parentPc  int32 // position of an instruction whose source position is the call site (offset from entry)
	startLine int32 // line number of start of function (func keyword/TEXT directive)
}

type inlineUnwinder struct {
	f       funcInfo
	cache   unsafe.Pointer
	inlTree *[1 << 20]inlinedCall
}

type inlineFrame struct {
	pc    uintptr
	index int32
}

type srcFunc struct {
	datap     unsafe.Pointer
	nameOff   int32
	startLine int32
	funcID    uint8
}

func pcNameFileLine(pc uintptr) (name, file string, line int32) {
	funcInfo := findfunc(pc)
	if funcInfo._func == nil {
		return
	}

	entry := funcInfoEntry(funcInfo)

	if pc > entry {
		// We store the pc of the start of the instruction following
		// the instruction in question (the call or the inline mark).
		// This is done for historical reasons, and to make FuncForPC
		// work correctly for entries in the result of runtime.Callers.
		pc--
	}

	file, line = funcline1(funcInfo, pc, false)

	// It's important that interpret pc non-strictly as cgoTraceback may
	// have added bogus PCs with a valid funcInfo but invalid PCDATA.
	u, uf := newInlineUnwinder(funcInfo, pc, nil)
	sf := inlineUnwinder_srcFunc(&u, uf)
	name = srcFunc_name(sf)
	// name = funcNameForPrint(srcFunc_name(sf))

	return
}

//go:linkname newInlineUnwinder runtime.newInlineUnwinder
func newInlineUnwinder(f funcInfo, pc uintptr, cache unsafe.Pointer) (inlineUnwinder, inlineFrame)

//go:linkname inlineUnwinder_srcFunc runtime.(*inlineUnwinder).srcFunc
func inlineUnwinder_srcFunc(*inlineUnwinder, inlineFrame) srcFunc

//go:linkname inlineUnwinder_isInlined runtime.(*inlineUnwinder).isInlined
func inlineUnwinder_isInlined(*inlineUnwinder, inlineFrame) bool

//go:linkname srcFunc_name runtime.srcFunc.name
func srcFunc_name(srcFunc) string

//go:linkname funcNameForPrint runtime.funcNameForPrint
func funcNameForPrint(name string) string

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(x uint32) uint32
