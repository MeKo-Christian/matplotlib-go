# Exported mutable fields vs. setter duplication

Companion to [`options-model.md`](options-model.md). That document
settled how options _enter_ an artist; this one settles who may _mutate_ an
artist afterwards.

## The problem

An audit of every exported struct in the public packages found **1750 exported
non-option fields** across 14 packages, and **29 of them are shadowed by a
`Set<Field>` method on the same type**. Two public writers for one concept is
not merely redundant — in most of these cases the two writers do _different
things_, and the field write is the wrong one:

```go
line.Dashes = []float64{4, 2}      // leaves DashUnits at DashUnitsPixels
line.SetDashes(4, 2)               // sets DashUnits to DashUnitsMatplotlib
```

Both compiled. Only the second rendered the dash pattern matplotlib would. (The
first no longer compiles — see the `DashPattern` section below.)

## Classification

The plan bullet asks for three buckets. Applying them to the audit:

| Bucket                      | Meaning                                                                        | Ownership                                                  |
| --------------------------- | ------------------------------------------------------------------------------ | ---------------------------------------------------------- |
| **Immutable configuration** | Fixed at construction; the renderer reads it and nothing writes it afterwards. | Exported field, no setter.                                 |
| **Observable state**        | User-mutable after construction; a write is visible in the next draw.          | _Depends on what else the write has to touch_ — see below. |
| **Internal cache**          | Derived from the other two; invalidated, never authored.                       | Unexported, no accessor.                                   |

The internal-cache bucket was already clean: stale bits, resolved transform
nodes, and memoized layouts are unexported throughout. The
immutable-configuration bucket is likewise uncontroversial — `Axes.RectFraction`,
`Figure.SizePx`, and the `style.*RC` structs are plain fields with no setters.

All 29 conflicts sit in **observable state**, and the deciding question is:

> **Does a correct write touch anything other than this field?**

## Three kinds of companion, three different fixes

Answering that question sorts the conflicts into three groups, and each wants a
different treatment. The first two splits were the main finding of the pass; the
third closed the families it could not classify.

### 1. The companion is a hand-rolled optional → widen the field

`Patch.FaceColor` was a plain `render.Color` paired with an unexported
`faceColorSet bool`. The flag exists only to distinguish "no fill assigned,
inherit `patch.facecolor`" from "deliberately transparent". That is precisely
`optional.Value[T]`, which the options model already adopted for exactly this
reason — and which the repo's own classification rule ("non-zero rc-derived
default → `optional.Value`") already prescribed.

So the fix is not to hide the field but to **widen its type and delete the
flag**:

```go
FaceColor optional.Value[render.Color]   // was: FaceColor render.Color + faceColorSet bool
```

The duplication dissolves rather than being papered over. There is no companion
left, so a direct write cannot desynchronize anything, the field stays exported
and literal-friendly, and `SetFaceColor` survives as a convenience wrapper that
no longer carries a hidden invariant. Three `bool` fields disappeared.

This also preserved the idiom the examples are built on. Encapsulating
`Patch.FaceColor` instead would have broken every

```go
ax.AddPatch(&core.Rectangle{Patch: core.Patch{FaceColor: c, EdgeWidth: 1.1}, ...})
```

into four statements. `FaceColor: optional.Of(c)` keeps it one expression.

Applied to: `Patch.FaceColor/EdgeColor/EdgeWidth` (+3 flags removed),
`Line2D.GapColor` (+`GapColorSet`), `PathCollection.OffsetCoords`
(+`offsetCoordsSet`).

### 2. The companion is real behavior → encapsulate

Where the setter clamps, normalizes, repositions a caret, or fires a callback,
no field type can carry that. There the setter must be the only writer, and the
field becomes unexported with an exported reader.

Applied to the axes title/label family and the widget value family (below).

### 3. The companion is a second value → merge them

Two families were carried over from the first pass because neither fix above
applied: their companion is not a flag but another _value_.

#### Mirroring belongs at read time, not write time

`PathCollection`, `PatchCollection`, and `QuadMesh` share an `EdgeColorsFace`
flag (Matplotlib's `edgecolors="face"`). It was honored by copying `FaceColors`
into `EdgeColors` from four different write sites — `SetFaceColors`,
`SetEdgeColorFace`, `refreshScalarMappedColors`, and again by hand in
`Scatter2D.toPathCollection`. A direct `c.FaceColors = …` skipped all four, and
even between the setters the result depended on call order.

Reading the flag where the stroke color is resolved makes the copy unnecessary:

```go
func (c *PathCollection) resolvedEdgeColors() []render.Color {
	if c != nil && c.EdgeColorsFace && len(c.FaceColors) > 0 {
		return c.FaceColors
	}
	return c.EdgeColors
}
```

The `len(c.FaceColors) > 0` guard is not incidental — it is what the write-time
mirror did, and it is load-bearing. `Scatter2D` builds an unfilled marker
(`MarkerFillNone`) by moving the face colors into `EdgeColors` and then clearing
the face, relying on the flag as a _snapshot_ rather than a live mirror. With the
guard, that path still reads its stored edge colors, so a naive
"`EdgeColorsFace` ⇒ return the face colors" would have silently blanked every
unfilled scatter marker.

Two write-order behaviors were made explicit rather than dropped:
`SetEdgeColors` no longer clears the flag (the flag simply outranks the field
while it is set), and the one internal caller that really did mean "explicit edge
colors win" now says `pc.EdgeColorsFace = false` in as many words.

#### One value, not two half-values

- **`Line2D.MarkerFaceColor` / `MarkerEdgeColor`** were read only when the
  neighboring `*Spec` was in its unset mode, where a positive alpha meant
  "explicit" and a zero alpha meant "inherit the line color". That is exactly
  what `MarkerColorSpec` encodes, so the plain fields are gone and
  `MarkerColorDefault` now resolves like `MarkerColorAuto`. The mode is still
  distinct from `MarkerColorAuto`, because only an unset spec is seeded from the
  `lines.marker*` rcParams.
- **`Dashes` + `DashUnits`** (on both `Line2D` and `Patch`) became one
  `DashPattern` value, with `PixelDashes` / `MatplotlibDashes` constructors and a
  `Scaled(lineWidth)` reader. This is the pair the opening example was about: a
  dash sequence is meaningless without its units, and assigning the slice alone
  left the units at their pixel zero value.

## Ownership is a property of the family, not the field

Applying the question field-by-field gives an incoherent API. The useful unit is
the **family** — the group of fields that together describe one configurable
aspect of an artist. If any member of a family is setter-owned, all of it is.

That rule resolves the awkward cases on its own. `Axes.Title` has no companion
state _of its own_, so the narrow question says "keep the field". But it belongs
to the axes-title family, whose other five members — `titleLocation`, `titleY`,
`titleYSet`, `titlePadPt`, `titleWeight` — are already unexported and
setter-owned, and every one of their setters calls `ensureRCTextDefaults`.
`Title` was the sole exported member and the sole writer that skipped that call.
It joins its family.

### Encapsulated

| Family           | Fields                                                                                        | What the field write skipped                                 |
| ---------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Axes title       | `Title`                                                                                       | `ensureRCTextDefaults` (rest of the family already private)  |
| Axes labels      | `XLabel`, `YLabel`                                                                            | `ensureRCTextDefaults`                                       |
| Figure suptitles | `SupTitle`, `SupXLabel`, `SupYLabel`                                                          | family consistency (these also had two setter spellings)     |
| Widget value     | `Slider.Value`, `TextBox.Value`, `RadioButtons.Active`, `RangeSlider.Low`, `RangeSlider.High` | clamp/normalize, caret and selection, the on-change callback |

Each gained a reader of the same name (`ax.Title()`, `slider.Value()`).

`Figure` had accumulated **two** setter spellings per field —
`SetSuptitle`/`SetSupTitle` — from the earlier getter-rename bullet. Only the
Go-cased spelling survives.

### Field-owned: the setter is what goes

`AnchoredTextBox.Locator` and `Legend.Locator` sit in a flat block of plain style
fields (`Location`, `Padding`, `Inset`, …) with no unexported companions and no
invariant. Their `SetLocator` methods were pure aliases, and were deleted.

## Cases the audit flagged that are not duplication

- **`Axes.XScale` / `Axes.YScale`.** `SetXScale(name string, …)` looks up a
  registered scale by name; the field holds the resulting `transform.Scale`
  object and is rewritten continuously by the autoscale and limit machinery
  (`replaceScaleDomain`, `toggleInvertedScale`). Two different operations that
  happen to share a noun. Both stay.
- **`Line2D.MarkEvery`.** `SetMarkEvery` did not write `MarkEvery` (an `int`) at
  all; it writes `MarkEverySpec`. The name was simply wrong, and is now
  `SetMarkEverySpec`.
- **`Axis.TickDirection`.** The field is already the typed `TickDirection`
  constant from the enum pass; only the setter still took a raw `string`. It is
  replaced by `ParseTickDirection(string) (TickDirection, error)` at the rc and
  `TickParams` boundaries, and the typed field is the writer. This closes a
  raw-string enum the previous bullet missed.

## A note on `SetStale`

Several setters call `SetStale(true)`, which looks like a cache invariant a field
write would break. It is not: `Stale()` has no consumer anywhere in the render
path. It mirrors matplotlib's `Artist.stale` as observable metadata, and no
drawing decision reads it. It therefore carries no weight in the classification
above — worth knowing before treating it as a reason to encapsulate something.

The 1721 exported fields with no setter were not audited individually. Many are
genuinely plain configuration, but the audit cannot distinguish those from fields
that _should_ have grown a companion and never did. The signal to watch for is a
new unexported `xSet`/`xUnits` sibling appearing next to an exported field: by
the finding above, that sibling should be an `optional.Value` on the field
itself, not a separate flag.
