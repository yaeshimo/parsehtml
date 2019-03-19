package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"regexp"

	"golang.org/x/net/html"
)

// provide convert functions or use regexp?
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

type HTMLNode struct {
	Type string            `json:"type"`
	Data string            `json:"data"`
	Attr map[string]string `json:"attr"`
}
type HTMLNodes []*HTMLNode

func (ns *HTMLNodes) MarshalIndent() ([]byte, error) {
	if len(*ns) == 0 {
		return nil, errors.New("not contain html nodes")
	}
	return json.MarshalIndent(ns, "", "\t")
}

func (p *HTMLNodes) Add(n *html.Node) {
	if n == nil {
		return
	}
	ts, ok := nodeTypeToString[n.Type]
	if !ok {
		// TODO: is need handle? maybe not
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
	*p = append(*p, nn)
}

// TODO: RE2 join to Filter?
// regular expression 2
type RE2 struct {
	// pattern
	Data *string            `json:"data"`
	Attr map[string]*string `json:"attr"`

	// store compiled regexp
	matchData *regexp.Regexp
	matchAttr map[string]*regexp.Regexp
}

func NewRE2() *RE2 {
	return &RE2{
		Attr:      make(map[string]*string),
		matchAttr: make(map[string]*regexp.Regexp),
	}
}

// need compile before use RE2.Match*
func (re2 *RE2) Compile() error {
	var err error
	if re2.Data != nil {
		re2.matchData, err = regexp.Compile(*re2.Data)
		if err != nil {
			return err
		}
	}
	re2.matchAttr = make(map[string]*regexp.Regexp)
	for key, pat := range re2.Attr {
		if pat == nil {
			re2.matchAttr[key] = nil
			continue
		}
		r, err := regexp.Compile(*pat)
		if err != nil {
			return err
		}
		re2.matchAttr[key] = r
	}
	return nil
}

// need compile before use
func (re2 *RE2) MatchData(s string) bool {
	if re2.matchData == nil {
		return true
	}
	return re2.matchData.MatchString(s)
}

// need compile before use
func (re2 *RE2) MatchAttr(attrs []html.Attribute) bool {
	nmap := len(re2.Attr)
	if nmap != 0 {
		for _, attr := range attrs {
			match, ok := re2.matchAttr[attr.Key]
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
	// filter for html.NodeType
	Type *string `json:"type"`

	// filter for html.Data
	Data *string `json:"data"`

	// filter for html.Attribute
	Attr map[string]*string `json:"attr"`

	// regexp filter for data and attribute
	RE2 *RE2 `json:"re2"`

	// store filtered nodes
	nodes *HTMLNodes
}

func NewFilter() *Filter {
	return &Filter{
		Attr:  make(map[string]*string),
		RE2:   NewRE2(),
		nodes: new(HTMLNodes),
	}
}

func (fil *Filter) Unmarshal(b []byte) error {
	err := json.Unmarshal(b, &fil)
	if err != nil {
		return err
	}
	return fil.RE2.Compile()
}

func (fil *Filter) ReadConfig(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return fil.Unmarshal(b)
}

// TODO: fix to simple
// set filter by filter args
// eg. []string{`type=element` `attr.class`}
// type=element is same as {"type":"element"}
// attr.class is same as {"attr":{"class":null}}
func (fil *Filter) ParseArgs(args []string) error {
	splitLine := func(line string) []string {
		var args []string
		var (
			r       rune
			rs      []rune
			literal bool // is need?
			nline   = len(line)
		)
	loop:
		for i := 0; i < nline; i++ {
			r = rune(line[i])
			// escape
			if r == '\\' {
				i++
				if i > nline {
					// treat last one?
					//rs = append(rs, r)
					break loop
				}
				rs = append(rs, rune(line[i]))
				continue
			}

			// in literal
			if literal {
				if r == '"' {
					literal = false
					continue
				}
				rs = append(rs, r)
				continue
			}

			switch r {
			case '.':
				// split
				args = append(args, string(rs))
				rs = []rune{}
				continue
			case '"':
				// to literal
				literal = true
				continue
			case '=':
				// for value append all runes
				args = append(args, string(rs), line[i+1:])
				rs = []rune{} // not graceful
				break loop
			}
			rs = append(rs, r)
		}
		if len(rs) != 0 {
			args = append(args, string(rs))
		}
		return args
	}

	for _, arg := range args {
		filters := splitLine(arg)
		nf := len(filters)
		if nf < 1 {
			continue
		}
		switch filters[0] {
		case "type":
			if nf == 2 {
				fil.Type = &filters[1]
				continue
			}
		case "data":
			if nf == 2 {
				fil.Data = &filters[1]
				continue
			}
		case "attr":
			if nf == 2 {
				fil.Attr[filters[1]] = nil
				continue
			} else if nf == 3 {
				fil.Attr[filters[1]] = &filters[2]
				continue
			}
		case "re2":
			if nf > 2 {
				switch filters[1] {
				case "data":
					if nf == 3 {
						fil.RE2.Data = &filters[2]
						continue
					}
				case "attr":
					if nf == 3 {
						fil.RE2.Attr[filters[2]] = nil
						continue
					} else if nf == 4 {
						fil.RE2.Attr[filters[2]] = &filters[3]
						continue
					}
				}
			}
		}
		return errors.New("filter parse error: " + arg)
	}

	return fil.RE2.Compile()
}

func (fil *Filter) MarshalIndent() ([]byte, error) {
	return json.MarshalIndent(fil, "", "\t")
}

// rename?
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
	if !fil.RE2.MatchData(n.Data) {
		return false
	}
	if !fil.RE2.MatchAttr(n.Attr) {
		return false
	}

	return true
}

// parse html
func (fil *Filter) ParseFile(file string) error {
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
		if fil.IsWant(n) {
			fil.nodes.Add(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return nil
}

// provide filtered *HTMLNodes
func (fil *Filter) Nodes() *HTMLNodes {
	return fil.nodes
}
