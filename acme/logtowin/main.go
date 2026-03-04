package main

// go build

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	"9fans.net/go/acme"
)

// - `cmd/logtowin` A simple program to log stdin to a window. Useful in
// shell scripts if the filesystem implementation isn't mounted.

var filename = flag.String("f", "+Log", "the filename in the current directory to update")
var debug = flag.Bool("d", false, "set for verbose debugging")

// need a wrapper for win to make it into a writable.

func main() {
	flag.Parse()
	if !*debug {
		log.SetOutput(io.Discard)
	}
	log.Println("hi", flag.Args())

	// have a single arg

	log.Println("filename", *filename)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("can't get current directory: %v", err)
	}

	targetfile := filepath.Join(cwd, *filename)
	log.Println(targetfile)

	wins, err := acme.Windows()
	if err != nil {
		log.Fatalf("acme.Windows fauked: %v", err)
	}

	var win *acme.Win
	for _, wi := range wins {
		if wi.Name == targetfile {
			win, err = acme.Open(wi.ID, nil)
			if err != nil {
				log.Fatalf("acme.Open failed to (re)open %s: %v", targetfile, err)
			}
			break
		}
	}
	if win == nil {
		win, err = acme.New()
		if err != nil {
			log.Fatalf("acme.New failed on %s: %v", targetfile, err)
		}
		if err := win.Name("%s", targetfile); err != nil {
			log.Fatalf("win.Name failed on %s: %v", targetfile, err)
		}
	}

	ww := win.NewWriter("body")
	if _, err := io.Copy(ww, os.Stdin); err != nil {
		log.Fatalf("can't copy stdin to %s: %v", targetfile, err)
	}
}
