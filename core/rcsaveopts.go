package core

import (
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// pdfFontPolicyFromRC maps the pdf.* rcParams to the backend font policy.
// pdf.fonttype 42 (TrueType) and pdf.use14corefonts request real PDF text
// (embedded), while the Type-3 default falls back to glyph outlines, matching
// the port's current default.
func pdfFontPolicyFromRC(pdf style.PDFRC) render.PDFFontPolicy {
	if pdf.Use14CoreFonts || pdf.FontType == 42 {
		return render.PDFFontPolicyEmbed
	}
	return render.PDFFontPolicyPath
}

// psFontPolicyFromRC maps the ps.* rcParams to the backend font policy.
// ps.useafm requests direct Base-14 text; otherwise glyph outlines are used,
// matching the port's current default.
func psFontPolicyFromRC(ps style.PSRC) render.PSFontPolicy {
	if ps.UseAFM {
		return render.PSFontPolicyBase14
	}
	return render.PSFontPolicyPath
}
