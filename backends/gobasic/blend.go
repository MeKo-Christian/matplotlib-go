package gobasic

import "image/color"

func (r *Renderer) blendPixel(x, y int, src color.RGBA) {
	if src.A == 0 {
		return
	}
	if len(r.clipPaths) > 0 {
		clipA := r.clipMaskAlphaAt(x, y)
		if clipA == 0 {
			return
		}
		src.A = uint8(uint16(src.A) * uint16(clipA) / 255)
		if src.A == 0 {
			return
		}
	}
	r.blendPixelNoClip(x, y, src)
}

func (r *Renderer) blendPixelNoClip(x, y int, src color.RGBA) {
	i := r.dst.PixOffset(x, y)
	dr := uint32(r.dst.Pix[i])
	dg := uint32(r.dst.Pix[i+1])
	db := uint32(r.dst.Pix[i+2])
	da := uint32(r.dst.Pix[i+3])

	sr := uint32(src.R)
	sg := uint32(src.G)
	sb := uint32(src.B)
	sa := uint32(src.A)

	if sa == 255 {
		r.dst.Pix[i] = src.R
		r.dst.Pix[i+1] = src.G
		r.dst.Pix[i+2] = src.B
		r.dst.Pix[i+3] = src.A
		return
	}

	outA := sa + ((255 - sa) * da / 255)
	if outA == 0 {
		r.dst.Pix[i] = 0
		r.dst.Pix[i+1] = 0
		r.dst.Pix[i+2] = 0
		r.dst.Pix[i+3] = 0
		return
	}

	r.dst.Pix[i] = uint8((sr*sa + dr*(255-sa)*da/255) / outA)
	r.dst.Pix[i+1] = uint8((sg*sa + dg*(255-sa)*da/255) / outA)
	r.dst.Pix[i+2] = uint8((sb*sa + db*(255-sa)*da/255) / outA)
	r.dst.Pix[i+3] = uint8(outA)
}
