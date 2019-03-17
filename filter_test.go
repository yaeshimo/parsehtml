package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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
		t.Logf("Test: %q", fi.Name())
		exp, err := ioutil.ReadFile(join(expJSON))
		if err != nil {
			t.Fatal(err)
		}

		fil := NewFilter()
		if err := fil.ReadConfig(join(config)); err != nil {
			t.Fatal(err)
		}

		// stored target nodes in ns.nodes
		if err := fil.ParseFile(join(fileHTML)); err != nil {
			t.Fatal(err)
		}

		b, err := fil.Nodes().MarshalIndent()
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

func TestParseArgs(t *testing.T) {
	var tests = []struct {
		args    []string
		expjson string
	}{
		{
			args:    []string{"type=element"},
			expjson: `{"type":"element"}`,
		},
		{
			args:    []string{"data=a"},
			expjson: `{"data":"a"}`,
		},
		{
			args:    []string{"attr.href"},
			expjson: `{"attr":{"href":null}}`,
		},

		// treat?
		//{
		//	args:    []string{`attr.href="http://example.com"`},
		//	expjson: `{"attr":{"href":"http://example.com"}}`,
		//},

		{
			args:    []string{`attr.href=http://example.com`},
			expjson: `{"attr":{"href":"http://example.com"}}`,
		},
		{
			args:    []string{`re2.data=^.*$`},
			expjson: `{"re2":{"data":"^.*$"}}`,
		},
		{
			args:    []string{`re2.attr.href=^https?://[\S]+$`},
			expjson: `{"re2":{"attr":{"href":"^https?://[\\S]+$"}}}`,
		},
	}

	sprint := func(fil *Filter) string {
		var str string
		str += fmt.Sprintf("type:%#v\n", fil.Type)
		str += fmt.Sprintf("data:%#v\n", fil.Data)
		str += fmt.Sprintf("attr:\n")
		for key, val := range fil.Attr {
			str += fmt.Sprintf("  key:%#v", key)
			if val != nil {
				str += fmt.Sprintf(" val:%#v", *val)
			}
			str += "\n"
		}
		str += fmt.Sprintf("re2:\n")
		str += fmt.Sprintf("  data:%#v\n", fil.Data)
		str += fmt.Sprintf("  attr:\n")
		for key, val := range fil.RE2.Attr {
			str += fmt.Sprintf("    key:%#v", key)
			if val != nil {
				str += fmt.Sprintf(" %#v", *val)
			}
			str += "\n"
		}
		str += fmt.Sprintf("  *match:\n")
		str += fmt.Sprintf("    data:%#v\n", fil.RE2.matchData)
		str += fmt.Sprintf("    attr:%#v\n", fil.RE2.matchAttr)
		return str
	}

	for _, test := range tests {
		expFil := NewFilter()
		if err := expFil.Unmarshal([]byte(test.expjson)); err != nil {
			t.Fatal(err)
		}
		fil := NewFilter()
		if err := fil.ParseArgs(test.args); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expFil, fil) {
			t.Errorf("fail case %q", test.args)
			t.Errorf("exp:\n%s", sprint(expFil))
			t.Errorf("out:\n%s", sprint(fil))
			t.FailNow()
		}
	}
}
