package watchnproduce

import (
	"sync"
)

// ProducerFunc is a function type which knows
// how to produce useful results given a list of PointResource
type ProducerFunc func(pointers []ResourcePointer) (interface{}, error)

// An Input registers
// a list of PointerResource to tell if a result needs to be produced
// and a ProducerFunc to produce a result.
type Input struct {
	lock              sync.RWMutex
	Pointers          []ResourcePointer
	KnownStats        Stats
	LogFunc           func(string, ...interface{})
	Producer          ProducerFunc
	LastValue         interface{}
	LastErr           error
	MarkedForDeletion bool
	OnChanged         func(i *Input)
	OnErrored         func(i *Input, err error)
}

// AddPointer adds a ResourcePointer to this Input.
func (i *Input) AddPointer(p ResourcePointer) *Input {
	i.lock.Lock()
	i.Pointers = append(i.Pointers, p)
	i.lock.Unlock()
	return i
}

// AddFiles adds a FilesPointer, a pointer to a list of file paths.
func (i *Input) AddFiles(files ...string) *Input {
	return i.AddPointer(NewFilesPointer(files...))
}

// AddGlob adds a GlobsPointer, to a list of glob path.
func (i *Input) AddGlob(globs ...string) *Input {
	return i.AddPointer(NewGlobsPointer(globs...))
}

// GetStats browse registered ResourcePointer
// and returns the new stats for each resource found.
func (i *Input) GetStats() (Stats, error) {
	ret := make([]Stat, 0)
	var err error
	var stats []Stat
	for _, v := range i.Pointers {
		stats, err = v.GetStats()
		if err == nil {
			ret = append(ret, stats...)
		}
	}
	if err != nil {
		i.LogFunc("Failed to resolve: %v", err)
	}
	return Stats(ret), err
}

// Update produces a new produce and saves its result.
func (i *Input) Update(newStats Stats) {
	i.lock.Lock()
	i.LogFunc("Updating resource...")
	value, err := i.Producer(i.Pointers)
	if err != nil {
		i.LogFunc("Failed to produce: %v", err)
	} else if err == nil && i.LastErr != nil {
		i.LogFunc("Resource produced successfully")
	}
	if i.LastValue == nil || err == nil {
		i.LogFunc("Resource updated...")
		i.KnownStats = newStats
		i.LastValue = value
		i.LastErr = err
		if i.OnChanged != nil && err == nil {
			i.OnChanged(i)
		}
		if i.OnErrored != nil && err != nil {
			i.OnErrored(i, err)
		}
	} else {
		if i.OnErrored != nil && err != nil {
			i.OnErrored(i, err)
		}
		i.LogFunc("Resource not updated LastErr: %v, err: %v", i.LastErr, err)
	}
	i.lock.Unlock()
}

// GetResult provide the last produced result.
func (i *Input) GetResult() (interface{}, error) {
	i.lock.RLock()
	value := i.LastValue
	err := i.LastErr
	i.lock.RUnlock()
	return value, err
}

// MarkForDeletion mark this input so the watcher can garbage it.
func (i *Input) MarkForDeletion() {
	i.lock.Lock()
	i.MarkedForDeletion = true
	i.lock.Unlock()
}
