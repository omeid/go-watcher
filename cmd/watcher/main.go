package main

import (
	"log"
	"os"

	watcher "github.com/canthefason/go-watcher"
)

func main() {
	params, err := watcher.PrepareArgs(os.Args[1:])

	if err != nil {
		log.Fatal(err)
	}

	w, err := watcher.MustRegisterWatcher(params)
	if err != nil {
		log.Fatal(err)
	}

	defer w.Close()

	r := watcher.NewRunner()

	// wait for build and run the binary with given params
	go r.Run(params)
	b := watcher.NewBuilder(w, r)

	// build given package
	go func() {
		err := b.Build(params)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// listen for further changes
	go func() {
		err := w.Watch()
		if err != nil {
			log.Fatal(err)
		}
	}()

	r.Wait()
}
