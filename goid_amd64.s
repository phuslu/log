// +build amd64

#include "textflag.h"
// func goid() int64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ 152(R14), R13
	MOVQ R13, ret+0(FP)
	RET
