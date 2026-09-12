//go:build gc

#include "textflag.h"

// The two nibble tables below encode the nine bytes appendEscapedString rewrites:
// 0x08, 0x09, 0x0a, 0x0c, 0x0d, '"', '\'', '<', '\\'. Bit i of a table entry
// marks a byte, and a byte b is one of them exactly when
// loTable[b&0xf] & hiTable[b>>4] != 0. 0x0c, 0x3c and 0x5c share bit 3 because
// their cross combinations are themselves in the set, so there are no false
// positives, and 0x0b stays out because appendEscapedString copies it verbatim
// while it does rewrite the 0x00 such a superset test would let through.
DATA	·escapeLoTable+0(SB)/8, $0x4000000000200000
DATA	·escapeLoTable+8(SB)/8, $0x0000100800040201
GLOBL	·escapeLoTable(SB), RODATA|NOPTR, $16

DATA	·escapeHiTable+0(SB)/8, $0x000008000860001f
DATA	·escapeHiTable+8(SB)/8, $0x0000000000000000
GLOBL	·escapeHiTable(SB), RODATA|NOPTR, $16

// func needEscapeBlocks(b string) bool
//   R0: b.base
//   R1: b.len, a positive multiple of 16
//
// Scans 16 bytes per iteration with NEON: one nibble extraction, two table
// lookups and one AND decide whether any byte in the block is in the set.
TEXT ·needEscapeBlocks(SB), NOSPLIT, $0-17
	MOVD	b_base+0(FP), R0
	MOVD	b_len+8(FP), R1

	MOVD	$·escapeLoTable(SB), R10
	VLD1	(R10), [V18.B16]
	MOVD	$·escapeHiTable(SB), R11
	VLD1	(R11), [V19.B16]
	MOVD	$0x0f0f0f0f0f0f0f0f, R12
	VMOV	R12, V17.B16		// 0x0f
	VEOR	V20.B16, V20.B16, V20.B16 // accumulator

loop:
	VLD1.P	16(R0), [V6.B16]
	VAND	V17.B16, V6.B16, V7.B16	// low nibbles
	VUSHR	$4, V6.B16, V8.B16	// high nibbles
	VTBL	V7.B16, [V18.B16], V9.B16
	VTBL	V8.B16, [V19.B16], V10.B16
	VAND	V10.B16, V9.B16, V9.B16
	VORR	V9.B16, V20.B16, V20.B16

	SUBS	$16, R1, R1
	BGT	loop

	VUMAXV	V20.B16, V20
	VMOV	V20.B[0], R8
	CMP	$0, R8
	CSET	NE, R8
	MOVB	R8, ret+16(FP)
	RET
