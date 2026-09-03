package csvio_test

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/schema"
)

// referenceSupplier is the DataSupplier the published index.xml carries.
var referenceSupplier = csvio.DataSupplier{Comment: "Datentraegerueberlassung nach GDPdU vom 12.11.2010"}

const referenceMedia = "CD Nummer 1"

func reportBytes(t *testing.T, got, want []byte) {
	t.Helper()

	n := min(len(got), len(want))
	off := n
	for i := range n {
		if got[i] != want[i] {
			off = i
			break
		}
	}

	t.Errorf("output differs at offset %d (got %d bytes, want %d)", off, len(got), len(want))
	t.Errorf("got:  %q", context(got, off))
	t.Errorf("want: %q", context(want, off))
}

// context returns up to 40 bytes of b starting at off.
func context(b []byte, off int) string {
	if off > len(b) {
		off = len(b)
	}
	return string(b[off:min(off+40, len(b))])
}

func TestWriteIndexXMLMatchesReference(t *testing.T) {
	var buf bytes.Buffer
	if err := csvio.WriteIndexXML(&buf, schema.Tables(), referenceSupplier, referenceMedia); err != nil {
		t.Fatalf("WriteIndexXML: %v", err)
	}
	if got, want := buf.Bytes(), schema.IndexXML(); !bytes.Equal(got, want) {
		reportBytes(t, got, want)
	}
}

func TestWriteIndexXMLWriteError(t *testing.T) {
	w := failingWriter{err: errors.New("boom")}
	if err := csvio.WriteIndexXML(w, schema.Tables(), referenceSupplier, referenceMedia); err == nil {
		t.Fatal("WriteIndexXML: want an error from the failing writer")
	}
}

func TestWriteIndexXMLOptionalElements(t *testing.T) {
	tables := []schema.Table{{
		File: "x.csv",
		Columns: []schema.Column{
			{Name: "A", Type: schema.ColumnAlphaNumeric},
			{Name: "B", Type: schema.ColumnNumeric, Accuracy: 3},
		},
	}}

	var buf bytes.Buffer
	if err := csvio.WriteIndexXML(&buf, tables, csvio.DataSupplier{Name: "N", Location: "L"}, "M"); err != nil {
		t.Fatalf("WriteIndexXML: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"<Name>N</Name>", "<Location>L</Location>", "<Comment />", "<Accuracy>3</Accuracy>"} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %q", want)
		}
	}
	for _, unwanted := range []string{"<MaxLength>", "<Description>"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("output should not contain %q", unwanted)
		}
	}
}

func TestWriteIndexXMLEscapes(t *testing.T) {
	tables := []schema.Table{{
		File:        "a&b.csv",
		Name:        "<t>",
		Description: "quotes 'x' \"y\" and \r",
		Columns:     []schema.Column{{Name: "A", Type: schema.ColumnNumeric}},
	}}

	var buf bytes.Buffer
	if err := csvio.WriteIndexXML(&buf, tables, csvio.DataSupplier{}, "M"); err != nil {
		t.Fatalf("WriteIndexXML: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"<URL>a&amp;b.csv</URL>", "<Name>&lt;t&gt;</Name>", "quotes 'x' \"y\" and &#xD;"} {
		if !strings.Contains(out, want) {
			t.Errorf("output misses %q\ngot: %s", want, out)
		}
	}
}

func TestReadIndexXMLReference(t *testing.T) {
	tables, supplier, media, err := csvio.ReadIndexXML(bytes.NewReader(schema.IndexXML()))
	if err != nil {
		t.Fatalf("ReadIndexXML: %v", err)
	}
	if !reflect.DeepEqual(tables, schema.Tables()) {
		t.Errorf("tables differ from schema.Tables()")
	}
	if supplier != referenceSupplier {
		t.Errorf("supplier = %+v, want %+v", supplier, referenceSupplier)
	}
	if media != referenceMedia {
		t.Errorf("media = %q, want %q", media, referenceMedia)
	}
}

// handwritten is an index.xml whose single table carries an extra column.
const handwritten = "\ufeff" + `<?xml version="1.0" encoding="utf-8"?>` + "\r\n" +
	`<!DOCTYPE DataSet SYSTEM "gdpdu-01-09-2004.dtd">` + "\r\n" +
	"<DataSet>\r\n" +
	"  <Version>1.0</Version>\r\n" +
	"  <DataSupplier>\r\n" +
	"    <Name>Supplier</Name>\r\n" +
	"    <Location />\r\n" +
	"    <Comment>Test</Comment>\r\n" +
	"  </DataSupplier>\r\n" +
	"  <Media>\r\n" +
	"    <Name>CD Nummer 1</Name>\r\n" +
	"    <Table>\r\n" +
	"      <URL>tse.csv</URL>\r\n" +
	"      <Name>TSE</Name>\r\n" +
	"      <Description>tse.csv</Description>\r\n" +
	"      <UTF8 />\r\n" +
	"      <DecimalSymbol>,</DecimalSymbol>\r\n" +
	"      <DigitGroupingSymbol>.</DigitGroupingSymbol>\r\n" +
	"      <Range>\r\n" +
	"        <From>2</From>\r\n" +
	"      </Range>\r\n" +
	"      <VariableLength>\r\n" +
	"        <ColumnDelimiter>;</ColumnDelimiter>\r\n" +
	"        <RecordDelimiter>&#xD;&#xA;</RecordDelimiter>\r\n" +
	`        <TextEncapsulator>"</TextEncapsulator>` + "\r\n" +
	"        <VariableColumn>\r\n" +
	"          <Name>Z_KASSE_ID</Name>\r\n" +
	"          <Description>ID der (Abschluss-) Kasse</Description>\r\n" +
	"          <AlphaNumeric />\r\n" +
	"          <MaxLength>50</MaxLength>\r\n" +
	"        </VariableColumn>\r\n" +
	"        <VariableColumn>\r\n" +
	"          <Name>Z_NR</Name>\r\n" +
	"          <Description>Nr. des Kassenabschlusses</Description>\r\n" +
	"          <Numeric />\r\n" +
	"        </VariableColumn>\r\n" +
	"        <VariableColumn>\r\n" +
	"          <Name>TSE_ZERTIFIKAT_III</Name>\r\n" +
	"          <Description>Drittes Teilzertifikat</Description>\r\n" +
	"          <AlphaNumeric />\r\n" +
	"          <MaxLength>1000</MaxLength>\r\n" +
	"        </VariableColumn>\r\n" +
	"      </VariableLength>\r\n" +
	"    </Table>\r\n" +
	"  </Media>\r\n" +
	"</DataSet>"

func TestReadIndexXMLRoundTrip(t *testing.T) {
	tables, supplier, media, err := csvio.ReadIndexXML(strings.NewReader(handwritten))
	if err != nil {
		t.Fatalf("ReadIndexXML: %v", err)
	}
	if len(tables) != 1 || len(tables[0].Columns) != 3 {
		t.Fatalf("got %d tables, first with %d columns", len(tables), len(tables[0].Columns))
	}
	extra := tables[0].Columns[2]
	if extra.Name != "TSE_ZERTIFIKAT_III" || extra.Type != schema.ColumnAlphaNumeric || extra.MaxLength != 1000 {
		t.Errorf("extra column = %+v", extra)
	}

	var buf bytes.Buffer
	if err := csvio.WriteIndexXML(&buf, tables, supplier, media); err != nil {
		t.Fatalf("WriteIndexXML: %v", err)
	}
	if got, want := buf.Bytes(), []byte(handwritten); !bytes.Equal(got, want) {
		reportBytes(t, got, want)
	}
}

func TestReadIndexXMLErrors(t *testing.T) {
	fixed := strings.Replace(handwritten, "<VariableLength>", "<FixedLength>", 1)
	fixed = strings.Replace(fixed, "</VariableLength>", "</FixedLength>", 1)

	noType := strings.Replace(handwritten, "          <Numeric />\r\n", "          <Date />\r\n", 1)
	twoMedia := strings.Replace(handwritten, "  </Media>\r\n", "  </Media>\r\n  <Media>\r\n    <Name>2</Name>\r\n  </Media>\r\n", 1)

	noBody := strings.Replace(handwritten, "<VariableLength>", "<Nope>", 1)
	noBody = strings.Replace(noBody, "</VariableLength>", "</Nope>", 1)

	tests := map[string]struct {
		in   string
		want error
	}{
		"not xml":       {"nonsense", csvio.ErrIndexXML},
		"wrong root":    {`<Other></Other>`, csvio.ErrIndexXML},
		"fixed length":  {fixed, csvio.ErrFixedLength},
		"unknown type":  {noType, csvio.ErrUnknownColumnType},
		"two media":     {twoMedia, csvio.ErrIndexXML},
		"no media":      {"<DataSet>\r\n  <Version>1.0</Version>\r\n</DataSet>", csvio.ErrIndexXML},
		"no table body": {noBody, csvio.ErrIndexXML},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, _, err := csvio.ReadIndexXML(strings.NewReader(tc.in))
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestReadIndexXMLReadError(t *testing.T) {
	if _, _, _, err := csvio.ReadIndexXML(failingReader{}); err == nil {
		t.Fatal("want an error from the failing reader")
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
