// +build amd64 amd64p32
// +build go1.9

#include "textflag.h"
// func getg() uintptr
TEXT ·getg(SB),NOSPLIT,$0-8
	MOVQ (TLS), BX
	MOVQ BX, ret+0(FP)
	RET
