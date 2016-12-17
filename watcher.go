// Watch n Produce is package
// to watch for resources pointers
// and produce a result anytime the resources changed.
package watchnproduce

import (
	"time"
)

// Watcher keep tracks of registered inputs
// and update their produce regularly.
type Watcher struct {
	Inputs  []*Input
	closed  bool
	LogFunc func(string, ...interface{})
}

func noop(s string, g ...interface{}) {}

func NewWatcher() *Watcher {
	return &Watcher{
		LogFunc: noop,
	}
}

// Run loops over watched inputs,
// detect they changed,
// update their results.
func (w *Watcher) Run() {
	for {
		time.Sleep(time.Millisecond * 500)
		l := len(w.Inputs)
		w.LogFunc("Checking %d items...\n", l)
		w.DoRemoval()
		w.DoUpdate()
		time.Sleep(time.Millisecond * 500)
	}
}

// DoUpdate run a single loop over watched inputs to update them.
func (w *Watcher) DoUpdate() {
	for _, i := range w.Inputs {
		if w.closed {
			return
		}
		newStats, err := i.GetStats()
		if err == nil {
			if w.isSameAs(i.KnownStats, newStats) == false {
				i.Update(newStats)
			}
		} else {
			w.LogFunc("Failed to resolve: %v", err)
		}
	}
}

// DoRemoval run a single loop over watched inputs to remove them.
func (w *Watcher) DoRemoval() {
	deletion := make([]*Input, 0)
	for _, i := range w.Inputs {
		if i.MarkedForDeletion {
			deletion = append(deletion, i)
		}
	}
	for i := 0; i < len(deletion); i++ {
		index := -1
		for e := 0; e < len(w.Inputs); e++ {
			if w.Inputs[e] == deletion[i] {
				index = e
				break
			}
		}
		if index > -1 {
			w.Inputs = append(w.Inputs[:index], w.Inputs[index+1:]...)
		}
	}
	if len(deletion) > 0 {
		w.LogFunc("Removed %d items\n", len(deletion))
	}
	deletion = deletion[0:0]
}

// Close the watcher, ends Run loop.
func (w *Watcher) Close() {
	w.LogFunc("Closing watcher...")
	w.closed = true
}

// Tells if two list of stats are identical.
func (w *Watcher) isSameAs(old Stats, newRes Stats) bool {
	if len(old) != len(newRes) {
		return false
	}
	for _, o := range old {
		if newRes.ContainsSame(o) == false {
			return false
		}
	}
	for _, n := range newRes {
		if old.ContainsSame(n) == false {
			return false
		}
	}
	return true
}

// NewInput create and add a new input to the watch list.
func (w *Watcher) NewInput(p ProducerFunc) *Input {
	ret := &Input{
		Producer: p,
		LogFunc:  w.LogFunc,
	}
	w.Inputs = append(w.Inputs, ret)
	return ret
}
