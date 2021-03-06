package watchnproduce

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResourcePointer returns a list of stats.
type ResourcePointer interface {
	GetStats() ([]Stat, error)
}

// Stat identify a resource with and Id, and provide its LastMod.
type Stat struct {
	ID  string
	Mod int64
}

// Stats is an useful wrapper of []Stat.
type Stats []Stat

// Contains tells if given Stat ID exists into
// the current list of Stats.
func (rlist Stats) Contains(s Stat) bool {
	for _, r := range rlist {
		if r.ID == s.ID {
			return true
		}
	}
	return false
}

// ContainsSame tells if given Stat ID and Mod
// exists in the current list of Stats.
func (rlist Stats) ContainsSame(s Stat) bool {
	for _, r := range rlist {
		if r.ID == s.ID && r.Mod == s.Mod {
			return true
		}
	}
	return false
}
func (rlist Stats) String() string {
	ret := ""
	for _, s := range rlist {
		ret += fmt.Sprintf("%v %v\n", s.ID, s.Mod)
	}
	return ret
}

// FilesPointer registers a list of file paths.
type FilesPointer struct {
	Files []string
}

// NewFilesPointer creates a new FilesPointer instance.
func NewFilesPointer(files ...string) *FilesPointer {
	return &FilesPointer{
		Files: files,
	}
}

// GetStats browse each file path,
// recursively for a directory,
// and returns theirs Stats.
func (p *FilesPointer) GetStats() ([]Stat, error) {
	ret := make([]Stat, 0)
	var err error
	for _, f := range p.Files {
		if err == nil {
			err = filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
				if err == nil {
					ret = append(ret, Stat{
						ID:  path,
						Mod: info.ModTime().Unix(),
					})
				}
				return err
			})
		}
	}
	return ret, err
}

//GlobsPointer registers a list of glob paths.
type GlobsPointer struct {
	Globs []string
}

// NewGlobsPointer creates a new instance of GlobsPointer.
func NewGlobsPointer(globs ...string) *GlobsPointer {
	return &GlobsPointer{
		Globs: globs,
	}
}

// GetStats browse each glob paths
// and returns their Stat.
func (p *GlobsPointer) GetStats() ([]Stat, error) {
	ret := make([]Stat, 0)
	var err error
	var items []string
	for _, f := range p.Globs {
		items, err = filepath.Glob(f)
		if err == nil {
			for _, item := range items {
				if err == nil {
					err = filepath.Walk(item, func(path string, info os.FileInfo, err error) error {
						if err == nil {
							ret = append(ret, Stat{
								ID:  path,
								Mod: info.ModTime().Unix(),
							})
						}
						return err
					})
				}
			}
		}
	}
	return ret, err
}
