// +build arm arm64
// +build go1.9

#include "textflag.h"
// func goid() int64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVW g, R14
	MOVW 152(R14), R13
	MOVW R13, ret+0(FP)
	RET
