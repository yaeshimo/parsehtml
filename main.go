package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const Name = "parsehtml"
const Version = "0.0.2"

// TODO: remove?
// change with ldflags?
var (
	// git rev-parse --verify --short HEAD
	Commit = ""
	// date -u +%Y-%m-%d
	Date = ""
)

// TODO: remove?
func printVersion() error {
	info := Name + " " + Version
	if Commit != "" {
		info += " (" + Commit
		if Date != "" {
			info += " " + Date
		}
		info += ")"
	}
	_, err := fmt.Println(info)
	return err
}

// example and comment for print usage
type Examples []struct {
	c string
	e string
}

func (es *Examples) Sprint() string {
	var s string
	for _, e := range *es {
		s += fmt.Sprintf("  %s\n", e.c)
		s += fmt.Sprintf("  $ %s\n\n", e.e)
	}
	return s
}

// append flag -example and hide from -help?
var examples = &Examples{
	{
		c: "Display help message",
		e: Name + " -help",
	},
	{
		c: "Output json format html nodes to stdout",
		e: Name + " -html /path/file.html",
	},
	{
		c: `Specify filter, same -json '{"attr":{"href":null}}'`,
		e: Name + " /path/file.html attr.href",
	},
	{
		c: "Filter with json",
		e: Name + ` -html file.html -json '{"type":"element"}'`,
	},
	{
		c: "The null means all match",
		e: Name + ` -html file.html -json '{"attr":{"href":null}}'`,
	},
	{
		c: "Output config template to stdout",
		e: Name + " -template",
	},
	{
		c: "Value filter with RE2",
		e: Name + ` /path/file.html 're2.attr.href=^https?://[\S]+$'`,
	},
}

func makeUsage(w *io.Writer) func() {
	return func() {
		flag.CommandLine.SetOutput(*w)
		// two spaces
		fmt.Fprintf(*w, "Description:\n")
		fmt.Fprintf(*w, "  Output json format html nodes\n\n")
		fmt.Fprintf(*w, "Usage:\n")
		fmt.Fprintf(*w, "  %s [Options]\n", Name)
		fmt.Fprintf(*w, "  %s [Options] File\n", Name)
		fmt.Fprintf(*w, "  %s [Options] File Filter...\n", Name)
		fmt.Fprintf(*w, "\n")
		fmt.Fprintf(*w, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(*w, "\n")
		fmt.Fprintf(*w, "Examples:\n%s", examples.Sprint())
	}
}

func template(w io.Writer) error {
	fil := NewFilter()
	b, err := fil.MarshalIndent()
	if err != nil {
		return nil
	}
	_, err = fmt.Fprintf(w, "%s\n", string(b))
	return err
}

var opt struct {
	help     bool
	version  bool
	template bool
	config   string
	html     string
	json     string
}

func init() {
	flag.BoolVar(&opt.help, "help", false, "Display help message")
	flag.BoolVar(&opt.version, "version", false, "Print version")
	flag.BoolVar(&opt.template, "template", false, "Output the filter template to stdout")
	flag.StringVar(&opt.html, "html", "", "Specify target html file")
	flag.StringVar(&opt.config, "config", "", "Specify JSON format config file for filter")
	flag.StringVar(&opt.json, "json", "", "Set filter")
}

func run() error {
	var usageWriter io.Writer = os.Stderr
	flag.Usage = makeUsage(&usageWriter)
	flag.Parse()

	var filterArgs []string
	if n := flag.NArg(); n != 0 {
		if opt.html == "" {
			opt.html = flag.Arg(0)
			filterArgs = flag.Args()[1:]
		} else {
			filterArgs = flag.Args()
		}
	}

	switch {
	case opt.help:
		usageWriter = os.Stdout
		flag.Usage()
		return nil
	case opt.version:
		return printVersion()
	case opt.template:
		return template(os.Stdout)
	}

	fil := NewFilter()
	if opt.config != "" {
		if err := fil.ReadConfig(opt.config); err != nil {
			return err
		}
	}
	if opt.json != "" {
		if err := fil.Unmarshal([]byte(opt.json)); err != nil {
			return err
		}
	}
	if len(filterArgs) != 0 {
		if err := fil.ParseArgs(filterArgs); err != nil {
			return err
		}
	}

	if err := fil.ParseFile(opt.html); err != nil {
		return err
	}
	b, err := fil.Nodes().MarshalIndent()
	if err != nil {
		return err
	}

	_, err = fmt.Printf("%s\n", string(b))
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
