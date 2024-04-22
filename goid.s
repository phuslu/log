#include "textflag.h"

// After go1.9, goid in "goroutine struct" has a stable offset.
// See https://github.com/golang/go/blob/master/src/runtime/runtime2.go#L458
//
// This file exposes "goid()" function by
//   1. get current "g" pointer,
//   2. extract goid with a hardcoded offset.
//      *) for GOARCH amd64/arm64, offset=152, size=8.
//      *) for GOARCH arm/386, offset=80, size=4.
//

#ifdef GOARCH_amd64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ 152(R14), R13
	MOVQ R13, ret+0(FP)
	RET
#endif

#ifdef GOARCH_arm64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVD g, R14
	MOVD 152(R14), R13
	MOVD R13, ret+0(FP)
	RET
#endif

#ifdef GOARCH_arm
TEXT ·goid(SB),NOSPLIT,$0-4
	MOVW g, R8
	MOVW 80(R8), R7
	MOVW R7, ret+0(FP)
	RET
#endif

#ifdef GOARCH_386
TEXT ·goid(SB),NOSPLIT,$0-4
	MOVL (TLS), AX
	MOVL 80(AX), BX
	MOVL BX, ret+0(FP)
	RET
#endif

#ifdef GOARCH_mipsle
TEXT ·goid(SB),NOSPLIT,$0-4
	MOVW g, R8
	MOVW 80(R8), R7
	MOVW R7, ret+0(FP)
	RET
#endif
