package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"gopkg.in/yaml.v2"
)

type (
	table struct {
		header []string
		rows   [][]string
	}
)

func main() {
	if err := cli(); err != nil {
		log.Fatalln(err)
	}
}

func cli() error {
	args, err := getArgs()
	if err != nil {
		return err
	}
	htm, err := getHTMLNode(args["url"])
	if err != nil {
		return err
	}
	tables := extractTables(htm)
	if args["ordinal"] == "" {
		for _, table := range tables {
			ts := ""
			switch args["format"] {
			case "yaml":
				{
					ts, err = table.yaml()
					if err != nil {
						return err
					}
					ts = "---\n" + ts
				}
			case "json":
				{
					ts, err = table.json()
					if err != nil {
						return err
					}
				}
			case "csv":
				{
					ts, err = table.csv(args["delimiter"])
					if err != nil {
						return err
					}
				}
			}
			fmt.Println(ts)
		}
	} else {
		n, err := strconv.Atoi(args["ordinal"])
		if err != nil {
			return err
		}
		if n > len(tables) {
			return fmt.Errorf("not enough tables on page")
		}
		ts := ""
		switch args["format"] {
		case "yaml":
			{
				ts, err = tables[n-1].yaml()
				if err != nil {
					return err
				}
				ts = "---\n" + ts
			}
		case "json":
			{
				ts, err = tables[n-1].json()
				if err != nil {
					return err
				}
			}
		case "csv":
			{
				ts, err = tables[n-1].csv(args["delimiter"])
				if err != nil {
					return err
				}
			}
		}
		fmt.Println(ts)
	}
	return nil
}

func getArgs() (map[string]string, error) {
	args := map[string]string{
		"format":  "",
		"url":     "",
		"ordinal": "",
	}
	dup := fmt.Errorf("cannot set more than one output format")
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-j", "-json":
			{
				if args["format"] != "" {
					return nil, dup
				}
				args["format"] = "json"
			}
		case "-y", "-yaml":
			{
				if args["format"] != "" {
					return nil, dup
				}
				args["format"] = "yaml"
			}
		}
		if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "http://") {
			if hasStdin() {
				return nil, fmt.Errorf("cant have url argument and process data on stdin")
			}
			u, err := url.Parse(arg)
			if err != nil {
				return nil, err
			}
			args["url"] = u.String()
		}
		if strings.HasPrefix(arg, "-n") {
			args["ordinal"] = strings.NewReplacer(
				"-n", "",
			).Replace(arg)
		}
		if strings.HasPrefix(arg, "-d") {
			if args["format"] != "" {
				return nil, dup
			}
			args["format"] = "csv"
			args["delimiter"] = strings.NewReplacer(
				"-d", "",
				`'`, "",
				`"`, "",
			).Replace(arg)
			if args["delimiter"] == "" {
				args["delimiter"] = ","
			}
		}
	}
	if args["format"] == "" {
		args["format"] = "csv"
		args["delimiter"] = ","
	}
	return args, nil
}

// func hasStdin() bool {
// 	si, err := os.Stdin.Stat()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return (si.Mode()&os.ModeNamedPipe) != 0 && si.Size() > 0
// }

func hasStdin() bool {
	si, _ := os.Stdin.Stat()
	return si.Mode() & os.ModeNamedPipe != 0
}

func grabTableHeader(tab *html.Node) []string {
	headerCells := grabAtoms(atom.Th, tab)
	cells := []string{}
	for _, cell := range headerCells {
		cells = append(cells, grabText(cell))
	}
	return cells
}

func grabTableBody(tab *html.Node) [][]string {
	rows := grabAtoms(atom.Tr, tab)
	if hasHeader(tab) {
		rows = rows[1:]
	}
	body := [][]string{}
	for _, row := range rows {
		r := []string{}
		cells := grabAtoms(atom.Td, row)
		for _, cell := range cells {
			r = append(r, grabText(cell))
		}
		body = append(body, r)
	}
	return body
}

func grabText(n *html.Node) string {
	b := bytes.Buffer{}
	fn := (func(*html.Node))(nil)
	fn = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(strings.TrimSpace(n.Data))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fn(c)
		}
	}
	fn(n)
	return b.String()
}

func getHTMLNode(url string) (doc *html.Node, err error) {
	// doc, err := &html.Node{}, (error)(nil)
	if url == "" {
		doc, err = html.Parse(os.Stdin)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		doc, err = html.Parse(resp.Body)
		if err != nil {
			return nil, err
		}
	}
	return doc, nil
}

func grabAtoms(targetType atom.Atom, doc *html.Node) []*html.Node {
	fn := (func(n *html.Node) []*html.Node)(nil)
	fn = func(n *html.Node) []*html.Node {
		nodes := []*html.Node{}
		if n.Type == html.ElementNode && n.DataAtom == targetType {
			nodes = append(nodes, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			nodes = append(nodes, fn(c)...)
		}
		return nodes
	}
	return fn(doc)
}

func (t *table) dump() {
	log.Printf("> \n%+v\n", t)
}

func (t *table) hasHeader() bool {
	return len(t.header) > 0
}

func hasHeader(tab *html.Node) bool {
	header := grabAtoms(atom.Th, tab)
	return len(header) > 0
}

func extractTable(tab *html.Node) *table {
	t := &table{}
	if hasHeader(tab) {
		t.header = grabTableHeader(tab)
	}
	t.rows = grabTableBody(tab)
	return t
}

func extractTables(htm *html.Node) []*table {
	tables := []*table{}
	tabs := grabAtoms(atom.Table, htm)
	for _, table := range tabs {
		tables = append(tables, extractTable(table))
	}
	return tables
}

func (t *table) json() (string, error) {
	// TODO: this has issues when processing non-standard tables
	b := bytes.Buffer{}
	j := []interface{}{}
	if t.hasHeader() {
		for _, row := range t.rows {
			rj := map[string]interface{}{}
			for i, key := range t.header {
				rj[key] = row[i]
			}
			j = append(j, rj)
		}
	} else {
		for _, row := range t.rows {
			j = append(j, row)
		}
	}
	err := json.NewEncoder(&b).Encode(j)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (t *table) yaml() (string, error) {
	j, err := t.json()
	if err != nil {
		return "", err
	}
	y := (interface{})(nil)
	err = yaml.Unmarshal([]byte(j), &y)
	if err != nil {
		return "", err
	}
	yy, err := yaml.Marshal(y)
	if err != nil {
		return "", err
	}
	return string(yy), nil
}

func (t *table) csv(delimiter string) (string, error) {
	csv := ""
	if t.hasHeader() {
		for i, hc := range t.header {
			csv = csv + hc
			if i != len(t.header)-1 {
				csv = csv + delimiter
			} else {
				csv = csv + "\n"
			}
		}
	}
	for _, row := range t.rows {
		for i, hc := range row {
			csv = csv + hc
			if i != len(row)-1 {
				csv = csv + delimiter
			} else {
				csv = csv + "\n"
			}
		}
	}
	return csv, nil
}
