package template

import (
	"github.com/mh-cbon/watchnproduce"
	"html/template"
)

// Producer of html.template instances.
func Producer(name string) watchnproduce.ProducerFunc {
	return func(pointers []watchnproduce.ResourcePointer) (interface{}, error) {
		t := template.New(name)
		var err error
		for _, p := range pointers {
			if err == nil {
				if filesP, ok := p.(*watchnproduce.FilesPointer); ok {
					t, err = t.ParseFiles(filesP.Files...)
				} else if globsP, ok := p.(*watchnproduce.GlobsPointer); ok {
					for _, g := range globsP.Globs {
						t, err = t.ParseGlob(g)
					}
				}
			}
		}
		return t, err
	}
}
