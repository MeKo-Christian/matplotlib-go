/* skia_cwrap.h — narrow C ABI over Skia for matplotlib-go's native backend.
 *
 * This header intentionally exposes only the small set of primitives the Go
 * Skia backend needs to promote its bridged (CPU) paths to native Skia
 * rasterization: an offscreen raster surface, path/marker/vertex draws, and a
 * deterministic unpremultiplied RGBA readback. All coordinates are in the
 * device space the Go caller has already resolved (y-down, top-left origin),
 * so the wrapper performs no coordinate flips of its own.
 *
 * The ABI is plain C (extern "C") so cgo can call it without C++ name mangling.
 * It is compiled by cgo only under the `skiacgo` build tag; see native_cgo.go.
 */
#ifndef MATPLOTLIB_GO_SKIA_CWRAP_H
#define MATPLOTLIB_GO_SKIA_CWRAP_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Opaque handle to an offscreen raster surface + its canvas. */
typedef struct MgSkSurface MgSkSurface;

/* Path verb codes. These mirror the values the Go side emits when translating
 * internal/geom.Cmd; they are part of the ABI contract, not Skia's own enum. */
enum {
    MGSK_VERB_MOVE  = 0,
    MGSK_VERB_LINE  = 1,
    MGSK_VERB_QUAD  = 2,
    MGSK_VERB_CUBIC = 3,
    MGSK_VERB_CLOSE = 4
};

/* Gradient kinds for MgSkPaint.grad_kind. */
enum {
    MGSK_GRAD_NONE   = 0,
    MGSK_GRAD_LINEAR = 1,
    MGSK_GRAD_RADIAL = 2
};

/* MgSkPaint describes one fill (solid or gradient) and one optional stroke.
 * Colors are straight (unpremultiplied) RGBA in [0,1]. A fill is drawn when
 * fill_a > 0 or a gradient is set; a stroke is drawn when stroke_a > 0 and
 * line_width > 0. Gradient stops, when present, override the solid fill. */
typedef struct {
    float fill_r, fill_g, fill_b, fill_a;
    float stroke_r, stroke_g, stroke_b, stroke_a;
    float line_width;
    int   antialias; /* 0 = off, nonzero = on */

    int   grad_kind; /* MGSK_GRAD_* */
    /* linear: (gx0,gy0) -> (gx1,gy1); radial: center (gx0,gy0), radius gr. */
    float gx0, gy0, gx1, gy1, gr;
    int   grad_nstops;
    const float *grad_offsets; /* grad_nstops floats */
    const float *grad_colors;  /* grad_nstops*4 floats (rgba, straight alpha) */
} MgSkPaint;

/* Create / destroy an offscreen raster surface sized width x height. Returns
 * NULL on failure (non-positive dimensions or allocation failure). */
MgSkSurface *mgsk_surface_new(int width, int height);
void         mgsk_surface_delete(MgSkSurface *s);

/* Clear the whole surface to a straight-alpha RGBA color. */
void mgsk_surface_clear(MgSkSurface *s, float r, float g, float b, float a);

/* Draw one path described by parallel verb/coord arrays. `verbs` holds nverbs
 * MGSK_VERB_* codes; `coords` holds ncoords floats consumed as (x,y) pairs in
 * the order the verbs require (move/line: 1 pt, quad: 2, cubic: 3, close: 0). */
void mgsk_draw_path(MgSkSurface *s,
                    const uint8_t *verbs, int nverbs,
                    const float *coords, int ncoords,
                    const MgSkPaint *paint);

/* Draw `count` instances of one base marker path. Each instance i is placed by
 * the 3x2 affine matrices[i*6] = {scaleX, skewX, transX, skewY, scaleY, transY}
 * (Skia row order) and filled with fill_colors[i*4]. If stroke_colors is
 * non-NULL and line_width > 0 each instance is also stroked with
 * stroke_colors[i*4]. This is the native batch replacement for the per-item
 * Path loop; it uses one SkPath and one surface for the whole batch. */
void mgsk_draw_markers(MgSkSurface *s,
                       const uint8_t *verbs, int nverbs,
                       const float *coords, int ncoords,
                       const float *matrices,     /* 6*count */
                       const float *fill_colors,  /* 4*count */
                       const float *stroke_colors,/* 4*count or NULL */
                       float line_width,
                       int antialias,
                       int count);

/* Draw `triangle_count` Gouraud-shaded triangles via SkVertices. `positions`
 * holds 6*triangle_count floats (3 (x,y) device-space vertices per triangle);
 * `colors` holds 12*triangle_count floats (3 straight-alpha rgba per triangle). */
void mgsk_draw_vertices(MgSkSurface *s,
                        const float *positions,
                        const float *colors,
                        int triangle_count,
                        int antialias);

/* Draw an RGBA image through an affine transform. `pixels` is straight-alpha
 * RGBA8888, `stride` bytes per source row. `matrix` is a 3x2 affine in Skia row
 * order {scaleX, skewX, transX, skewY, scaleY, transY} mapping source image
 * pixels into device space. sampling: 0 nearest, 1 linear, 2 cubic Mitchell. */
void mgsk_draw_image(MgSkSurface *s,
                     const uint8_t *pixels,
                     int width,
                     int height,
                     int stride,
                     const float *matrix,
                     float alpha,
                     int sampling);

/* Draw hatch pixels clipped by `path` using a repeated SkShader tile. `hatch`
 * holds the Matplotlib hatch character sequence (`|`, `-`, `/`, `\`, `+`, `x`,
 * `X`, `o`, `O`, `.`, `*`). Colors are straight-alpha RGBA in [0,1]. */
void mgsk_draw_hatch_path(MgSkSurface *s,
                          const uint8_t *verbs, int nverbs,
                          const float *coords, int ncoords,
                          const char *hatch, int hatch_len,
                          float r, float g, float b, float a,
                          float line_width,
                          float spacing,
                          int antialias);

/* Copy the surface into dst as straight-alpha RGBA8888, `stride` bytes per row.
 * dst must hold at least height*stride bytes. */
void mgsk_surface_read_pixels(MgSkSurface *s, uint8_t *dst, int stride);

/* Returns a static Skia milestone/version string (never NULL). */
const char *mgsk_version(void);

#ifdef __cplusplus
}
#endif

#endif /* MATPLOTLIB_GO_SKIA_CWRAP_H */
