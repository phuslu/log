#include "textflag.h"

#ifdef GOOS_linux
#ifdef GOARCH_amd64
// func walltime() (sec int64, nsec int32)
TEXT ·walltime(SB), NOSPLIT, $24-12
	CALL	runtime·nanotime1(SB)
	MOVQ	AX, ret+0(FP)  // sec
	MOVL	DX, ret+8(FP)  // nsec
	RET
#endif
#endif

#ifdef GOOS_linux
#ifdef GOARCH_arm64
// func walltime() (sec int64, nsec int32)
TEXT ·walltime(SB), NOSPLIT, $24-12
    // Call runtime.walltime
    CALL      runtime·walltime(SB)
    // Get the return values
    MOVD    R3, ret+0(FP)  // sec
    MOVW    R5, ret+8(FP)  // nsec
    RET
#endif
#endif

