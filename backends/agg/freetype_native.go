//go:build cgo && !purego

package agg

/*
// Default build: statically link the vendored FreeType 2.6.1 — matplotlib's
// pinned version, the one used to generate every reference image. It is built
// by `just freetype261-build` (third_party/freetype/build.sh) into the
// gitignored prefix. Pinning the FreeType version is what makes the AGG text
// rasterization byte-match the matplotlib references (the autohinter changed
// between 2.6.1 and current system FreeType, ~20 RMSE on dense text). ${SRCDIR}
// keeps the paths relocatable; only libfreetype.a ships in the prefix, so
// -lfreetype links statically.
//
// Compile fallback (-tags systemfreetype): link the system FreeType via
// pkg-config for environments without the vendored prefix (IDEs, quick vet).
// This is NOT parity-exact — golden/reference tests are expected to diverge —
// and exists only so the cgo packages compile without building FreeType 2.6.1.
#cgo !systemfreetype CFLAGS: -I${SRCDIR}/../../third_party/freetype/prefix/include/freetype2
#cgo !systemfreetype LDFLAGS: -L${SRCDIR}/../../third_party/freetype/prefix/lib -lfreetype -lm
#cgo systemfreetype pkg-config: freetype2
#include <ft2build.h>
#include FT_FREETYPE_H

static void mpl_go_version_freetype_version(FT_Library library, FT_Int *major, FT_Int *minor, FT_Int *patch) {
	FT_Library_Version(library, major, minor, patch);
}
*/
import "C"

import "fmt"

func nativeFreetypeVersion() string {
	var library C.FT_Library
	if C.FT_Init_FreeType(&library) != 0 {
		return ""
	}
	defer C.FT_Done_FreeType(library)

	var major, minor, patch C.FT_Int
	C.mpl_go_version_freetype_version(library, &major, &minor, &patch)
	return fmt.Sprintf("%d.%d.%d", int(major), int(minor), int(patch))
}
