#include "textflag.h"

// func getg() int64
TEXT ·getg(SB),NOSPLIT,$0-8
	MOVW g, ret+0(FP)
	RET
