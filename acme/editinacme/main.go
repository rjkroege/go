// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Editinacme can be used as $EDITOR in a Unix environment.
//
// Usage:
//
//	editinacme [-nw] <file...>
//
// Editinacme uses the plumber to ask acme to open the file, waits until
// the file's acme window is deleted, and exits. Use the -nw flag to exit
// immediately after opening the file.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"9fans.net/go/acme"
	"9fans.net/go/plan9"
	"9fans.net/go/plumb"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("editinacme: ")

	nowait := flag.Bool("nw", false, "Don't wait for Acme to close the file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: editinacme file\n")
		os.Exit(2)
	}
	flag.Parse()

	// Absolute all of the files.
	files := make(map[string]struct{})
	for _, f := range flag.Args() {
		fp, err := filepath.Abs(f)
		if err != nil {
			log.Fatal(err)
		}
		files[fp] = struct{}{}
	}

	r, err := acme.Log()
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	pathschan := make(chan map[string]struct{})
	go func() {
		fid, err := plumb.Open("edit", plan9.OREAD)
		if err != nil {
			log.Fatalf("can't open plumber: %v", err)
		}
		defer fid.Close()
		brd := bufio.NewReader(fid)
		m := new(plumb.Message)

		basepaths := make(map[string]struct{})

		// This assumes that I will see plumb messages for all files. If plumber
		// crashes part way through, this tool is unlikely to recover successfully.
		for len(files) > len(basepaths) {
			err := m.Recv(brd)
			if err != nil {
				log.Fatalf("recv: %s", err)
			}
			// Consider making this into some kind of helper function.
			addr := m.LookupAttr("addr")
			path := string(m.Data)

			if addr == "" {
				if _, ok := files[path]; ok {
					basepaths[path] = struct{}{}
				}
			} else {
				if _, ok := files[path+":"+addr]; ok {
					basepaths[path] = struct{}{}
				}
			}
		}
		pathschan <- basepaths
	}()

	for k := range files {
		// TODO(rjk): It's possible to lift this into a helper function.
		sfid, err := plumb.Open("send", plan9.OWRITE)
		if err != nil {
			log.Fatalf("can't open plumber: %v", err)
		}
		defer sfid.Close()
		spm := new(plumb.Message)

		spm.Src = "editinacme"
		spm.Dst = "edit"
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("no current dir: %v", err)
		}
		spm.Dir = pwd
		spm.Type = "text"
		spm.Data = []byte(k)

		// I explicitly assumed that the plumber service must stay up.
		// Changing the edit destination will confuse the tool.
		if err := spm.Send(sfid); err != nil {
			log.Fatalf("can't send plumb message: %v", err)
		}
	}

	// Plumber has processed all of the messages and we know now their actual
	// paths.
	paths := <-pathschan

	// TODO(rjk): Loop here over all of the paths and set tags.

	if !*nowait {
		for len(paths) > 0 {
			ev, err := r.Read()
			if err != nil {
				log.Fatalf("reading acme log: %v", err)
			}
			if ev.Op == "del" {
				delete(paths, ev.Name)
			}
		}
	}
}
