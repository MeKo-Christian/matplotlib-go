# Phase 2.3: the options model

Phase 2.3 calls for replacing "the 83 variadic option structs and 408
pointer-to-primitive fields with one consistent options model", where "extra
option sets must be impossible or rejected". The previous sub-bullet delivered
the _rejected_ half at run time (see
[phase2-extra-option-rejection.md](phase2-extra-option-rejection.md)). This
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

## Follow-on work

The remaining option families migrate under the next Phase 2.3 sub-bullet.
`PlotOptions` is the largest and is deliberately not in the pilot: it is shared
by 30 entry points across the line, contour, surface, and 3D families, so
migrating it migrates four families at once and would not be a pilot. Once the
variadic tails are gone repo-wide, `internal/optarg` and the run-time rejection
it implements disappear with them.
