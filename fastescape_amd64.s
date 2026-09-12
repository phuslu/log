//go:build gc

#include "textflag.h"

// func needEscapeBlocks(b string) bool
//   SI: b.base
//   BX: b.len, a positive multiple of 16
//
// Scans 16 bytes per iteration with SSE2 and reports whether any byte is one
// of the nine bytes appendEscapedString rewrites: 0x08, 0x09, 0x0a, 0x0c, 0x0d,
// '"', '\'', '<', '\\'. The range test is min(b-8, 5) == b-8, an exact
// unsigned 8 <= b <= 13 test, with 0x0b removed because appendEscapedString copies
// it verbatim while it does rewrite a 0x00 that a superset test would let
// through.
TEXT ·needEscapeBlocks(SB), NOSPLIT, $0-17
	MOVQ	b_base+0(FP), SI
	MOVQ	b_len+8(FP), BX

	MOVD	$0x08080808, AX
	MOVD	AX, X0
	PUNPCKLBW X0, X0
	PUNPCKLBW X0, X0		// 0x08
	MOVD	$0x05050505, AX
	MOVD	AX, X1
	PUNPCKLBW X1, X1
	PUNPCKLBW X1, X1		// 0x05
	MOVD	$0x22222222, AX
	MOVD	AX, X2
	PUNPCKLBW X2, X2
	PUNPCKLBW X2, X2		// '"'
	MOVD	$0x27272727, AX
	MOVD	AX, X3
	PUNPCKLBW X3, X3
	PUNPCKLBW X3, X3		// '\''
	MOVD	$0x3c3c3c3c, AX
	MOVD	AX, X4
	PUNPCKLBW X4, X4
	PUNPCKLBW X4, X4		// '<'
	MOVD	$0x5c5c5c5c, AX
	MOVD	AX, X5
	PUNPCKLBW X5, X5
	PUNPCKLBW X5, X5		// '\\'
	MOVD	$0x0b0b0b0b, AX
	MOVD	AX, X10
	PUNPCKLBW X10, X10
	PUNPCKLBW X10, X10		// 0x0b
	PXOR	X12, X12		// accumulator

	PCALIGN	$16
loop:
	MOVOU	(SI), X6
	MOVO	X6, X7
	PSUBB	X0, X7			// t = b - 0x08
	MOVO	X7, X8
	PMINUB	X1, X8			// min(t, 0x05)
	PCMPEQB	X7, X8			// 0x08 <= b <= 0x0d
	MOVO	X6, X11
	PCMPEQB	X10, X11		// b == 0x0b
	PANDN	X8, X11			// X11 = X8 &^ (b == 0x0b)
	MOVO	X6, X9
	PCMPEQB	X2, X9			// '"'
	POR	X9, X11
	MOVO	X6, X9
	PCMPEQB	X3, X9			// '\''
	POR	X9, X11
	MOVO	X6, X9
	PCMPEQB	X4, X9			// '<'
	POR	X9, X11
	PCMPEQB	X5, X6			// '\\'
	POR	X6, X11
	POR	X11, X12

	ADDQ	$16, SI
	SUBQ	$16, BX
	JNZ	loop

	MOVO	X12, X8
	PSHUFD	$0x4e, X8, X9		// swap the two 64-bit halves
	POR	X9, X8
	MOVQ	X8, AX
	TESTQ	AX, AX
	JNZ	found

	MOVB	$0, ret+16(FP)
	RET

found:
	MOVB	$1, ret+16(FP)
	RET
