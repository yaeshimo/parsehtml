package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"golang.org/x/net/html"
)

// TODO:
// 1. make test?
// 2. impl flags for more useful
// 3. split to other source files?

// provide convert functions?
var nodeTypeToString = map[html.NodeType]string{
	html.ErrorNode:    "error",
	html.TextNode:     "text",
	html.DocumentNode: "document",
	html.ElementNode:  "element",
	html.CommentNode:  "comment",
	html.DoctypeNode:  "doctype",
}
var stringToNodeType = map[string]html.NodeType{
	"error":    html.ErrorNode,
	"text":     html.TextNode,
	"document": html.DocumentNode,
	"element":  html.ElementNode,
	"comment":  html.CommentNode,
	"doctype":  html.DoctypeNode,
}

// regular expression 2
type RE2 struct {
	// pattern
	PatData *string            `json:"pat_data"`
	PatAttr map[string]*string `json:"pat_attr"`

	// store compiled regexp
	matchData *regexp.Regexp
	matchAttr map[string]*regexp.Regexp
}

func NewRE2() *RE2 {
	return &RE2{matchAttr: make(map[string]*regexp.Regexp)}
}

func (p *RE2) Compile() error {
	var err error
	if p.PatData != nil {
		p.matchData, err = regexp.Compile(*p.PatData)
		if err != nil {
			return err
		}
	}
	p.matchAttr = make(map[string]*regexp.Regexp)
	for key, pat := range p.PatAttr {
		if pat == nil {
			p.matchAttr[key] = nil
			continue
		}
		r, err := regexp.Compile(*pat)
		if err != nil {
			return err
		}
		p.matchAttr[key] = r
	}
	return nil
}

func (p *RE2) MatchData(s string) bool {
	if p.matchData == nil {
		return true
	}
	return p.matchData.MatchString(s)
}

func (p *RE2) MatchAttr(attrs []html.Attribute) bool {
	nmap := len(p.PatAttr)
	if nmap != 0 {
		for _, attr := range attrs {
			match, ok := p.matchAttr[attr.Key]
			if ok && (match == nil || match.MatchString(attr.Val)) {
				nmap--
			}
		}
	}
	return nmap == 0
}

// filter for html.Node
// if value is nil then *Filter.IsWant return true
type Filter struct {
	// filter by html.NodeType
	Type *string `json:"type"`

	// filter by html.Data
	Data *string `json:"data"`

	// filter by html.Attribute
	Attr map[string]*string `json:"attr"`

	// regexp filter
	// use RE2 for attribute valus and node data
	RE2 *RE2 `json:"re2"`
}

func NewFilter() *Filter {
	return &Filter{
		Attr: make(map[string]*string),
		RE2:  NewRE2(),
	}
}

func (fil *Filter) readConfig(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &fil)
}

func (fil *Filter) ReadConfig(file string) error {
	if err := fil.readConfig(file); err != nil {
		return err
	}
	return fil.RE2.Compile()
}

// TODO: rename?
func (fil *Filter) IsWant(n *html.Node) bool {
	if n == nil {
		return false
	}

	// TODO: to method?
	if fil.Type != nil {
		expType, ok := stringToNodeType[*fil.Type]
		if !ok {
			return false
		}
		if n.Type != expType {
			return false
		}
	}

	// TODO: to method?
	if fil.Data != nil {
		if *fil.Data != n.Data {
			return false
		}
	}

	// TODO: to method?
	// needs all matched
	if nmap := len(fil.Attr); nmap != 0 {
		for _, attr := range n.Attr {
			val, ok := fil.Attr[attr.Key]
			if ok && (val == nil || *val == attr.Val) {
				nmap--
			}
		}
		if nmap != 0 {
			return false
		}
	}

	// RE2
	// need compile before use
	if !fil.RE2.MatchData(n.Data) {
		return false
	}
	if !fil.RE2.MatchAttr(n.Attr) {
		return false
	}

	//return fil.RE2.MatchAttr(n.Attr) && fil.RE2.MatchData(n.Data)

	return true
}

type HTMLNode struct {
	Type string            `json:"type"`
	Data string            `json:"data"`
	Attr map[string]string `json:"attr"`
}
type HTMLNodes struct {
	Filter *Filter

	// store parsed nodes
	nodes []*HTMLNode
}

func NewHTMLNodes() *HTMLNodes {
	return &HTMLNodes{Filter: NewFilter()}
}

func (p *HTMLNodes) MarshalIndent() ([]byte, error) {
	return json.MarshalIndent(p.nodes, "", "\t")
}

func (p *HTMLNodes) Add(n *html.Node) {
	if n == nil {
		return
	}
	ts, ok := nodeTypeToString[n.Type]
	if !ok {
		// TODO: handle?
		panic("fail *HTMLNodes.Add: can not convert html.NodeType to string")
	}
	nn := &HTMLNode{
		Type: ts,
		Data: n.Data,
		Attr: make(map[string]string),
	}
	for _, attr := range n.Attr {
		nn.Attr[attr.Key] = attr.Val
	}
	p.nodes = append(p.nodes, nn)
}

// parse html
func (p *HTMLNodes) ParseFile(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	n, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n == nil {
			return
		}
		if p.Filter.IsWant(n) {
			p.Add(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return nil
}

func run() error {
	file := "testdata/ex/test.html"
	config := "testdata/ex/config.json"

	ns := NewHTMLNodes()
	if err := ns.Filter.ReadConfig(config); err != nil {
		return err
	}

	if err := ns.ParseFile(file); err != nil {
		return err
	}

	b, err := ns.MarshalIndent()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", b)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
