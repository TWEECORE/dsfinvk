// Command gen turns the GDPdU index.xml shipped with DSFinV-K into Go table
// definitions in package schema.
package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")

	in := flag.String("in", "schema/index.xml", "path to index.xml")
	out := flag.String("out", "schema/tables_gen.go", "path of the Go file to write")
	pkg := flag.String("pkg", "schema", "package name of the generated file")
	flag.Parse()

	if err := run(*in, *out, *pkg); err != nil {
		log.Fatal(err)
	}
}

func run(in, out, pkg string) error {
	raw, err := os.ReadFile(in) //nolint:gosec // build-time path from the -in flag
	if err != nil {
		return err
	}
	ds, err := parse(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}
	f, err := ds.format()
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}
	src, err := render(pkg, f, ds.tables())
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}
	return os.WriteFile(out, src, 0o644) //nolint:gosec // a source file must be world-readable
}

// bom is the UTF-8 byte order mark, which encoding/xml refuses to parse.
var bom = []byte{0xEF, 0xBB, 0xBF}

type dataSet struct {
	XMLName xml.Name `xml:"DataSet"`
	Media   []media  `xml:"Media"`
}

type media struct {
	Tables []xmlTable `xml:"Table"`
}

type xmlTable struct {
	URL                 string    `xml:"URL"`
	Name                string    `xml:"Name"`
	Description         string    `xml:"Description"`
	UTF8                *struct{} `xml:"UTF8"`
	DecimalSymbol       string    `xml:"DecimalSymbol"`
	DigitGroupingSymbol string    `xml:"DigitGroupingSymbol"`
	Range               struct {
		From string `xml:"From"`
	} `xml:"Range"`
	VariableLength varLength `xml:"VariableLength"`
}

func (t xmlTable) columns() []xmlColumn { return t.VariableLength.Columns }

type varLength struct {
	ColumnDelimiter  string      `xml:"ColumnDelimiter"`
	RecordDelimiter  string      `xml:"RecordDelimiter"`
	TextEncapsulator string      `xml:"TextEncapsulator"`
	Columns          []xmlColumn `xml:"VariableColumn"`
}

type xmlColumn struct {
	Name         string      `xml:"Name"`
	Description  string      `xml:"Description"`
	AlphaNumeric *struct{}   `xml:"AlphaNumeric"`
	MaxLength    *int        `xml:"MaxLength"`
	Numeric      *xmlNumeric `xml:"Numeric"`
	Date         *struct{}   `xml:"Date"`
}

type xmlNumeric struct {
	// Accuracy is the fraction digit count; the GDPdU DTD defaults it to 0.
	Accuracy *int `xml:"Accuracy"`
}

func parse(raw []byte) (*dataSet, error) {
	var ds dataSet
	if err := xml.Unmarshal(bytes.TrimPrefix(raw, bom), &ds); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	tables := ds.tables()
	if len(tables) == 0 {
		return nil, fmt.Errorf("no <Table> elements found")
	}
	seen := make(map[string]bool, len(tables))
	for _, t := range tables {
		if t.URL == "" {
			return nil, fmt.Errorf("table with empty <URL>")
		}
		if seen[t.URL] {
			return nil, fmt.Errorf("duplicate table %q", t.URL)
		}
		seen[t.URL] = true
		if len(t.columns()) == 0 {
			return nil, fmt.Errorf("table %q has no columns", t.URL)
		}
		cols := make(map[string]bool, len(t.columns()))
		for _, c := range t.columns() {
			if c.Name == "" {
				return nil, fmt.Errorf("table %q: column with empty <Name>", t.URL)
			}
			if cols[c.Name] {
				return nil, fmt.Errorf("table %q: duplicate column %q", t.URL, c.Name)
			}
			cols[c.Name] = true
			switch {
			case c.Date != nil:
				return nil, fmt.Errorf("table %q column %q: <Date> columns are not supported", t.URL, c.Name)
			case c.AlphaNumeric != nil && c.Numeric != nil:
				return nil, fmt.Errorf("table %q column %q: both <AlphaNumeric> and <Numeric>", t.URL, c.Name)
			case c.AlphaNumeric != nil:
				if c.MaxLength == nil {
					return nil, fmt.Errorf("table %q column %q: <AlphaNumeric> without <MaxLength>", t.URL, c.Name)
				}
				if *c.MaxLength <= 0 {
					return nil, fmt.Errorf("table %q column %q: MaxLength %d", t.URL, c.Name, *c.MaxLength)
				}
			case c.Numeric != nil:
				if c.MaxLength != nil {
					return nil, fmt.Errorf("table %q column %q: <Numeric> with <MaxLength>", t.URL, c.Name)
				}
				if c.Numeric.Accuracy != nil && *c.Numeric.Accuracy < 0 {
					return nil, fmt.Errorf("table %q column %q: Accuracy %d", t.URL, c.Name, *c.Numeric.Accuracy)
				}
			default:
				return nil, fmt.Errorf("table %q column %q: neither <AlphaNumeric> nor <Numeric>", t.URL, c.Name)
			}
		}
	}
	return &ds, nil
}

func (ds *dataSet) tables() []xmlTable {
	var out []xmlTable
	for _, m := range ds.Media {
		out = append(out, m.Tables...)
	}
	return out
}

// formatSettings are the CSV dialect settings index.xml repeats on every table.
type formatSettings struct {
	ColumnDelimiter     rune
	TextEncapsulator    rune
	RecordDelimiter     string
	DecimalSymbol       rune
	DigitGroupingSymbol rune
	Encoding            string
	HeaderRow           int
	DataStartRow        int
}

// format extracts the shared dialect, failing when any table disagrees.
func (ds *dataSet) format() (formatSettings, error) {
	var f formatSettings
	tables := ds.tables()
	if len(tables) == 0 {
		return f, fmt.Errorf("no tables")
	}

	first := tables[0]
	dataStartRow, err := strconv.Atoi(strings.TrimSpace(first.Range.From))
	if err != nil {
		return f, fmt.Errorf("table %q: Range/From %q: %w", first.URL, first.Range.From, err)
	}

	str := map[string]string{
		"DecimalSymbol":       first.DecimalSymbol,
		"DigitGroupingSymbol": first.DigitGroupingSymbol,
		"ColumnDelimiter":     first.VariableLength.ColumnDelimiter,
		"TextEncapsulator":    first.VariableLength.TextEncapsulator,
		"RecordDelimiter":     first.VariableLength.RecordDelimiter,
	}

	for _, t := range tables {
		if t.UTF8 == nil {
			return f, fmt.Errorf("table %q: missing <UTF8/>; DSFinV-K exports are UTF-8 only", t.URL)
		}
		got := map[string]string{
			"DecimalSymbol":       t.DecimalSymbol,
			"DigitGroupingSymbol": t.DigitGroupingSymbol,
			"ColumnDelimiter":     t.VariableLength.ColumnDelimiter,
			"TextEncapsulator":    t.VariableLength.TextEncapsulator,
			"RecordDelimiter":     t.VariableLength.RecordDelimiter,
		}
		for _, k := range []string{"DecimalSymbol", "DigitGroupingSymbol", "ColumnDelimiter", "TextEncapsulator", "RecordDelimiter"} {
			if got[k] != str[k] {
				return f, fmt.Errorf("table %q: %s is %q but table %q uses %q; the format constants assume one dialect for the whole export",
					t.URL, k, got[k], first.URL, str[k])
			}
		}
		n, err := strconv.Atoi(strings.TrimSpace(t.Range.From))
		if err != nil {
			return f, fmt.Errorf("table %q: Range/From %q: %w", t.URL, t.Range.From, err)
		}
		if n != dataStartRow {
			return f, fmt.Errorf("table %q: RangeFrom is %d but table %q uses %d", t.URL, n, first.URL, dataStartRow)
		}
	}

	for _, k := range []string{"DecimalSymbol", "DigitGroupingSymbol", "ColumnDelimiter", "TextEncapsulator"} {
		if len([]rune(str[k])) != 1 {
			return f, fmt.Errorf("%s is %q; expected exactly one character", k, str[k])
		}
	}
	if str["RecordDelimiter"] == "" {
		return f, fmt.Errorf("RecordDelimiter is empty")
	}

	firstRune := func(s string) rune {
		r, _ := utf8.DecodeRuneInString(s)
		return r
	}
	if dataStartRow < 2 {
		return f, fmt.Errorf("Range/From is %d; the first row must be the header row", dataStartRow)
	}

	f = formatSettings{
		ColumnDelimiter:     firstRune(str["ColumnDelimiter"]),
		TextEncapsulator:    firstRune(str["TextEncapsulator"]),
		RecordDelimiter:     str["RecordDelimiter"],
		DecimalSymbol:       firstRune(str["DecimalSymbol"]),
		DigitGroupingSymbol: firstRune(str["DigitGroupingSymbol"]),
		// Every table carries <UTF8/>, which the loop above enforces.
		Encoding:     "UTF-8",
		HeaderRow:    dataStartRow - 1,
		DataStartRow: dataStartRow,
	}
	return f, nil
}

const header = "// Code generated by internal/gen from index.xml; DO NOT EDIT.\n"

func render(pkg string, f formatSettings, tables []xmlTable) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(header)
	b.WriteString("\n")
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	b.WriteString("import \"strconv\"\n\n")

	b.WriteString(`// ColumnType is the GDPdU data type of a column, as declared in index.xml.
type ColumnType uint8

const (
	ColumnAlphaNumeric ColumnType = iota
	ColumnNumeric
)

// String returns the GDPdU element name of the type.
func (t ColumnType) String() string {
	switch t {
	case ColumnAlphaNumeric:
		return "AlphaNumeric"
	case ColumnNumeric:
		return "Numeric"
	default:
		return "ColumnType(" + strconv.Itoa(int(t)) + ")"
	}
}

// Column is one <VariableColumn> of a DSFinV-K table.
type Column struct {
	Name        string
	Description string
	Type        ColumnType
	MaxLength   int
	Accuracy    int
}

// Table is one <Table> of a DSFinV-K export.
type Table struct {
	File        string
	Name        string
	Description string
	Columns     []Column
}

`)

	b.WriteString("// CSV dialect of every DSFinV-K table, read from index.xml.\n")
	b.WriteString("const (\n")
	fmt.Fprintf(&b, "\tColumnDelimiter = %s\n", strconv.QuoteRune(f.ColumnDelimiter))
	fmt.Fprintf(&b, "\tTextEncapsulator = %s\n", strconv.QuoteRune(f.TextEncapsulator))
	fmt.Fprintf(&b, "\tRecordDelimiter = %s\n", strconv.Quote(f.RecordDelimiter))
	fmt.Fprintf(&b, "\tDecimalSymbol = %s\n", strconv.QuoteRune(f.DecimalSymbol))
	b.WriteString("\t// DigitGroupingSymbol is what index.xml declares; grouping must not be emitted.\n")
	fmt.Fprintf(&b, "\tDigitGroupingSymbol = %s\n", strconv.QuoteRune(f.DigitGroupingSymbol))
	fmt.Fprintf(&b, "\tEncoding = %s\n", strconv.Quote(f.Encoding))
	fmt.Fprintf(&b, "\tHeaderRow = %d\n", f.HeaderRow)
	fmt.Fprintf(&b, "\tDataStartRow = %d\n", f.DataStartRow)
	b.WriteString(")\n\n")

	b.WriteString("// tables are the DSFinV-K tables in index.xml order; Tables clones them.\n")
	b.WriteString("var tables = []Table{\n")
	for _, t := range tables {
		b.WriteString("\t{\n")
		fmt.Fprintf(&b, "\t\tFile:        %s,\n", strconv.Quote(t.URL))
		fmt.Fprintf(&b, "\t\tName:        %s,\n", strconv.Quote(t.Name))
		fmt.Fprintf(&b, "\t\tDescription: %s,\n", strconv.Quote(t.Description))
		b.WriteString("\t\tColumns: []Column{\n")
		for _, c := range t.columns() {
			fmt.Fprintf(&b, "\t\t\t{Name: %s, Description: %s, Type: %s",
				strconv.Quote(c.Name), strconv.Quote(c.Description), goColumnType(c))
			if c.AlphaNumeric != nil {
				fmt.Fprintf(&b, ", MaxLength: %d", *c.MaxLength)
			} else if c.Numeric.Accuracy != nil && *c.Numeric.Accuracy != 0 {
				fmt.Fprintf(&b, ", Accuracy: %d", *c.Numeric.Accuracy)
			}
			b.WriteString("},\n")
		}
		b.WriteString("\t\t},\n")
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")

	src, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("gofmt the generated source: %w", err)
	}
	return src, nil
}

func goColumnType(c xmlColumn) string {
	if c.AlphaNumeric != nil {
		return "ColumnAlphaNumeric"
	}
	return "ColumnNumeric"
}
