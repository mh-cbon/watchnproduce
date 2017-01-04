package watchnproduce_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mh-cbon/watchnproduce"
)

type MemoryPointer struct {
	StatstoReturn []watchnproduce.Stat
	ErrToReturn   error
}

func NewMemoryPointer() *MemoryPointer {
	return &MemoryPointer{
		StatstoReturn: make([]watchnproduce.Stat, 0),
		ErrToReturn:   nil,
	}
}

func (m *MemoryPointer) AddStat(id string, mod int64) {
	s := watchnproduce.Stat{ID: id, Mod: mod}
	m.StatstoReturn = append(m.StatstoReturn, s)
}

func (m *MemoryPointer) SetStat(id string, mod int64) {
	for e, i := range m.StatstoReturn {
		if i.ID == id {
			m.StatstoReturn[e].Mod = mod
			break
		}
	}
}

func (m *MemoryPointer) RmStat(id string) {
	index := -1
	for e, i := range m.StatstoReturn {
		if i.ID == id {
			index = e
			break
		}
	}
	if index > -1 {
		m.StatstoReturn = append(m.StatstoReturn[:index], m.StatstoReturn[index+1:]...)
	}
}

func (m *MemoryPointer) SetErr(s string) {
	if s == "" {
		m.ErrToReturn = nil
	} else {
		m.ErrToReturn = errors.New(s)
	}
}

func (m *MemoryPointer) GetStats() ([]watchnproduce.Stat, error) {
	b := make([]watchnproduce.Stat, 0)
	b = append(b, m.StatstoReturn...)
	return b, m.ErrToReturn
}

func noopProducer(pointers []watchnproduce.ResourcePointer) (interface{}, error) {
	return struct{}{}, nil
}
func errProducer(pointers []watchnproduce.ResourcePointer) (interface{}, error) {
	return struct{}{}, errors.New("south")
}

type someRes struct {
	ID string
}

func resProducer(pointers []watchnproduce.ResourcePointer) (interface{}, error) {
	return someRes{ID: "1"}, nil
}

type recoverErrProducer struct {
	i int
}

func (r *recoverErrProducer) produce(pointers []watchnproduce.ResourcePointer) (interface{}, error) {
	if r.i == 0 {
		r.i++
		return nil, errors.New("south")
	}
	return someRes{ID: "1"}, nil
}

type errRecoverProducer struct {
	i int
}

func (r *errRecoverProducer) produce(pointers []watchnproduce.ResourcePointer) (interface{}, error) {
	if r.i == 0 {
		r.i++
		return someRes{ID: "1"}, nil
	}
	return nil, errors.New("south")
}

func eq(t *testing.T, expected, got int, reason string) {
	if got != expected {
		t.Errorf(
			"%v,\n expected=%v got=%v",
			reason, expected, got)
	}
}

func eqErr(t *testing.T, expected, got error) {
	if got == nil {
		t.Errorf(
			"Errors does not match,\n expected=%v got=%v",
			expected.Error(), got)
	} else if got.Error() != expected.Error() {
		t.Errorf(
			"Errors does not match,\n expected=%v got=%v",
			expected.Error(), got.Error())
	}
}

func TestChange(t *testing.T) {
	changed := 0
	expected := 2
	w := watchnproduce.NewWatcher()
	i := w.NewInput(noopProducer)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	w.DoUpdate()
	pointer.SetStat("1", 1)
	w.DoUpdate()
	eq(t, expected, changed, "watcher should detect a change")
}

func TestRun(t *testing.T) {
	changed := 0
	expected := 2
	w := watchnproduce.NewWatcher()
	w.Interval = 5
	i := w.NewInput(noopProducer)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	go w.Run()
	<-time.After(20 * time.Millisecond)
	pointer.SetStat("1", 1)
	<-time.After(20 * time.Millisecond)
	eq(t, expected, changed, "watcher should detect a change")
	w.Close()
}

func TestClose(t *testing.T) {
	changed := 0
	expected := 2
	w := watchnproduce.NewWatcher()
	i := w.NewInput(noopProducer)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	w.DoUpdate()
	pointer.SetStat("1", 1)
	w.DoUpdate()
	eq(t, expected, changed, "watcher should detect a change")
	w.Close()
	w.Run() // will quit immediately rather than keep looping
	pointer.SetStat("1", 2)
	w.DoUpdate()
	eq(t, expected, changed, "watcher should not detect a change after Close")
}

func TestErrOnStat(t *testing.T) {
	changed := 0
	errored := 0
	var gotError error
	expectedErrored := 1
	expectedChanged := 1
	expectedError := errors.New("error while stat")
	w := watchnproduce.NewWatcher()
	i := w.NewInput(noopProducer)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	i.OnErrored = func(e *watchnproduce.Input, err error) {
		errored++
		gotError = err
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	w.DoUpdate()
	pointer.SetErr("error while stat")
	w.DoUpdate()
	eq(t, expectedChanged, changed, "watcher should detect a change")
	eq(t, expectedErrored, errored, "watcher should detect an error")
	eqErr(t, expectedError, gotError)
}

func TestDeletion(t *testing.T) {
	w := watchnproduce.NewWatcher()
	i := w.NewInput(noopProducer)
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	eq(t, 1, len(w.Inputs), "watcher should have a resources pointer")
	i.MarkForDeletion()
	w.DoRemoval()
	eq(t, 0, len(w.Inputs), "watcher should not have a resources pointer")
}

func TestStatsComparison(t *testing.T) {
	changed := 0
	w := watchnproduce.NewWatcher()
	i := w.NewInput(noopProducer)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	w.DoUpdate()
	eq(t, 1, changed, "watcher should detect a change")
	pointer.AddStat("2", 0)
	w.DoUpdate()
	eq(t, 2, changed, "watcher should detect a change")
	pointer.SetStat("1", 1)
	w.DoUpdate()
	eq(t, 3, changed, "watcher should detect a change")
	pointer.RmStat("1")
	pointer.AddStat("3", 0)
	w.DoUpdate()
	eq(t, 4, changed, "watcher should detect a change")
}

func TestErrOnProduce(t *testing.T) {
	changed := 0
	errored := 0
	var gotError error
	expectedError := errors.New("south")
	w := watchnproduce.NewWatcher()
	i := w.NewInput(errProducer)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	i.OnErrored = func(e *watchnproduce.Input, err error) {
		errored++
		gotError = err
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	w.DoUpdate()
	eq(t, 0, changed, "watcher should detect a change")
	eq(t, 1, errored, "watcher should detect an error")
	eqErr(t, expectedError, gotError)
}

func TestRecoverErrOnProduce(t *testing.T) {
	changed := 0
	errored := 0
	var gotError error
	expectedError := errors.New("south")
	w := watchnproduce.NewWatcher()
	p := &recoverErrProducer{}
	i := w.NewInput(p.produce)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	i.OnErrored = func(e *watchnproduce.Input, err error) {
		errored++
		gotError = err
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	w.DoUpdate()
	eq(t, 0, changed, "watcher should not detect a change")
	eq(t, 1, errored, "watcher should detect an error")
	eqErr(t, expectedError, gotError)
	w.DoUpdate()
	eq(t, 1, changed, "watcher should detect a change")
	eq(t, 1, errored, "watcher should not detect additional error")
}

func TestErrRecoverOnProduce(t *testing.T) {
	changed := 0
	errored := 0
	var gotError error
	expectedError := errors.New("south")
	w := watchnproduce.NewWatcher()
	p := &errRecoverProducer{}
	i := w.NewInput(p.produce)
	i.OnChanged = func(e *watchnproduce.Input) {
		changed++
	}
	i.OnErrored = func(e *watchnproduce.Input, err error) {
		errored++
		gotError = err
	}
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)
	w.DoUpdate()
	eq(t, 1, changed, "watcher should detect a change")
	eq(t, 0, errored, "watcher should not detect an error")
	pointer.SetStat("1", 1)
	w.DoUpdate()
	eq(t, 1, changed, "watcher should not detect additional change")
	eq(t, 1, errored, "watcher should detect an error")
	eqErr(t, expectedError, gotError)
}

func TestGetResults(t *testing.T) {
	w := watchnproduce.NewWatcher()
	i := w.NewInput(resProducer)
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)

	w.DoUpdate()
	val, err := i.GetResult()
	if val == nil {
		t.Errorf("Result should not be nil")
	}
	if err != nil {
		t.Errorf("Err should be nil")
	}
	if x, ok := val.(someRes); !ok {
		t.Errorf("Result be of type SomeRes, got %T", x)
	} else if x.ID != "1" {
		t.Errorf("Result ID shold eq=%v, got=%v", "1", x.ID)
	}
}

func TestErrProduceGetResults(t *testing.T) {
	w := watchnproduce.NewWatcher()
	p := &recoverErrProducer{}
	i := w.NewInput(p.produce)
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)

	w.DoUpdate()
	val, err := i.GetResult()
	if val != nil {
		t.Errorf("Result should be nil")
	}
	if err == nil {
		t.Errorf("Err should not be nil, got=%v", err)
	}

	w.DoUpdate()
	val, err = i.GetResult()
	if val == nil {
		t.Errorf("Result should not be nil")
	}
	if err != nil {
		t.Errorf("Err should be nil")
	}
	if x, ok := val.(someRes); !ok {
		t.Errorf("Result be of type SomeRes, got %T", x)
	} else if x.ID != "1" {
		t.Errorf("Result ID shold eq=%v, got=%v", "1", x.ID)
	}
}

// ensure that lastValue is always returned on error producing.
func TestRecoverErrProduceGetResults(t *testing.T) {
	w := watchnproduce.NewWatcher()
	p := &errRecoverProducer{}
	i := w.NewInput(p.produce)
	pointer := NewMemoryPointer()
	pointer.AddStat("1", 0)
	i.AddPointer(pointer)

	w.DoUpdate()
	val, err := i.GetResult()
	if val == nil {
		t.Errorf("Result should not be nil")
	}
	if err != nil {
		t.Errorf("Err should be nil")
	}
	if x, ok := val.(someRes); !ok {
		t.Errorf("Result be of type SomeRes, got %T", x)
	} else if x.ID != "1" {
		t.Errorf("Result ID shold eq=%v, got=%v", "1", x.ID)
	}

	w.DoUpdate()
	val, err = i.GetResult()
	if val == nil {
		t.Errorf("Result should not be nil")
	}
	if err != nil {
		t.Errorf("Err should be nil")
	}
	if x, ok := val.(someRes); !ok {
		t.Errorf("Result be of type SomeRes, got %T", x)
	} else if x.ID != "1" {
		t.Errorf("Result ID shold eq=%v, got=%v", "1", x.ID)
	}
}
