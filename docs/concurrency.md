# Concurrency

matplotlib-go separates synchronized process-wide configuration from mutable
plot state. Registry and configuration functions documented as concurrency-safe
may be called from multiple goroutines. Figures, axes, artists, renderers, and
the stateful pyplot view require ownership or external synchronization.

## Global state

The runtime rc state in `style` is process-wide. Its accessors and mutators
serialize memory access, and a new figure snapshots the current defaults.
Changing rc state does not rewrite existing figures.

`style.PushContext` and `pyplot.RCContext` use a single process-wide LIFO stack.
Contexts must be restored in nesting order. Do not overlap rc contexts between
independent goroutines unless the application serializes the complete
push/use/restore lifetime. Prefer explicit figure or axes style options when
goroutines need independent styles. Treat exported baseline values such as
`style.Default` as read-only.

The public registration and lookup operations for colormaps, color sequences,
themes, projections, unit converters, and backends synchronize their registry
maps. A lookup returns a value or factory snapshot; it does not make an object
created from that value safe for concurrent mutation. Do not reassign exported
default-registry variables while another goroutine uses them.

The pyplot current-figure registry also serializes its own bookkeeping. That
lock protects which figure or axes is current, not the returned figure, axes,
manager, or artist. Concurrent pyplot calls can still interfere logically by
changing shared current state. Prefer the explicit `core.Figure` and
`core.Axes` APIs for concurrent work.

## Figures, axes, artists, and renderers

A figure and its axes and artists form one mutable object graph. They are not
safe for concurrent mutation, or for mutation concurrent with drawing. Give the
graph to one goroutine at a time, or guard every read, mutation, layout, and
draw with the same application-owned lock.

Independent figures may be built and rendered concurrently when they do not
share mutable artists or renderers. Unless a backend explicitly documents a
stronger guarantee, use a renderer from only one goroutine at a time. GUI
backends may additionally require a particular event-loop or main goroutine;
their package documentation defines those constraints.

Callbacks run under the rules of the component invoking them. A callback that
mutates its figure must not race an application draw or another callback.

## Typical ownership

```go
var figureMu sync.Mutex

figureMu.Lock()
ax.Plot(x, y)
core.DrawFigure(fig, renderer)
figureMu.Unlock()
```

Keeping construction, mutation, and drawing on one goroutine is usually
simpler than locking. The mutex pattern is useful when an event loop and a data
producer share a figure.
