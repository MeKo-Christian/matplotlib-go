# Phase 2.3: rejecting extra option values

Phase 2.3 requires that "extra option sets must be impossible or rejected". The
final options model delivers the _impossible_ half by removing the variadic
tail. This document covers the intermediate safety step that keeps every current
signature and makes the misuse _rejected_ today.

## The rule

Almost every plotting entry point spells its options as a variadic tail:

```go
func (a *Axes) ImShow(data [][]float64, opts ...ImShowOptions) *Image2D
```

The tail is a stand-in for one optional argument, so exactly zero or one value
is meaningful. Before this pass, `ax.ImShow(data, first, second)` used `first`
and silently dropped `second` — the same class of quiet failure that the
warn-and-skip conversion removed from the data path.

The rule is now uniform: **at most one option value; more is rejected.**

## Where the rejection is enforced

`internal/optarg` is the single implementation. It offers three entry points,
all built on the same `*TooManyError`:

| Helper                        | Used by                                                                                            | Behavior on two or more values |
| ----------------------------- | -------------------------------------------------------------------------------------------------- | ------------------------------ |
| `optarg.Only(call, opts)`     | entry points that already return an error                                                          | returns `*TooManyError`        |
| `optarg.One(call, opts)`      | entry points with no error result                                                                  | panics with `*TooManyError`    |
| `optarg.Optional(call, opts)` | entry points that merge options over prepared defaults, so "absent" must stay distinct from "zero" | panics with `*TooManyError`    |

The split is deliberate. Rejected _input_ stays on the error channel, matching
the convention adopted by the first Phase 2.3 bullet. An extra option value is
not input: it cannot arrive from plot data, because it requires two option
literals written at the call site. That makes it a caller bug, and the same bug
the final options model will reject at compile time — so the non-error entry
points panic instead of inventing a second failure channel or growing an error
result they do not otherwise need.

`optarg` also derives the option type's name by reflection, so the message
names both the call and the type without every site repeating them:

```
imshow accepts at most one ImShowOptions value (got 2)
```

## Entry points that became non-variadic

Where an internal helper received an already-unpacked option set, its variadic
tail was removed outright rather than re-checked. Extra values are now
impossible there, and the exported caller owns the check:

- `Axes.plot`, `Axes.scatter`, `Axes.bar` — take one option value; `Plot`,
  `Scatter`, `Bar`/`BarH`, `SemilogX`, `SemilogY`, `LogLog`, and
  `BarContainer` unpack before calling.
- `Axes.pcolorMesh` — `PColor`, `PColorFast`, and `PColorMesh` unpack.
- `Axes.buildContourSet`, `contourGridCoordsValues`, `contourGridTriangulation`
  — `Contour`, `Contourf`, `TriContour`, and `TriContourf` unpack.
- `Axes.lineCollectionFromSegments` — `HLines` and `VLines` unpack.
- `validateScatterInput`, `resolveImShowRGBOptions`, `vectorScalarOptions`,
  `barbsScalarOptions`, and `pyplot.makeColorbarRoom` — take the resolved value
  plus, where the distinction matters, a `supplied` flag.
- `plot3d.firstPlotOptions` was removed; its callers use `optarg.One` directly.

This is why the pass leaves the frozen public API unchanged: every signature
that moved was unexported.

## Variadic parameters deliberately left alone

Not every variadic tail is an optional-argument stand-in. These take a genuine
list and are untouched:

- Functional-option lists: `style.Option`, `render.SaveOption`, `PSOption`,
  `PGFOption`, `PDFOption`, `SVGOption`, `transform.ScaleOption`,
  `animation.SaveOption`, `core.AxesDividerOption`, `SubplotOption`,
  `SubplotAxesOption`, `GridSpecOption`, `InsetAxesOption`,
  `ParasiteAxesOption`, and `color.ToRGBAOption`. Supplying several is the
  intended usage.
- True varargs: `geom.RectFromPoints`, `geom.UnionRects`,
  `geom.NewBezierSegment`, `cycler.New`, and `diag.Warnf`.
- Optional trailing scalars and flags (`...float64`, `...bool`, `...string`)
  used as poor-man's optional parameters inside unexported helpers. These
  belong with the raw-string/typed-constant bullet, not this one.

## Follow-on work

This step does not choose the final options representation; that is the next
Phase 2.3 sub-bullet. When it lands, the variadic tails disappear and
`internal/optarg` disappears with them. Until then it is the only place the
"at most one" rule is written down.
