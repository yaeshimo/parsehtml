package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFilter(t *testing.T) {
	var dir = filepath.Join("testdata", "filter")
	// file base
	const (
		expJSON  = "exp.json"
		config   = "config.json"
		fileHTML = "test.html"
	)

	var dfis []os.FileInfo
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range fis {
		if fi.IsDir() {
			dfis = append(dfis, fi)
		}
	}

	var (
		fi   os.FileInfo
		join = func(base string) string {
			return filepath.Join(dir, fi.Name(), base)
		}
	)
	for _, fi = range dfis {
		exp, err := ioutil.ReadFile(join(expJSON))
		if err != nil {
			t.Fatal(err)
		}

		ns := NewHTMLNodes()
		if err := ns.Filter.ReadConfig(join(config)); err != nil {
			t.Fatal(err)
		}

		// stored target nodes in ns.nodes
		if err := ns.ParseFile(join(fileHTML)); err != nil {
			t.Fatal(err)
		}

		b, err := ns.MarshalIndent()
		if err != nil {
			t.Fatal(err)
		}

		// TODO: fix?
		b = append(b, '\n')

		if !bytes.Equal(exp, b) {
			t.Fatalf("exp:%s but out:%s", exp, b)
		}
	}
}
