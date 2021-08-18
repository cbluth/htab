package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"encoding/json"

	// "gopkg.in/yaml.v3"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type (
	table struct {
		header []string
		rows [][]string
	}
)

func main() {
	if err := cli(); err != nil {
		log.Fatalln(err)
	}
}

func cli() error {
	// TODO: here, read stdin or from url arg
	htm, err := getHTMLNode()
	if err != nil {
		return err
	}
	tables := extractTables(htm)
	for _, table := range tables {
		jj, err := table.json()
		if err != nil {
			return err
		}
		fmt.Println(jj)
	}
	
	
	return nil
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

func getHTMLNode() (*html.Node, error) {
	si, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if (si.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("%s", "missing stdin")
	}
	doc, err := html.Parse(os.Stdin)
	if err != nil {
		return nil, err
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
