# Phase 2.3: the options model

Phase 2.3 calls for replacing "the 83 variadic option structs and 408
pointer-to-primitive fields with one consistent options model", where "extra
option sets must be impossible or rejected". The previous sub-bullet delivered
the _rejected_ half at run time (see
[extra-option-rejection.md](extra-option-rejection.md)). This
document chooses the final representation, which delivers the _impossible_
half, and records the pilot that proves it on one line, collection, image, and
annotation API.

## What was wrong

Two independent problems, both visible in a single option struct:

```go
type ImShowOptions struct {
    Colormap *string   // nil means "use the rc default"
    Alpha    *float64  // nil means "opaque"
    Aspect   string    // "" means "use the rc default"
    Origin   ImageOrigin // the zero value doubles as "unset"
}

func (a *Axes) ImShow(data [][]float64, opts ...ImShowOptions) *Image2D
```

**The variadic tail** stands in for one optional argument. Zero or one value is
meaningful; two or more was silently truncated to the first.

**The optional fields** spell "absent" two different ways, and both lose
information:

- The pointer spelling needs an addressable temporary at every call site
  (`interp := "bilinear"; ... Interpolation: &interp`), shares mutable state
  between the caller and the artist, and invites nil dereferences. The repo had
  accumulated eleven separate `ptr`/`floatPtr`/`boolPtr` helpers to paper over
  it.
- The magic-zero spelling cannot express a legitimate zero. `ImShowOptions.Origin`
  carried a comment admitting this: with a non-default rc of `image.origin:
lower`, upper could not be forced per call. `Axes.Annotate` could not place
  text at offset `(0, 0)`, or draw a zero-width arrow. `Axes.HLines` forced
  `Alpha: 0` to 1 and `LineWidth: 0` to 1.

## The model

**1. Options arrive as exactly one value.** The variadic tail is removed:

```go
func (a *Axes) ImShow(data [][]float64, opt ImShowOptions) *Image2D
```

A second option value is now a compile error, which is the "impossible" half of
the requirement. Callers who want defaults pass the zero struct:

```go
ax.ImShow(data, core.ImShowOptions{})
```

The extra `{}` is the price of the guarantee, and it makes the options
parameter's existence obvious at the call site.

**2. Optional fields are `optional.Value[T]`.** The new
[`optional`](../../optional/optional.go) package holds one comparable tri-state
value whose zero value is "absent":

```go
ax.ImShow(data, core.ImShowOptions{
    Colormap: optional.Of("viridis"),
    Alpha:    optional.Of(0.0), // explicitly transparent, not "unspecified"
})
```

Resolving an option against its default becomes a single expression, so the
merge ladders and the "was anything supplied?" flags that used to accompany
them disappear:

```go
aspect := opt.Aspect.Or(imshowAspectDefault(&rc))
origin := opt.Origin.Or(imageOriginFromRC(&rc))
```

`Value` is copied by value, so an options struct never shares storage with the
artist built from it. `Value.Ptr()` bridges to artist fields and renderer
structs that still spell optional values as pointers, returning a pointer to a
fresh copy.

### Which fields become optional

Not every field. Wrapping all of them would add ceremony without adding
information. The rule is:

| Field                                                        | Spelling                             |
| ------------------------------------------------------------ | ------------------------------------ |
| Has a non-zero default (rc-derived or a documented constant) | `optional.Value[T]`                  |
| Its zero value _is_ the default and is not otherwise special | plain `T`                            |
| Interface, slice, or map                                     | plain `T` — nil already means absent |

So `ImShowOptions.Label` stays a plain `string` (empty means "no legend entry",
and there is no default to fall back to), `ImShowOptions.Norm` stays a plain
interface, and `StemOptions.Baseline` was **demoted** from `*float64` to
`float64` because its default is 0 and 0 is an ordinary value. The rule prunes
as well as it wraps.

**3. Typed constants** for the remaining raw-string option enums are the next
sub-bullet. `StemOptions.Orientation`, `LineCollectionOptions.LineStyle`, and
`ImShowOptions.Aspect`/`Interpolation` are still strings.

## Alternatives considered

- **Functional options** (`ImShow(data, WithCmap("viridis"))`). Idiomatic Go,
  but 83 option structs averaging ~15 fields means roughly 1,200 constructor
  functions to write, document, and freeze — and it discards the struct-literal
  shape that makes this port readable next to matplotlib's kwargs.
- **Pointer to an options struct** (`ImShow(data, *ImShowOptions)`, nil for
  defaults). Makes extras impossible, but does nothing about the 400-odd
  optional fields and adds a nil check to every entry point.
- **Keeping the variadic tail** and fixing only the fields. Leaves the extra
  option set rejectable but not impossible, which the phase explicitly rules
  out.

## The pilot

Four APIs, one per artist family, migrated end to end — signature and fields:

| Family     | API                           | Options                       | What it proves                                                                                       |
| ---------- | ----------------------------- | ----------------------------- | ---------------------------------------------------------------------------------------------------- |
| line       | `Axes.Stem`                   | `StemOptions`                 | a wide struct: ten `!= nil` ladders collapse to `.Or(...)`, and one field demotes to plain `float64` |
| collection | `Axes.HLines` / `Axes.VLines` | `LineCollectionOptions` (new) | the worst case: the options type _was_ the artist type                                               |
| image      | `Axes.ImShow`                 | `ImShowOptions`               | rc-derived defaults, and the documented `Origin` tri-state defect                                    |
| annotation | `Axes.Annotate`               | `AnnotationOptions`           | magic-zero defaults for `OffsetX/Y`, `ArrowWidth`, `ArrowHeadSize`                                   |

Their `pyplot` wrappers migrated with them. No golden or reference image
changed.

`HLines`/`VLines` previously took a `LineCollection` — the artist — as their
options, so callers could set `Segments` (which the helper overwrote anyway)
and the defaults were inferred from magic zeros. `LineCollectionOptions`
separates the two and covers what Matplotlib's `hlines`/`vlines` accept.
Scalar-mapping fields are deliberately not carried over; a caller who needs
them builds a `LineCollection` and calls `Axes.AddCollection`.

### Behavior notes

Rendering is unchanged for every call site in the repo, and the migration
preserved two subtleties worth naming:

- `HLines`/`VLines` only fall back to the property cycle when the per-segment
  `Colors` slice is empty, so a caller supplying `Colors` still does not advance
  the cycle.
- `Stem` still ignores an `Alpha` outside `[0, 1]` rather than clamping it.

The newly expressible values (`Alpha: optional.Of(0.0)`, an annotation offset of
`(0, 0)`, `Origin: optional.Of(ImageOriginUpper)` against an rc of `lower`) had
no way to be requested before, so no existing caller can regress. One case does
change meaning: an `AnnotationOptions` literal that set only one of `OffsetX`
or `OffsetY` used to get 0 for the other; now the unset one falls back to the
Matplotlib default. Call sites in this repository were migrated accordingly.

## The full migration

The pilot's model was applied to every remaining option family in one pass. The
end state:

| Measure                        | Before | After |
| ------------------------------ | ------ | ----- |
| variadic `...FooOptions` tails | 205    | 0     |
| pointer-to-primitive fields    | 408    | 0     |
| `*Options` structs             | 97     | 101   |

The eight pointers left in an `Options` struct are references to live objects
(`*core.Figure`, `*canvas.Navigation`, `*log.Logger`, …), not optional
primitives, so they stay pointers.

Two shapes recurred often enough to be worth naming, because collapsing them is
what made the migration a net deletion rather than a net rewrite:

- **The prepared-defaults literal.** `Axes.Table`, `Axes.Pie`, `Axes.Hexbin`,
  `Axes.EventPlot`, `Axes.Violinplot`, `Sankey`, `MatShow`, `Spy`,
  `AnnotatedHeatmap`, and the widget constructors all built a fully populated
  defaults value, overwrote it wholesale with the caller's options, and then
  re-applied every default through a magic-zero guard. Once the options arrive
  as a single value the prepared literal is dead code: `cfg := opt` plus the
  guards is exactly equivalent.
- **The `supplied` flag.** `optarg.Optional` handed back "was anything passed?"
  so a merge block could be skipped. Every one of those blocks was a no-op
  against a zero options value, and `Axes3D.Plot3D`/`Axes3D.Scatter3D` had two
  textually identical branches selected by the flag. `optional.Value` carries
  the same information per field, so the flag and the branches both went away.

The four `clone*` helpers (`cloneBool`, `cloneTextAlign`,
`cloneTextBBoxOptions`, `clonePoint`) existed only to stop a caller and an
artist sharing a pointer. `optional.Value` copies by value, so they are gone;
`cloneFontProperties` survives as `cloneFontPropertiesValue` because it also
deep-copies the slices inside `render.FontProperties`.

`internal/optarg` and the run-time rejection it implemented are deleted: with
no variadic tails left there is nothing to reject.

## Typed constants

The raw-string option enums are now defined string types in
[`core/option_enums.go`](../../core/option_enums.go): `PlotOrientation`,
`ColorbarExtend`, `ColorbarLocation`, `ImageAspect`, `VectorPivot`, and
`ViolinSide`. They join the types that already followed this convention
(`LineStyle`, `MeshShading`, `SignalDetrend`, `SignalSpectrumScale`,
`SignalSpectrumSides`, `TextRotationMode`, `FillBetween3DMode`).

A defined string type still accepts an untyped string constant, so
`Orientation: "horizontal"` keeps compiling; what changes is that the value can
no longer be produced by a plain `string` variable without an explicit
conversion, and the constants document the accepted set. Call sites in this
repository were moved to the constants (`Aspect: core.AspectAuto`).

Fields that look like enums but are not were deliberately left as `string`:
`TextBBoxOptions.Style` is a Matplotlib boxstyle _spec_ ("round,pad=0.3"),
`BarLabelOptions.Format` and `AnnotatedHeatmapOptions.Format` are printf
formats, and `Colormap`/`FontKey`/`Label` are open-ended names.

## Follow-on work

Two refinements are out of scope here and remain for later Phase 2 work:

- **Magic-zero plain fields.** A handful of non-optional fields still spell
  "use the default" as their zero value where the default is non-zero —
  `TableOptions.FontSize`, `TableOptions.CellLoc`, `EventPlotOptions.LineWidth`.
  These are expressiveness gaps of the same kind the pointer fields had, but
  the phase bullet enumerates variadic tails, pointer fields, and raw-string
  enums, and all three are now closed.
- **Artist fields.** Several artists still expose optional state as pointers
  (`Text.ParseMath`, `Text.BBox`, `ErrorBar.CapSize`). Those are exported
  mutable artist fields, which the next Phase 2.3 bullet covers; the option
  structs bridge to them with `Value.Ptr()`.
