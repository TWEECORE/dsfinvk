package csvio

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/tweecore/dsfinvk/schema"
)

var (
	// ErrIndexXML reports an index.xml that does not describe a DSFinV-K export.
	ErrIndexXML = errors.New("malformed index.xml")
	// ErrFixedLength reports a fixed length table, which DSFinV-K does not use.
	ErrFixedLength = errors.New("fixed length tables are not supported")
	// ErrUnknownColumnType reports a column whose GDPdU type is neither AlphaNumeric nor Numeric.
	ErrUnknownColumnType = errors.New("unknown column type")
)

// gdpduVersion is the GDPdU description version index.xml declares.
const gdpduVersion = "1.0"

// dtdName is the file name of the DTD index.xml refers to.
const dtdName = "gdpdu-01-09-2004.dtd"

// indexXMLName is the file name of the description file of an export.
const indexXMLName = "index.xml"

// crlf ends every line of index.xml.
const crlf = "\r\n"

// DataSupplier is the <DataSupplier> block of index.xml.
type DataSupplier struct {
	Name     string
	Location string
	Comment  string
}

// WriteIndexXML writes the index.xml describing tables, byte-identical to the published one.
func WriteIndexXML(w io.Writer, tables []schema.Table, s DataSupplier, mediaName string) error {
	var b bytes.Buffer
	b.WriteString(bom)
	line(&b, 0, `<?xml version="1.0" encoding="utf-8"?>`)
	line(&b, 0, `<!DOCTYPE DataSet SYSTEM "`+dtdName+`">`)
	line(&b, 0, "<DataSet>")
	line(&b, 1, element("Version", gdpduVersion))
	line(&b, 1, "<DataSupplier>")
	line(&b, 2, element("Name", s.Name))
	line(&b, 2, element("Location", s.Location))
	line(&b, 2, element("Comment", s.Comment))
	line(&b, 1, "</DataSupplier>")
	line(&b, 1, "<Media>")
	line(&b, 2, element("Name", mediaName))
	for _, t := range tables {
		writeTable(&b, t)
	}
	line(&b, 1, "</Media>")
	b.WriteString("</DataSet>")

	if _, err := w.Write(b.Bytes()); err != nil {
		return fmt.Errorf("csvio: %s: %w", indexXMLName, err)
	}
	return nil
}

func writeTable(b *bytes.Buffer, t schema.Table) {
	line(b, 2, "<Table>")
	line(b, 3, element("URL", t.File))
	if t.Name != "" {
		line(b, 3, element("Name", t.Name))
	}
	if t.Description != "" {
		line(b, 3, element("Description", t.Description))
	}
	line(b, 3, "<UTF8 />")
	line(b, 3, element("DecimalSymbol", string(rune(schema.DecimalSymbol))))
	line(b, 3, element("DigitGroupingSymbol", string(rune(schema.DigitGroupingSymbol))))
	line(b, 3, "<Range>")
	line(b, 4, element("From", strconv.Itoa(schema.DataStartRow)))
	line(b, 3, "</Range>")
	line(b, 3, "<VariableLength>")
	line(b, 4, element("ColumnDelimiter", string(rune(schema.ColumnDelimiter))))
	line(b, 4, "<RecordDelimiter>&#xD;&#xA;</RecordDelimiter>")
	line(b, 4, element("TextEncapsulator", string(rune(schema.TextEncapsulator))))
	for _, c := range t.Columns {
		writeColumn(b, c)
	}
	line(b, 3, "</VariableLength>")
	line(b, 2, "</Table>")
}

func writeColumn(b *bytes.Buffer, c schema.Column) {
	line(b, 4, "<VariableColumn>")
	line(b, 5, element("Name", c.Name))
	if c.Description != "" {
		line(b, 5, element("Description", c.Description))
	}
	if c.Type == schema.ColumnNumeric {
		if c.Accuracy > 0 {
			line(b, 5, "<Numeric>")
			line(b, 6, element("Accuracy", strconv.Itoa(c.Accuracy)))
			line(b, 5, "</Numeric>")
		} else {
			line(b, 5, "<Numeric />")
		}
	} else {
		line(b, 5, "<AlphaNumeric />")
		if c.MaxLength > 0 {
			line(b, 5, element("MaxLength", strconv.Itoa(c.MaxLength)))
		}
	}
	line(b, 4, "</VariableColumn>")
}

// line writes s indented by depth levels of two spaces and terminated by CRLF.
func line(b *bytes.Buffer, depth int, s string) {
	for range depth {
		b.WriteString("  ")
	}
	b.WriteString(s)
	b.WriteString(crlf)
}

// element renders name with text, self-closing with a space when text is empty.
func element(name, text string) string {
	if text == "" {
		return "<" + name + " />"
	}
	return "<" + name + ">" + escape(text) + "</" + name + ">"
}

// escapeReplacer escapes exactly what the published index.xml escapes; quotes stay literal.
var escapeReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\r", "&#xD;")

func escape(s string) string { return escapeReplacer.Replace(s) }

type xmlEmpty struct{}

type xmlNumeric struct {
	Accuracy *int `xml:"Accuracy"`
}

type xmlColumn struct {
	Name         string      `xml:"Name"`
	Description  string      `xml:"Description"`
	AlphaNumeric *xmlEmpty   `xml:"AlphaNumeric"`
	Numeric      *xmlNumeric `xml:"Numeric"`
	MaxLength    int         `xml:"MaxLength"`
}

type xmlVariableLength struct {
	Columns []xmlColumn `xml:"VariableColumn"`
}

type xmlTable struct {
	URL            string             `xml:"URL"`
	Name           string             `xml:"Name"`
	Description    string             `xml:"Description"`
	VariableLength *xmlVariableLength `xml:"VariableLength"`
	FixedLength    *xmlEmpty          `xml:"FixedLength"`
}

type xmlMedia struct {
	Name   string     `xml:"Name"`
	Tables []xmlTable `xml:"Table"`
}

type xmlDataSet struct {
	XMLName  xml.Name      `xml:"DataSet"`
	Supplier *DataSupplier `xml:"DataSupplier"`
	Media    []xmlMedia    `xml:"Media"`
}

// ReadIndexXML parses an index.xml into the tables it describes.
func ReadIndexXML(r io.Reader) ([]schema.Table, DataSupplier, string, error) {
	br := bufio.NewReader(r)
	if prefix, err := br.Peek(len(bom)); err == nil && string(prefix) == bom {
		if _, err = br.Discard(len(bom)); err != nil {
			return nil, DataSupplier{}, "", fmt.Errorf("csvio: %s: %w", indexXMLName, err)
		}
	}

	var set xmlDataSet
	if err := xml.NewDecoder(br).Decode(&set); err != nil {
		return nil, DataSupplier{}, "", fmt.Errorf("csvio: %s: %w: %w", indexXMLName, ErrIndexXML, err)
	}
	if len(set.Media) != 1 {
		return nil, DataSupplier{}, "", fmt.Errorf("csvio: %s: %w: %d Media elements, want 1", indexXMLName, ErrIndexXML, len(set.Media))
	}

	tables := make([]schema.Table, 0, len(set.Media[0].Tables))
	for _, xt := range set.Media[0].Tables {
		t, err := convertTable(xt)
		if err != nil {
			return nil, DataSupplier{}, "", err
		}
		tables = append(tables, t)
	}

	var supplier DataSupplier
	if set.Supplier != nil {
		supplier = *set.Supplier
	}
	return tables, supplier, set.Media[0].Name, nil
}

func convertTable(xt xmlTable) (schema.Table, error) {
	if xt.FixedLength != nil {
		return schema.Table{}, fmt.Errorf("csvio: %s: table %s: %w", indexXMLName, xt.URL, ErrFixedLength)
	}
	if xt.VariableLength == nil {
		return schema.Table{}, fmt.Errorf("csvio: %s: table %s: %w: no VariableLength", indexXMLName, xt.URL, ErrIndexXML)
	}

	t := schema.Table{File: xt.URL, Name: xt.Name, Description: xt.Description}
	t.Columns = make([]schema.Column, 0, len(xt.VariableLength.Columns))
	for _, xc := range xt.VariableLength.Columns {
		c, err := convertColumn(xt.URL, xc)
		if err != nil {
			return schema.Table{}, err
		}
		t.Columns = append(t.Columns, c)
	}
	return t, nil
}

func convertColumn(file string, xc xmlColumn) (schema.Column, error) {
	c := schema.Column{Name: xc.Name, Description: xc.Description}
	switch {
	case xc.Numeric != nil:
		c.Type = schema.ColumnNumeric
		if xc.Numeric.Accuracy != nil {
			c.Accuracy = *xc.Numeric.Accuracy
		}
	case xc.AlphaNumeric != nil:
		c.Type = schema.ColumnAlphaNumeric
		c.MaxLength = xc.MaxLength
	default:
		return schema.Column{}, fmt.Errorf("csvio: %s: table %s: column %s: %w", indexXMLName, file, xc.Name, ErrUnknownColumnType)
	}
	return c, nil
}
