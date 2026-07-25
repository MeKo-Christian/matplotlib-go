package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cwbudde/matplotlib-go/dates"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/ticker"
)

var timeType = reflect.TypeOf(time.Time{})

type AxisInfo struct {
	Locator        ticker.Locator
	Formatter      ticker.Formatter
	MinorLocator   ticker.Locator
	MinorFormatter ticker.Formatter
}

type UnitsConverter interface {
	Convert(value any) (float64, error)
	AxisInfo(values []float64) AxisInfo
}

type UnitsConverterFactory func() UnitsConverter

type UnitsRegistry struct {
	mu        sync.RWMutex
	converter map[reflect.Type]UnitsConverterFactory
}

func NewUnitsRegistry() *UnitsRegistry {
	return &UnitsRegistry{
		converter: make(map[reflect.Type]UnitsConverterFactory),
	}
}

func (r *UnitsRegistry) Register(sample any, factory UnitsConverterFactory) error {
	if factory == nil {
		return errors.New("unit converter factory cannot be nil")
	}
	typ := reflect.TypeOf(sample)
	if typ == nil {
		return errors.New("unit converter sample cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.converter[typ]; exists {
		return fmt.Errorf("unit converter already registered for %v", typ)
	}
	r.converter[typ] = factory
	return nil
}

func (r *UnitsRegistry) MustRegister(sample any, factory UnitsConverterFactory) {
	if err := r.Register(sample, factory); err != nil {
		panic(err)
	}
}

func (r *UnitsRegistry) lookup(typ reflect.Type) UnitsConverterFactory {
	if r == nil || typ == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.converter[typ]
}

var DefaultUnitsRegistry = NewUnitsRegistry()

func RegisterUnitConverter(sample any, factory UnitsConverterFactory) error {
	return DefaultUnitsRegistry.Register(sample, factory)
}

func MustRegisterUnitConverter(sample any, factory UnitsConverterFactory) {
	DefaultUnitsRegistry.MustRegister(sample, factory)
}

type unitAxisKind uint8

const (
	unitAxisNone unitAxisKind = iota
	unitAxisDate
	unitAxisCategory
	unitAxisCustom
)

type axisUnitsState struct {
	kind       unitAxisKind
	customType reflect.Type
	converter  UnitsConverter
	info       AxisInfo
	location   *time.Location
	categories categoryAxisState
	applied    map[*Axis]AxisInfo
}

func (s *axisUnitsState) name() string {
	if s == nil {
		return "numeric"
	}
	switch s.kind {
	case unitAxisDate:
		return "date"
	case unitAxisCategory:
		return "categorical"
	case unitAxisCustom:
		if s.customType != nil {
			return s.customType.String()
		}
		return "custom"
	default:
		return "numeric"
	}
}

func (s *axisUnitsState) scaleCompatible(name string) bool {
	if s == nil || s.kind == unitAxisNone {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(name), "linear")
}

type categoryAxisState struct {
	order     []string
	positions map[string]float64
}

func (s *categoryAxisState) convert(label string) float64 {
	if s.positions == nil {
		s.positions = make(map[string]float64)
	}
	if pos, ok := s.positions[label]; ok {
		return pos
	}
	pos := float64(len(s.order))
	s.order = append(s.order, label)
	s.positions[label] = pos
	return pos
}

func (s *categoryAxisState) axisInfo() AxisInfo {
	ticks := make([]float64, len(s.order))
	copyLabels := append([]string(nil), s.order...)
	for i := range ticks {
		ticks[i] = float64(i)
	}
	return AxisInfo{
		Locator:      ticker.FixedLocator{TicksList: ticks},
		Formatter:    ticker.FixedFormatter{Labels: copyLabels},
		MinorLocator: nil,
	}
}

// PlotDate plots date-time x-values against numeric y-values and configures the
// x-axis with date locators/formatters.
func (a *Axes) PlotDate(xVals []time.Time, yVals []float64, opts ...PlotOptions) (*Line2D, error) {
	return a.Plot(xVals, yVals, opts...)
}

type unitAxisSnapshot struct {
	axis *Axis
	info AxisInfo
}

type unitConversionTransaction struct {
	xRoot      *Axes
	yRoot      *Axes
	xUnits     *axisUnitsState
	yUnits     *axisUnitsState
	axisStates []unitAxisSnapshot
	active     bool
}

func (a *Axes) beginUnitConversion() *unitConversionTransaction {
	tx := &unitConversionTransaction{active: true}
	if a == nil {
		return tx
	}

	tx.xRoot = a.axisRoot(true)
	tx.yRoot = a.axisRoot(false)
	if tx.xRoot != nil {
		tx.xUnits = tx.xRoot.xUnits
		tx.xRoot.xUnits = cloneAxisUnitsState(tx.xUnits)
		tx.snapshotAxis(tx.xRoot.XAxis)
		tx.snapshotAxis(tx.xRoot.XAxisTop)
	}
	if tx.yRoot != nil {
		tx.yUnits = tx.yRoot.yUnits
		tx.yRoot.yUnits = cloneAxisUnitsState(tx.yUnits)
		tx.snapshotAxis(tx.yRoot.YAxis)
		tx.snapshotAxis(tx.yRoot.YAxisRight)
	}
	return tx
}

func (tx *unitConversionTransaction) snapshotAxis(axis *Axis) {
	if tx == nil || axis == nil {
		return
	}
	for _, snapshot := range tx.axisStates {
		if snapshot.axis == axis {
			return
		}
	}
	tx.axisStates = append(tx.axisStates, unitAxisSnapshot{
		axis: axis,
		info: AxisInfo{
			Locator:        axis.Locator,
			Formatter:      axis.Formatter,
			MinorLocator:   axis.MinorLocator,
			MinorFormatter: axis.MinorFormatter,
		},
	})
}

func (tx *unitConversionTransaction) commit() {
	if tx != nil {
		tx.active = false
	}
}

func (tx *unitConversionTransaction) rollback() {
	if tx == nil || !tx.active {
		return
	}
	tx.active = false
	if tx.xRoot != nil {
		tx.xRoot.xUnits = tx.xUnits
	}
	if tx.yRoot != nil {
		tx.yRoot.yUnits = tx.yUnits
	}
	for _, snapshot := range tx.axisStates {
		snapshot.axis.Locator = snapshot.info.Locator
		snapshot.axis.Formatter = snapshot.info.Formatter
		snapshot.axis.MinorLocator = snapshot.info.MinorLocator
		snapshot.axis.MinorFormatter = snapshot.info.MinorFormatter
	}
}

func cloneAxisUnitsState(state *axisUnitsState) *axisUnitsState {
	if state == nil {
		return nil
	}
	cloned := *state
	cloned.categories.order = append([]string(nil), state.categories.order...)
	if state.categories.positions != nil {
		cloned.categories.positions = make(map[string]float64, len(state.categories.positions))
		for label, position := range state.categories.positions {
			cloned.categories.positions[label] = position
		}
	}
	if state.applied != nil {
		cloned.applied = make(map[*Axis]AxisInfo, len(state.applied))
		for axis, info := range state.applied {
			cloned.applied[axis] = info
		}
	}
	if state.kind == unitAxisCustom {
		if factory := DefaultUnitsRegistry.lookup(state.customType); factory != nil {
			cloned.converter = factory()
		}
	}
	return &cloned
}

func (a *Axes) applyCategoricalBarValueLocator(categoryIsX bool) {
	if a == nil {
		return
	}
	categoryState := a.axisRoot(categoryIsX).unitState(categoryIsX)
	if categoryState == nil || categoryState.kind != unitAxisCategory {
		return
	}

	valueRoot := a.axisRoot(!categoryIsX)
	if valueRoot == nil {
		return
	}
	if categoryIsX {
		applyAutoLocatorIfDefault(valueRoot.YAxis)
		applyAutoLocatorIfDefault(valueRoot.YAxisRight)
		return
	}
	applyAutoLocatorIfDefault(valueRoot.XAxis)
	applyAutoLocatorIfDefault(valueRoot.XAxisTop)
}

func applyAutoLocatorIfDefault(axis *Axis) {
	if axis == nil {
		return
	}
	if _, ok := axis.Locator.(ticker.LinearLocator); ok {
		axis.Locator = ticker.AutoLocator{}
	}
}

func (a *Axes) convertValues(values any, isX bool) ([]float64, error) {
	slice, elemType, err := sliceValue(values)
	if err != nil {
		return nil, err
	}
	if slice.Len() == 0 {
		return nil, nil
	}

	if factory := DefaultUnitsRegistry.lookup(elemType); factory != nil {
		return a.convertCustomValues(slice, elemType, factory, isX)
	}
	if elemType == timeType {
		return a.convertDateValues(slice, isX)
	}
	if elemType.Kind() == reflect.String {
		return a.convertCategoryValues(slice, isX)
	}
	if isNumericType(elemType) {
		return numericValues(slice), nil
	}

	state := a.axisRoot(isX).unitState(isX)
	if state != nil {
		switch state.kind {
		case unitAxisDate, unitAxisCategory, unitAxisCustom:
			if isNumericType(elemType) {
				return numericValues(slice), nil
			}
		}
	}

	return nil, fmt.Errorf("unsupported plot values type %v", elemType)
}

func (a *Axes) convertCustomValues(slice reflect.Value, elemType reflect.Type, factory UnitsConverterFactory, isX bool) ([]float64, error) {
	root := a.axisRoot(isX)
	state := root.unitState(isX)
	if state == nil {
		state = &axisUnitsState{
			kind:       unitAxisCustom,
			customType: elemType,
			converter:  factory(),
		}
		root.setUnitState(isX, state)
	}
	if state.kind != unitAxisCustom || state.customType != elemType || state.converter == nil {
		return nil, fmt.Errorf("axis already configured for %s units", state.name())
	}

	out := make([]float64, slice.Len())
	for i := range out {
		v, err := state.converter.Convert(slice.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	state.info = state.converter.AxisInfo(out)
	root.refreshUnitAxis(isX)
	return out, nil
}

func (a *Axes) convertDateValues(slice reflect.Value, isX bool) ([]float64, error) {
	root := a.axisRoot(isX)
	state := root.unitState(isX)
	if state == nil {
		state = &axisUnitsState{kind: unitAxisDate}
		root.setUnitState(isX, state)
	}
	if state.kind != unitAxisDate {
		return nil, fmt.Errorf("axis already configured for %s units", state.name())
	}

	out := make([]float64, slice.Len())
	for i := range out {
		timestamp, ok := slice.Index(i).Interface().(time.Time)
		if !ok {
			return nil, fmt.Errorf("date axis expected time.Time values")
		}
		if state.location == nil {
			state.location = timestamp.Location()
		}
		out[i] = dates.Date2Num(timestamp)
	}
	root.refreshUnitAxis(isX)
	return out, nil
}

func (a *Axes) convertCategoryValues(slice reflect.Value, isX bool) ([]float64, error) {
	root := a.axisRoot(isX)
	state := root.unitState(isX)
	if state == nil {
		state = &axisUnitsState{kind: unitAxisCategory}
		root.setUnitState(isX, state)
	}
	if state.kind != unitAxisCategory {
		return nil, fmt.Errorf("axis already configured for %s units", state.name())
	}

	out := make([]float64, slice.Len())
	for i := range out {
		out[i] = state.categories.convert(slice.Index(i).String())
	}
	root.refreshUnitAxis(isX)
	return out, nil
}

func (a *Axes) axisRoot(isX bool) *Axes {
	if isX {
		return a.xScaleRoot()
	}
	return a.yScaleRoot()
}

func (a *Axes) unitState(isX bool) *axisUnitsState {
	if a == nil {
		return nil
	}
	if isX {
		return a.xUnits
	}
	return a.yUnits
}

func (a *Axes) setUnitState(isX bool, state *axisUnitsState) {
	if a == nil {
		return
	}
	if isX {
		a.xUnits = state
	} else {
		a.yUnits = state
	}
}

func (a *Axes) refreshUnitAxis(isX bool) {
	root := a.axisRoot(isX)
	if root == nil {
		return
	}
	state := root.unitState(isX)
	if state == nil || state.kind == unitAxisNone {
		return
	}

	var primary, secondary *Axis
	var minVal, maxVal float64
	if isX {
		primary, secondary = root.XAxis, root.XAxisTop
		minVal, maxVal = currentScaleDomain(root.XScale)
	} else {
		primary, secondary = root.YAxis, root.YAxisRight
		minVal, maxVal = currentScaleDomain(root.YScale)
	}
	state.applyAxisInfo(primary, secondary, state.axisInfo(minVal, maxVal))
}

func (s *axisUnitsState) axisInfo(minVal, maxVal float64) AxisInfo {
	switch s.kind {
	case unitAxisDate:
		// The date.converter rcParam switches the default formatter family,
		// mirroring matplotlib's _SwitchableDateConverter which consults the
		// rc value on every axisinfo call.
		var formatter ticker.Formatter = dates.AutoDateFormatter{Min: minVal, Max: maxVal, Location: s.location}
		if strings.EqualFold(strings.TrimSpace(style.CurrentDefaults().Date.Converter), "concise") {
			formatter = dates.ConciseDateFormatter{Location: s.location}
		}
		return AxisInfo{
			Locator:        dates.DateLocator{Location: s.location},
			Formatter:      formatter,
			MinorLocator:   nil,
			MinorFormatter: nil,
		}
	case unitAxisCategory:
		return s.categories.axisInfo()
	case unitAxisCustom:
		return s.info
	default:
		return AxisInfo{}
	}
}

func (s *axisUnitsState) applyAxisInfo(primary, secondary *Axis, info AxisInfo) {
	s.applyAxisInfoToAxis(primary, info)
	s.applyAxisInfoToAxis(secondary, info)
}

func (s *axisUnitsState) applyAxisInfoToAxis(axis *Axis, info AxisInfo) {
	if axis == nil {
		return
	}

	if s.applied == nil {
		s.applied = make(map[*Axis]AxisInfo)
	}
	previous, hadPrevious := s.applied[axis]
	applied := previous

	if info.Locator != nil && canApplyUnitAxisInfo(axis.Locator, previous.Locator, hadPrevious, isDefaultMajorLocator) {
		axis.Locator = info.Locator
		applied.Locator = info.Locator
	}
	if info.Formatter != nil && canApplyUnitAxisInfo(axis.Formatter, previous.Formatter, hadPrevious, isDefaultMajorFormatter) {
		axis.Formatter = info.Formatter
		applied.Formatter = info.Formatter
	}
	if canApplyUnitAxisInfo(axis.MinorLocator, previous.MinorLocator, hadPrevious, isDefaultMinorLocator) {
		axis.MinorLocator = info.MinorLocator
		applied.MinorLocator = info.MinorLocator
	}
	if canApplyUnitAxisInfo(axis.MinorFormatter, previous.MinorFormatter, hadPrevious, isDefaultMinorFormatter) {
		axis.MinorFormatter = info.MinorFormatter
		applied.MinorFormatter = info.MinorFormatter
	}

	if applied.Locator != nil || applied.Formatter != nil || applied.MinorLocator != nil || applied.MinorFormatter != nil {
		s.applied[axis] = applied
	}
}

func canApplyUnitAxisInfo[T any](current, previous T, hadPrevious bool, isDefault func(T) bool) bool {
	if hadPrevious {
		return reflect.DeepEqual(current, previous) || isDefault(current)
	}
	return isDefault(current)
}

func isDefaultMajorLocator(locator ticker.Locator) bool {
	switch locator.(type) {
	case nil, ticker.AutoLocator, ticker.LinearLocator:
		return true
	default:
		return false
	}
}

func isDefaultMajorFormatter(formatter ticker.Formatter) bool {
	switch formatter.(type) {
	case nil, ticker.ScalarFormatter:
		return true
	default:
		return false
	}
}

func isDefaultMinorLocator(locator ticker.Locator) bool {
	return locator == nil
}

func isDefaultMinorFormatter(formatter ticker.Formatter) bool {
	return formatter == nil
}

func sliceValue(values any) (reflect.Value, reflect.Type, error) {
	v := reflect.ValueOf(values)
	if !v.IsValid() {
		return reflect.Value{}, nil, errors.New("plot values cannot be nil")
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return reflect.Value{}, nil, fmt.Errorf("plot values must be a slice or array, got %T", values)
	}
	return v, v.Type().Elem(), nil
}

func numericValues(v reflect.Value) []float64 {
	out := make([]float64, v.Len())
	for i := range out {
		out[i] = numericValue(v.Index(i))
	}
	return out
}

func numericValue(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Convert(reflect.TypeOf(float64(0))).Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint())
	default:
		return 0
	}
}

func isNumericType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	default:
		return false
	}
}
