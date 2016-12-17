package main

import (
	"github.com/mh-cbon/watchnproduce"
	"github.com/mh-cbon/watchnproduce/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var w *watchnproduce.Watcher

func main() {

	w = watchnproduce.NewWatcher()
	w.LogFunc = log.Printf

	inputFile := w.NewInput(template.Producer("one"))
	inputFile.AddFiles("demo/tpl/a.tpl", "demo/tpl/b.tpl")

	inputGlob := w.NewInput(template.Producer("two"))
	inputGlob.AddGlob("demo/tpl/*.tpl")

	inputMixed := w.NewInput(template.Producer("3"))
	inputMixed.AddGlob("demo/tpl/c*.tpl")
	inputMixed.AddFiles("demo/tpl/a.tpl", "demo/tpl/b.tpl")

	inputBuggy := w.NewInput(template.Producer("4"))
	inputBuggy.AddGlob("demo/hhhh*.tpl")

	go w.Run()

	go func() {
		<-time.After(5 * time.Second)
		ioutil.WriteFile("demo/tpl/c.tpl", []byte("{{BUUUGGG}}"), os.ModePerm)
	}()

	go func() {
		<-time.After(10 * time.Second)
		ioutil.WriteFile("demo/tpl/c.tpl", []byte("OK"), os.ModePerm)
	}()

	make(chan bool) <- true
}
