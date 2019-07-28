// +build arm,arm64
// +build go1.9

#include "textflag.h"
// func getg() *g
TEXT ·getg(SB),NOSPLIT,$0-8
	MOVW g, ret+0(FP)
	RET
