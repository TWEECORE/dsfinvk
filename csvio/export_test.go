package csvio_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/schema"
)

const dtdFile = "gdpdu-01-09-2004.dtd"

// wildRow is one record of values that stress the CSV dialect.
var wildRow = []string{"Kasse Öl", "a;b", `say "hi"`, "line1\nline2", "quote “x”", "", "1"}

func mustClose(t *testing.T, c io.Closer) {
	t.Helper()

	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// exportEmpty writes an export with no rows into sink and returns nothing but a failed test on error.
func exportEmpty(t *testing.T, sink csvio.Sink, tables []schema.Table) {
	t.Helper()

	e, err := csvio.NewExportWriter(sink, tables, referenceSupplier)
	if err != nil {
		t.Fatalf("NewExportWriter: %v", err)
	}
	mustClose(t, e)
}

func readFile(t *testing.T, dir, name string) []byte {
	t.Helper()

	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", name, err)
	}
	return b
}

func headerLine(t schema.Table) string {
	names := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		names[i] = c.Name
	}
	return strings.Join(names, ";") + "\r\n"
}

func TestExportEmptyToDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "export")
	sink, err := csvio.NewDirSink(dir, false)
	if err != nil {
		t.Fatalf("NewDirSink: %v", err)
	}
	exportEmpty(t, sink, schema.Tables())

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 22 {
		t.Fatalf("got %d files, want 22", len(entries))
	}

	for _, tbl := range schema.Tables() {
		if got, want := string(readFile(t, dir, tbl.File)), headerLine(tbl); got != want {
			t.Errorf("%s = %q, want %q", tbl.File, got, want)
		}
	}
	if got := readFile(t, dir, "index.xml"); !bytes.Equal(got, schema.IndexXML()) {
		reportBytes(t, got, schema.IndexXML())
	}
	if got := readFile(t, dir, dtdFile); !bytes.Equal(got, schema.DTD()) {
		t.Errorf("%s differs from schema.DTD()", dtdFile)
	}
}

func TestExportEmptyToZip(t *testing.T) {
	var buf bytes.Buffer
	exportEmpty(t, csvio.NewZipSink(&buf), schema.Tables())

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}

	names := make([]string, len(zr.File))
	for i, f := range zr.File {
		names[i] = f.Name
		if f.Method != zip.Deflate {
			t.Errorf("%s: method = %d, want Deflate", f.Name, f.Method)
		}
		if f.Modified.IsZero() {
			t.Errorf("%s: Modified is zero", f.Name)
		}
	}
	if len(names) != 22 {
		t.Fatalf("got %d entries, want 22: %v", len(names), names)
	}

	want := append(schema.Files(), "index.xml", dtdFile)
	if !slices.Equal(names, want) {
		t.Errorf("entry order = %v, want %v", names, want)
	}
}

func TestZipSinkIsReproducible(t *testing.T) {
	var a, b bytes.Buffer
	exportEmpty(t, csvio.NewZipSink(&a), schema.Tables())
	exportEmpty(t, csvio.NewZipSink(&b), schema.Tables())
	if !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Error("two identical exports produced different archives")
	}
}

// writeRows fills two tables with rows and closes the export.
func writeRows(t *testing.T, sink csvio.Sink, tables []schema.Table) {
	t.Helper()

	e, err := csvio.NewExportWriter(sink, tables, referenceSupplier)
	if err != nil {
		t.Fatalf("NewExportWriter: %v", err)
	}

	w, err := e.Table("lines.csv")
	if err != nil {
		t.Fatalf("Table(lines.csv): %v", err)
	}
	again, err := e.Table("lines.csv")
	if err != nil {
		t.Fatalf("Table(lines.csv) again: %v", err)
	}
	if w != again {
		t.Error("Table returned a different Writer on the second call")
	}

	row := padRow(w.Table(), wildRow)
	if err := w.Write(row); err != nil {
		t.Fatalf("Write: %v", err)
	}
	mustClose(t, e)
}

// padRow stretches fields to the table's column count.
func padRow(tbl schema.Table, fields []string) []string {
	row := make([]string, len(tbl.Columns))
	for i := range row {
		if i < len(fields) {
			row[i] = fields[i]
		}
	}
	return row
}

func checkRoundTrip(t *testing.T, src csvio.Source) {
	t.Helper()

	tables, supplier, media, err := csvio.ReadIndex(src)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if supplier != referenceSupplier || media != referenceMedia {
		t.Errorf("supplier = %+v, media = %q", supplier, media)
	}

	names, err := src.Names()
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if len(names) != 22 {
		t.Errorf("got %d names, want 22", len(names))
	}

	tbl, ok := schema.TableByFile("lines.csv")
	if !ok {
		t.Fatal("lines.csv is not a DSFinV-K table")
	}
	if !slices.ContainsFunc(tables, func(x schema.Table) bool { return x.File == "lines.csv" }) {
		t.Fatal("index.xml does not describe lines.csv")
	}

	rc, err := src.Open("lines.csv")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = rc.Close() }()

	r, err := csvio.NewReader(rc, tbl)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	record, _, err := r.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if want := padRow(tbl, wildRow); !slices.Equal(record, want) {
		t.Errorf("record = %q, want %q", record, want)
	}
	if _, _, err := r.Read(); !errors.Is(err, io.EOF) {
		t.Errorf("second Read: %v, want EOF", err)
	}
}

func TestExportRoundTripDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "export")
	sink, err := csvio.NewDirSink(dir, false)
	if err != nil {
		t.Fatalf("NewDirSink: %v", err)
	}
	writeRows(t, sink, schema.Tables())

	src, err := csvio.OpenDir(dir)
	if err != nil {
		t.Fatalf("OpenDir: %v", err)
	}
	checkRoundTrip(t, src)
}

func TestExportRoundTripZip(t *testing.T) {
	var buf bytes.Buffer
	writeRows(t, csvio.NewZipSink(&buf), schema.Tables())

	path := filepath.Join(t.TempDir(), "export.zip")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	src, err := csvio.OpenZip(path)
	if err != nil {
		t.Fatalf("OpenZip: %v", err)
	}
	closer, ok := src.(io.Closer)
	if !ok {
		t.Fatal("OpenZip source is not an io.Closer")
	}
	defer func() { _ = closer.Close() }()

	checkRoundTrip(t, src)

	mem, err := csvio.NewZipSource(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewZipSource: %v", err)
	}
	checkRoundTrip(t, mem)
}

// extendedTables replaces tse.csv with a version carrying one extra column.
func extendedTables(t *testing.T) []schema.Table {
	t.Helper()

	base, ok := schema.TableByFile("tse.csv")
	if !ok {
		t.Fatal("tse.csv is not a DSFinV-K table")
	}
	extended, err := base.WithExtraColumns(schema.Column{
		Name:        "TSE_ZERTIFIKAT_III",
		Description: "Drittes Teilzertifikat",
		Type:        schema.ColumnAlphaNumeric,
		MaxLength:   1000,
	})
	if err != nil {
		t.Fatalf("WithExtraColumns: %v", err)
	}

	tables := schema.Tables()
	for i, tbl := range tables {
		if tbl.File == "tse.csv" {
			tables[i] = extended
		}
	}
	return tables
}

func TestExportExtraColumn(t *testing.T) {
	tables := extendedTables(t)
	dir := filepath.Join(t.TempDir(), "export")
	sink, err := csvio.NewDirSink(dir, false)
	if err != nil {
		t.Fatalf("NewDirSink: %v", err)
	}
	exportEmpty(t, sink, tables)

	src, err := csvio.OpenDir(dir)
	if err != nil {
		t.Fatalf("OpenDir: %v", err)
	}
	got, _, _, err := csvio.ReadIndex(src)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}

	i := slices.IndexFunc(got, func(x schema.Table) bool { return x.File == "tse.csv" })
	if i < 0 {
		t.Fatal("index.xml does not describe tse.csv")
	}
	if n := len(got[i].Columns); n != 12 {
		t.Errorf("tse.csv has %d columns in index.xml, want 12", n)
	}

	header, _, _ := strings.Cut(string(readFile(t, dir, "tse.csv")), "\r\n")
	if n := len(strings.Split(header, ";")); n != 12 {
		t.Errorf("tse.csv header has %d fields, want 12", n)
	}
	if !strings.HasSuffix(header, ";TSE_ZERTIFIKAT_III") {
		t.Errorf("header = %q", header)
	}
}

func TestDirSinkGuardsAnExistingDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "old.csv"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := csvio.NewDirSink(dir, false); !errors.Is(err, csvio.ErrDirNotEmpty) {
		t.Fatalf("NewDirSink without overwrite: %v, want ErrDirNotEmpty", err)
	}
	sink, err := csvio.NewDirSink(dir, true)
	if err != nil {
		t.Fatalf("NewDirSink with overwrite: %v", err)
	}
	mustClose(t, sink)

	file := filepath.Join(dir, "old.csv")
	if _, err := csvio.NewDirSink(file, true); err == nil {
		t.Error("NewDirSink on a plain file: want an error")
	}
}

func TestExportWriterErrors(t *testing.T) {
	dir := t.TempDir()

	if _, err := csvio.NewExportWriter(nopSink{}, nil, referenceSupplier); !errors.Is(err, csvio.ErrNoTables) {
		t.Errorf("empty tables: %v, want ErrNoTables", err)
	}

	dup := []schema.Table{{File: "a.csv"}, {File: "a.csv"}}
	if _, err := csvio.NewExportWriter(nopSink{}, dup, referenceSupplier); !errors.Is(err, csvio.ErrDuplicateFile) {
		t.Errorf("duplicate file: %v, want ErrDuplicateFile", err)
	}

	sink, err := csvio.NewDirSink(filepath.Join(dir, "export"), false)
	if err != nil {
		t.Fatalf("NewDirSink: %v", err)
	}
	e, err := csvio.NewExportWriter(sink, schema.Tables(), referenceSupplier)
	if err != nil {
		t.Fatalf("NewExportWriter: %v", err)
	}
	if _, err := e.Table("nope.csv"); !errors.Is(err, csvio.ErrUnknownFile) {
		t.Errorf("unknown file: %v, want ErrUnknownFile", err)
	}
	mustClose(t, e)
	if _, err := e.Table("lines.csv"); !errors.Is(err, csvio.ErrExportClosed) {
		t.Errorf("Table after Close: %v, want ErrExportClosed", err)
	}
	if err := e.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestExportWriterReportsSinkFailure(t *testing.T) {
	e, err := csvio.NewExportWriter(failingSink{}, schema.Tables(), referenceSupplier)
	if err != nil {
		t.Fatalf("NewExportWriter: %v", err)
	}
	if _, err := e.Table("lines.csv"); err == nil {
		t.Fatal("Table: want the sink error")
	}
	first := e.Close()
	if first == nil {
		t.Fatal("Close: want the sink error")
	}
	if second := e.Close(); !errors.Is(second, first) {
		t.Errorf("second Close = %v, want %v", second, first)
	}
}

func TestZipSinkReportsWriteFailure(t *testing.T) {
	boom := errors.New("boom")
	sink := csvio.NewZipSink(failingWriter{err: boom})

	e, err := csvio.NewExportWriter(sink, schema.Tables(), referenceSupplier)
	if err != nil {
		t.Fatalf("NewExportWriter: %v", err)
	}
	first := e.Close()
	if !errors.Is(first, boom) {
		t.Fatalf("Close = %v, want %v", first, boom)
	}
	if again := sink.Close(); !errors.Is(again, boom) {
		t.Errorf("second sink Close = %v, want %v", again, boom)
	}
}

func TestZipSinkClosesOnlyOnce(t *testing.T) {
	var buf bytes.Buffer
	sink := csvio.NewZipSink(&buf)
	exportEmpty(t, sink, schema.Tables())

	size := buf.Len()
	if err := sink.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if buf.Len() != size {
		t.Errorf("second Close wrote %d more bytes", buf.Len()-size)
	}
}

func TestZipSourceErrors(t *testing.T) {
	if _, err := csvio.NewZipSource(strings.NewReader("not a zip"), 9); err == nil {
		t.Fatal("NewZipSource on garbage: want an error")
	}

	var buf bytes.Buffer
	exportEmpty(t, csvio.NewZipSink(&buf), schema.Tables())

	src, err := csvio.NewZipSource(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewZipSource: %v", err)
	}
	if _, err := src.Open("nope.csv"); err == nil {
		t.Error("Open on a missing entry: want an error")
	}
	if closer, ok := src.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}
}

type nopSink struct{}

func (nopSink) Create(string) (io.WriteCloser, error) { return nopWriteCloser{}, nil }
func (nopSink) Close() error                          { return nil }

type nopWriteCloser struct{}

func (nopWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (nopWriteCloser) Close() error                { return nil }

type failingSink struct{}

func (failingSink) Create(string) (io.WriteCloser, error) { return nil, errors.New("no room") }
func (failingSink) Close() error                          { return errors.New("no room") }

func TestSourceErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	if _, err := csvio.OpenDir(missing); err == nil {
		t.Error("OpenDir on a missing directory: want an error")
	}
	if _, err := csvio.OpenZip(missing); err == nil {
		t.Error("OpenZip on a missing file: want an error")
	}

	file := filepath.Join(t.TempDir(), "plain.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := csvio.OpenDir(file); err == nil {
		t.Error("OpenDir on a plain file: want an error")
	}

	empty := t.TempDir()
	src, err := csvio.OpenDir(empty)
	if err != nil {
		t.Fatalf("OpenDir: %v", err)
	}
	if _, _, _, err := csvio.ReadIndex(src); err == nil {
		t.Error("ReadIndex without index.xml: want an error")
	}
	if _, err := src.Open("nope.csv"); err == nil {
		t.Error("Open on a missing file: want an error")
	}
}
