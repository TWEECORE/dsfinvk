package csvio_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/schema"
)

func testTable(cols ...string) schema.Table {
	t := schema.Table{File: "test.csv", Name: "Test", Description: "test.csv"}
	for _, name := range cols {
		t.Columns = append(t.Columns, schema.Column{Name: name, Type: schema.ColumnAlphaNumeric, MaxLength: 50})
	}
	return t
}

func writeAll(t *testing.T, tbl schema.Table, rows ...[]string) string {
	t.Helper()

	var buf bytes.Buffer
	w := csvio.NewWriter(&buf, tbl)
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			t.Fatalf("Write(%q): %v", row, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatalf("Error: %v", err)
	}
	return buf.String()
}

func TestWriterHeader(t *testing.T) {
	var buf bytes.Buffer
	w := csvio.NewWriter(&buf, testTable("A", "B", "C"))
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	w.Flush()

	if got, want := buf.String(), "A;B;C\r\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestWriterWritesNoBOM(t *testing.T) {
	out := writeAll(t, testTable("A"), []string{"x"})
	if strings.HasPrefix(out, "\ufeff") {
		t.Error("output starts with a BOM")
	}
	if got, want := out, "A\r\nx\r\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestWriterWritesTheHeaderOnFirstWrite(t *testing.T) {
	out := writeAll(t, testTable("A", "B"), []string{"1", "2"}, []string{"3", "4"})
	if got, want := out, "A;B\r\n1;2\r\n3;4\r\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestWriterHeaderOnlyOnce(t *testing.T) {
	var buf bytes.Buffer
	w := csvio.NewWriter(&buf, testTable("A"))
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if err := w.WriteHeader(); !errors.Is(err, csvio.ErrHeaderWritten) {
		t.Errorf("second WriteHeader err = %v, want ErrHeaderWritten", err)
	}
	if err := w.Write([]string{"x"}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	w.Flush()
	if got, want := buf.String(), "A\r\nx\r\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestWriterRejectsAWrongFieldCount(t *testing.T) {
	var buf bytes.Buffer
	w := csvio.NewWriter(&buf, testTable("A", "B"))
	for _, row := range [][]string{{"only one"}, {"a", "b", "c"}, nil} {
		if err := w.Write(row); !errors.Is(err, csvio.ErrFieldCount) {
			t.Errorf("Write(%q) err = %v, want ErrFieldCount", row, err)
		}
	}
}

func TestWriterQuoting(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "abc", "abc"},
		{"empty", "", ""},
		{"umlauts", "Straße", "Straße"},
		{"typographic quotes", "“Hallo”", "“Hallo”"},
		{"decimal comma", "12,50", "12,50"},
		{"dot", "1.2", "1.2"},
		{"delimiter", "a;b", `"a;b"`},
		{"quote", `a"b`, `"a""b"`},
		{"only a quote", `"`, `""""`},
		{"linefeed", "a\nb", "\"a\nb\""},
		{"carriage return", "a\rb", "\"a\rb\""},
		{"leading space", " a", `" a"`},
		{"trailing space", "a ", `"a "`},
		{"leading tab", "\ta", "\"\ta\""},
		{"trailing tab", "a\t", "\"a\t\""},
		{"single space", " ", `" "`},
		{"inner space", "a b", "a b"},
		{"everything", "a;\"b\nc ", "\"a;\"\"b\nc \""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := writeAll(t, testTable("A"), []string{tc.in})
			if got, want := out, "A\r\n"+tc.want+"\r\n"; got != want {
				t.Errorf("output = %q, want %q", got, want)
			}
		})
	}
}

func TestWriterEmptyTableStillWritesTheHeader(t *testing.T) {
	var buf bytes.Buffer
	w := csvio.NewWriter(&buf, testTable("A", "B"))
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	w.Flush()
	if got, want := buf.String(), "A;B\r\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

type failingWriter struct{ err error }

func (f failingWriter) Write([]byte) (int, error) { return 0, f.err }

func TestWriterKeepsTheFirstError(t *testing.T) {
	boom := errors.New("boom")
	w := csvio.NewWriter(failingWriter{err: boom}, testTable(strings.Repeat("A", 8192)))
	_ = w.Write([]string{"x"})
	w.Flush()
	if err := w.Error(); !errors.Is(err, boom) {
		t.Fatalf("Error = %v, want boom", err)
	}
	if err := w.Write([]string{"y"}); !errors.Is(err, boom) {
		t.Errorf("later Write err = %v, want boom", err)
	}
	if err := w.WriteHeader(); !errors.Is(err, boom) {
		t.Errorf("later WriteHeader err = %v, want boom", err)
	}
}

func TestWriterTableIsACopy(t *testing.T) {
	tbl := testTable("A", "B")
	var buf bytes.Buffer
	w := csvio.NewWriter(&buf, tbl)
	tbl.Columns[0].Name = "TAMPERED"
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	w.Flush()
	if got, want := buf.String(), "A;B\r\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
