#include "textflag.h"
#define g_goid 152

#ifdef GOARCH_amd64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ g_goid(R14), R13
	MOVQ R13, ret+0(FP)
	RET
#endif

#ifdef GOARCH_arm64
TEXT ·goid(SB),NOSPLIT,$0-8
	MOVD g, R14
	MOVD g_goid(R14), R13
	MOVD R13, ret+0(FP)
	RET
#endif
