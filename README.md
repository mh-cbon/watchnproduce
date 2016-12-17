# watchnproduce

[![GoDoc](https://godoc.org/github.com/mh-cbon/watchnproduce?status.svg)](https://godoc.org/github.com/mh-cbon/watchnproduce)

Watch resources, produce results.

## Install

```sh
go get github.com/mh-cbon/watchnproduce
glide install github.com/mh-cbon/watchnproduce
```

## Example

```go
package main

import (
	"github.com/mh-cbon/watchnproduce"
	producer "github.com/mh-cbon/watchnproduce/template"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var w *watchnproduce.Watcher

func main() {

	w = watchnproduce.NewWatcher()
	w.LogFunc = log.Printf

	inputFile := w.NewInput(producer.Producer("one"))
	inputFile.AddFiles("demo/tpl/a.tpl", "demo/tpl/b.tpl")

	go w.Run()

	go func() {
		<-time.After(2 * time.Second)
		t, err := inputFile.GetResult()
		if err == nil {
			log.Println(
				t.(*template.Template).ExecuteTemplate(os.Stdout, "a", nil),
			)
		}
	}()

	go func() {
		<-time.After(4 * time.Second)
		ioutil.WriteFile("demo/tpl/a.tpl", []byte("{{BUUUGGG}}"), os.ModePerm)
	}()

	go func() {
		<-time.After(5 * time.Second)
		t, err := inputFile.GetResult()
		if err == nil {
			log.Println(
				t.(*template.Template).ExecuteTemplate(os.Stdout, "a", nil),
			)
		}
	}()

	go func() {
		<-time.After(6 * time.Second)
		ioutil.WriteFile("demo/tpl/a.tpl", []byte("{{define \"a\"}}This a template.\n{{end}}"), os.ModePerm)
	}()

	go func() {
		<-time.After(8 * time.Second)
		t, err := inputFile.GetResult()
		if err == nil {
			log.Println(
				t.(*template.Template).ExecuteTemplate(os.Stdout, "a", nil),
			)
		}
	}()

	make(chan bool) <- true
}
```
