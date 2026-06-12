//go:build skia && skiacgo

// skia_cwrap.cpp — implementation of the narrow C ABI declared in skia_cwrap.h.
//
// Compiled by cgo only under the `skiacgo` build tag (it is a C++ file in the
// package directory; cgo ignores it unless a .go file imports "C"). It links
// against a Skia shared library; build/link flags are supplied via
// CGO_CXXFLAGS / CGO_LDFLAGS (see the `skia-cgo-*` just recipes) so the Skia
// location stays out of the source tree.

#include "skia_cwrap.h"

#include <algorithm>
#include <cstring>
#include <vector>

#include "include/core/SkCanvas.h"
#include "include/core/SkColor.h"
#include "include/core/SkImage.h"
#include "include/core/SkImageInfo.h"
#include "include/core/SkMatrix.h"
#include "include/core/SkPaint.h"
#include "include/core/SkPath.h"
#include "include/core/SkPathBuilder.h"
#include "include/core/SkPoint.h"
#include "include/core/SkPixmap.h"
#include "include/core/SkRect.h"
#include "include/core/SkSamplingOptions.h"
#include "include/core/SkSurface.h"
#include "include/core/SkTileMode.h"
#include "include/core/SkVertices.h"
#include "include/core/SkSpan.h"
#include "include/effects/SkGradient.h"
#include "include/core/SkTypes.h"
#include "include/core/SkMilestone.h"

#define MGSK_STR2(x) #x
#define MGSK_STR(x) MGSK_STR2(x)
#ifndef SK_MILESTONE
#define SK_MILESTONE 0
#endif

struct MgSkSurface {
    sk_sp<SkSurface> surface;
    SkCanvas *canvas = nullptr;
    int width = 0;
    int height = 0;
};

namespace {

inline uint8_t toByte(float v) {
    if (v <= 0.0f) return 0;
    if (v >= 1.0f) return 255;
    return static_cast<uint8_t>(v * 255.0f + 0.5f);
}

inline SkColor colorFromFloat(float r, float g, float b, float a) {
    return SkColorSetARGB(toByte(a), toByte(r), toByte(g), toByte(b));
}

// Build an SkPath from the parallel verb/coord arrays.
SkPath buildPath(const uint8_t *verbs, int nverbs, const float *coords, int ncoords) {
    SkPathBuilder builder;
    int ci = 0;
    for (int i = 0; i < nverbs; ++i) {
        switch (verbs[i]) {
            case MGSK_VERB_MOVE:
                if (ci + 2 <= ncoords) {
                    builder.moveTo(coords[ci], coords[ci + 1]);
                    ci += 2;
                }
                break;
            case MGSK_VERB_LINE:
                if (ci + 2 <= ncoords) {
                    builder.lineTo(coords[ci], coords[ci + 1]);
                    ci += 2;
                }
                break;
            case MGSK_VERB_QUAD:
                if (ci + 4 <= ncoords) {
                    builder.quadTo(coords[ci], coords[ci + 1], coords[ci + 2], coords[ci + 3]);
                    ci += 4;
                }
                break;
            case MGSK_VERB_CUBIC:
                if (ci + 6 <= ncoords) {
                    builder.cubicTo(coords[ci], coords[ci + 1], coords[ci + 2],
                                    coords[ci + 3], coords[ci + 4], coords[ci + 5]);
                    ci += 6;
                }
                break;
            case MGSK_VERB_CLOSE:
                builder.close();
                break;
            default:
                break;
        }
    }
    return builder.detach();
}

sk_sp<SkShader> makeGradient(const MgSkPaint *p) {
    if (p == nullptr || p->grad_kind == MGSK_GRAD_NONE || p->grad_nstops <= 0 ||
        p->grad_offsets == nullptr || p->grad_colors == nullptr) {
        return nullptr;
    }
    std::vector<SkColor4f> colors(p->grad_nstops);
    std::vector<float> pos(p->grad_nstops);
    for (int i = 0; i < p->grad_nstops; ++i) {
        const float *c = p->grad_colors + i * 4;
        colors[i] = SkColor4f{c[0], c[1], c[2], c[3]};
        pos[i] = p->grad_offsets[i];
    }
    SkGradient::Colors gradColors(SkSpan<const SkColor4f>(colors.data(), colors.size()),
                                  SkSpan<const float>(pos.data(), pos.size()),
                                  SkTileMode::kClamp);
    SkGradient gradient(gradColors, SkGradient::Interpolation{});
    if (p->grad_kind == MGSK_GRAD_LINEAR) {
        SkPoint pts[2] = {{p->gx0, p->gy0}, {p->gx1, p->gy1}};
        return SkShaders::LinearGradient(pts, gradient);
    }
    if (p->grad_kind == MGSK_GRAD_RADIAL) {
        SkPoint center = {p->gx0, p->gy0};
        return SkShaders::RadialGradient(center, p->gr, gradient);
    }
    return nullptr;
}

// Configure and emit the fill pass (gradient or solid) for `paint` onto `path`.
void drawFill(SkCanvas *canvas, const SkPath &path, const MgSkPaint *paint) {
    sk_sp<SkShader> shader = makeGradient(paint);
    if (shader == nullptr && paint->fill_a <= 0.0f) {
        return;
    }
    SkPaint sk;
    sk.setAntiAlias(paint->antialias != 0);
    sk.setStyle(SkPaint::kFill_Style);
    if (shader != nullptr) {
        sk.setShader(shader);
        sk.setAlpha(255);
    } else {
        sk.setColor(colorFromFloat(paint->fill_r, paint->fill_g, paint->fill_b, paint->fill_a));
    }
    canvas->drawPath(path, sk);
}

void drawStroke(SkCanvas *canvas, const SkPath &path, const MgSkPaint *paint) {
    if (paint->stroke_a <= 0.0f || paint->line_width <= 0.0f) {
        return;
    }
    SkPaint sk;
    sk.setAntiAlias(paint->antialias != 0);
    sk.setStyle(SkPaint::kStroke_Style);
    sk.setStrokeWidth(paint->line_width);
    sk.setColor(colorFromFloat(paint->stroke_r, paint->stroke_g, paint->stroke_b, paint->stroke_a));
    canvas->drawPath(path, sk);
}

}  // namespace

extern "C" {

MgSkSurface *mgsk_surface_new(int width, int height) {
    if (width <= 0 || height <= 0) {
        return nullptr;
    }
    SkImageInfo info = SkImageInfo::Make(width, height, kRGBA_8888_SkColorType, kPremul_SkAlphaType);
    sk_sp<SkSurface> surface = SkSurfaces::Raster(info);
    if (surface == nullptr) {
        return nullptr;
    }
    MgSkSurface *s = new MgSkSurface();
    s->surface = surface;
    s->canvas = surface->getCanvas();
    s->width = width;
    s->height = height;
    s->canvas->clear(SK_ColorTRANSPARENT);
    return s;
}

void mgsk_surface_delete(MgSkSurface *s) {
    delete s;
}

void mgsk_surface_clear(MgSkSurface *s, float r, float g, float b, float a) {
    if (s == nullptr || s->canvas == nullptr) {
        return;
    }
    s->canvas->clear(colorFromFloat(r, g, b, a));
}

void mgsk_draw_path(MgSkSurface *s,
                    const uint8_t *verbs, int nverbs,
                    const float *coords, int ncoords,
                    const MgSkPaint *paint) {
    if (s == nullptr || s->canvas == nullptr || paint == nullptr || nverbs <= 0) {
        return;
    }
    SkPath path = buildPath(verbs, nverbs, coords, ncoords);
    drawFill(s->canvas, path, paint);
    drawStroke(s->canvas, path, paint);
}

void mgsk_draw_markers(MgSkSurface *s,
                       const uint8_t *verbs, int nverbs,
                       const float *coords, int ncoords,
                       const float *matrices,
                       const float *fill_colors,
                       const float *stroke_colors,
                       float line_width,
                       int antialias,
                       int count) {
    if (s == nullptr || s->canvas == nullptr || nverbs <= 0 || count <= 0 ||
        matrices == nullptr || fill_colors == nullptr) {
        return;
    }
    SkPath base = buildPath(verbs, nverbs, coords, ncoords);
    SkCanvas *canvas = s->canvas;
    for (int i = 0; i < count; ++i) {
        const float *m = matrices + i * 6;
        SkMatrix mat;
        mat.setAll(m[0], m[1], m[2], m[3], m[4], m[5], 0, 0, 1);

        canvas->save();
        canvas->concat(mat);

        const float *fc = fill_colors + i * 4;
        if (fc[3] > 0.0f) {
            SkPaint fill;
            fill.setAntiAlias(antialias != 0);
            fill.setStyle(SkPaint::kFill_Style);
            fill.setColor(colorFromFloat(fc[0], fc[1], fc[2], fc[3]));
            canvas->drawPath(base, fill);
        }
        if (stroke_colors != nullptr && line_width > 0.0f) {
            const float *sc = stroke_colors + i * 4;
            if (sc[3] > 0.0f) {
                SkPaint stroke;
                stroke.setAntiAlias(antialias != 0);
                stroke.setStyle(SkPaint::kStroke_Style);
                stroke.setStrokeWidth(line_width);
                stroke.setColor(colorFromFloat(sc[0], sc[1], sc[2], sc[3]));
                canvas->drawPath(base, stroke);
            }
        }
        canvas->restore();
    }
}

void mgsk_draw_vertices(MgSkSurface *s,
                        const float *positions,
                        const float *colors,
                        int triangle_count,
                        int antialias) {
    if (s == nullptr || s->canvas == nullptr || positions == nullptr ||
        colors == nullptr || triangle_count <= 0) {
        return;
    }
    const int vertexCount = triangle_count * 3;
    std::vector<SkPoint> pts(vertexCount);
    std::vector<SkColor> cols(vertexCount);
    for (int i = 0; i < vertexCount; ++i) {
        pts[i] = SkPoint{positions[i * 2], positions[i * 2 + 1]};
        const float *c = colors + i * 4;
        cols[i] = colorFromFloat(c[0], c[1], c[2], c[3]);
    }
    sk_sp<SkVertices> vertices = SkVertices::MakeCopy(
        SkVertices::kTriangles_VertexMode, vertexCount, pts.data(), nullptr, cols.data());
    if (vertices == nullptr) {
        return;
    }
    SkPaint paint;
    paint.setAntiAlias(antialias != 0);
    paint.setColor(SK_ColorWHITE);
    // kModulate with an opaque white paint yields the raw interpolated vertex
    // colors (white * color == color).
    s->canvas->drawVertices(vertices, SkBlendMode::kModulate, paint);
}

void mgsk_draw_image(MgSkSurface *s,
                     const uint8_t *pixels,
                     int width,
                     int height,
                     int stride,
                     const float *matrix,
                     float alpha,
                     int sampling) {
    if (s == nullptr || s->canvas == nullptr || pixels == nullptr || matrix == nullptr ||
        width <= 0 || height <= 0 || stride < width * 4 || alpha <= 0.0f) {
        return;
    }
    SkImageInfo info = SkImageInfo::Make(width, height, kRGBA_8888_SkColorType,
                                         kUnpremul_SkAlphaType);
    SkPixmap pixmap(info, pixels, static_cast<size_t>(stride));
    sk_sp<SkImage> image = SkImages::RasterFromPixmapCopy(pixmap);
    if (image == nullptr) {
        return;
    }

    SkSamplingOptions options;
    if (sampling == 1) {
        options = SkSamplingOptions(SkFilterMode::kLinear);
    } else if (sampling == 2) {
        options = SkSamplingOptions(SkCubicResampler::Mitchell());
    }

    SkPaint paint;
    paint.setAntiAlias(true);
    paint.setAlphaf(std::clamp(alpha, 0.0f, 1.0f));

    SkMatrix mat;
    mat.setAll(matrix[0], matrix[1], matrix[2],
               matrix[3], matrix[4], matrix[5],
               0, 0, 1);

    s->canvas->save();
    s->canvas->concat(mat);
    SkRect src = SkRect::MakeWH(static_cast<SkScalar>(width), static_cast<SkScalar>(height));
    s->canvas->drawImageRect(image, src, src, options, &paint, SkCanvas::kStrict_SrcRectConstraint);
    s->canvas->restore();
}

void mgsk_surface_read_pixels(MgSkSurface *s, uint8_t *dst, int stride) {
    if (s == nullptr || s->surface == nullptr || dst == nullptr || stride <= 0) {
        return;
    }
    SkImageInfo info = SkImageInfo::Make(s->width, s->height, kRGBA_8888_SkColorType,
                                         kUnpremul_SkAlphaType);
    s->surface->readPixels(info, dst, static_cast<size_t>(stride), 0, 0);
}

const char *mgsk_version(void) {
    static const char version[] = "Skia milestone " MGSK_STR(SK_MILESTONE);
    return version;
}

}  // extern "C"
