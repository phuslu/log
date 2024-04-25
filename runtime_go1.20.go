//go:build go1.20 && !go1.21
// +build go1.20,!go1.21

// MIT license, copy and modify from https://github.com/tlog-dev/loc

//nolint:unused
package log

import (
	"unsafe"
)

// Fastrandn returns a pseudorandom uint32 in [0,n).
//
//go:noescape
//go:linkname Fastrandn runtime.fastrandn
func Fastrandn(x uint32) uint32

type funcInfo struct {
	_func *uintptr
	datap unsafe.Pointer
}

//go:linkname findfunc runtime.findfunc
func findfunc(pc uintptr) funcInfo

//go:linkname funcInfoEntry runtime.funcInfo.entry
func funcInfoEntry(f funcInfo) uintptr

//go:linkname funcline1 runtime.funcline1
func funcline1(f funcInfo, targetpc uintptr, strict bool) (file string, line int32)

func pcFileLine(pc uintptr) (file string, line int32) {
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

	return funcline1(funcInfo, pc, false)
}

// inlinedCall is the encoding of entries in the FUNCDATA_InlTree table.
type inlinedCall struct {
	funcID    uint8 // type of the called function
	_         [3]byte
	nameOff   int32 // offset into pclntab for name of called function
	parentPc  int32 // position of an instruction whose source position is the call site (offset from entry)
	startLine int32 // line number of start of function (func keyword/TEXT directive)
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

	name = funcname(funcInfo)
	file, line = funcline1(funcInfo, pc, false)

	const (
		_PCDATA_InlTreeIndex = 2
		_FUNCDATA_InlTree    = 3
	)

	if inldata := funcdata(funcInfo, _FUNCDATA_InlTree); inldata != nil {
		inltree := (*[1 << 20]inlinedCall)(inldata)
		// Non-strict as cgoTraceback may have added bogus PCs
		// with a valid funcInfo but invalid PCDATA.
		ix := pcdatavalue1(funcInfo, _PCDATA_InlTreeIndex, pc, nil, false)
		if ix >= 0 {
			// Note: entry is not modified. It always refers to a real frame, not an inlined one.
			ic := inltree[ix]
			name = funcnameFromNameOff(funcInfo, ic.nameOff)
			// File/line from funcline1 below are already correct.
		}
	}

	return
}

//go:linkname funcname runtime.funcname
func funcname(f funcInfo) string

//go:linkname funcdata runtime.funcdata
func funcdata(f funcInfo, i uint8) unsafe.Pointer

//go:linkname pcdatavalue runtime.pcdatavalue
func pcdatavalue(f funcInfo, table int32, targetpc uintptr, cache unsafe.Pointer) int32

//go:linkname pcdatavalue1 runtime.pcdatavalue1
func pcdatavalue1(f funcInfo, table int32, targetpc uintptr, cache unsafe.Pointer, strict bool) int32

//go:linkname funcnameFromNameOff runtime.funcnameFromNameOff
func funcnameFromNameOff(f funcInfo, nameoff int32) string
