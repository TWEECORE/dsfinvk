package csvio_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/schema"
)

func readAll(t *testing.T, tbl schema.Table, in string) ([][]string, []int) {
	t.Helper()

	r, err := csvio.NewReader(strings.NewReader(in), tbl)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	var (
		rows  [][]string
		lines []int
	)
	for {
		row, line, err := r.Read()
		if errors.Is(err, io.EOF) {
			return rows, lines
		}
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		rows = append(rows, row)
		lines = append(lines, line)
	}
}

func TestReaderReadsRecords(t *testing.T) {
	rows, lines := readAll(t, testTable("A", "B"), "A;B\r\n1;2\r\n3;4\r\n")
	if len(rows) != 2 {
		t.Fatalf("read %d rows, want 2", len(rows))
	}
	if got := strings.Join(rows[0], "|"); got != "1|2" {
		t.Errorf("row 0 = %q, want \"1|2\"", got)
	}
	if lines[0] != 2 || lines[1] != 3 {
		t.Errorf("lines = %v, want [2 3]", lines)
	}
}

func TestReaderAcceptsLineFeedOnlyRecords(t *testing.T) {
	rows, lines := readAll(t, testTable("A", "B"), "A;B\n1;2\n3;4\n")
	if len(rows) != 2 {
		t.Fatalf("read %d rows, want 2", len(rows))
	}
	if lines[1] != 3 {
		t.Errorf("line = %d, want 3", lines[1])
	}
}

func TestReaderAcceptsAMissingFinalRecordDelimiter(t *testing.T) {
	rows, _ := readAll(t, testTable("A"), "A\r\nx")
	if len(rows) != 1 || rows[0][0] != "x" {
		t.Errorf("rows = %v, want [[x]]", rows)
	}
}

func TestReaderStripsTheBOM(t *testing.T) {
	rows, _ := readAll(t, testTable("A", "B"), "\ufeffA;B\r\n1;2\r\n")
	if len(rows) != 1 {
		t.Fatalf("read %d rows, want 1", len(rows))
	}
}

func TestReaderWithoutABOM(t *testing.T) {
	if _, err := csvio.NewReader(strings.NewReader("A;B\r\n"), testTable("A", "B")); err != nil {
		t.Errorf("NewReader: %v", err)
	}
}

func TestReaderHeaderMismatch(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		contains []string
	}{
		{"wrong name", "A;X\r\n", []string{"position 2", `"B"`, `"X"`, "line 1"}},
		{"too few columns", "A\r\n", []string{"1", "2", "line 1"}},
		{"too many columns", "A;B;C\r\n", []string{"3", "2", "line 1"}},
		{"case differs", "a;B\r\n", []string{"position 1", `"A"`, `"a"`}},
		{"empty input", "", []string{"header"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := csvio.NewReader(strings.NewReader(tc.in), testTable("A", "B"))
			if !errors.Is(err, csvio.ErrHeader) {
				t.Fatalf("err = %v, want ErrHeader", err)
			}
			for _, want := range tc.contains {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

func TestReaderRejectsAWrongFieldCount(t *testing.T) {
	r, err := csvio.NewReader(strings.NewReader("A;B\r\n1;2\r\n3\r\n"), testTable("A", "B"))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if _, _, err = r.Read(); err != nil {
		t.Fatalf("Read: %v", err)
	}

	_, line, err := r.Read()
	if !errors.Is(err, csvio.ErrFieldCount) {
		t.Fatalf("err = %v, want ErrFieldCount", err)
	}
	if line != 3 {
		t.Errorf("line = %d, want 3", line)
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("err = %q, want it to name line 3", err)
	}
}

func TestReaderCountsPhysicalLines(t *testing.T) {
	in := "A;B\r\n1;2\r\n\"x\ny\";4\r\n5;6\r\n"
	rows, lines := readAll(t, testTable("A", "B"), in)
	if len(rows) != 3 {
		t.Fatalf("read %d rows, want 3", len(rows))
	}
	if rows[1][0] != "x\ny" {
		t.Errorf("row 1 field 0 = %q, want %q", rows[1][0], "x\ny")
	}
	if want := []int{2, 3, 5}; lines[0] != want[0] || lines[1] != want[1] || lines[2] != want[2] {
		t.Errorf("lines = %v, want %v", lines, want)
	}
}

func TestReaderUnquotes(t *testing.T) {
	rows, _ := readAll(t, testTable("A", "B"), "A;B\r\n\"a;b\";\"c\"\"d\"\r\n")
	if rows[0][0] != "a;b" {
		t.Errorf("field 0 = %q, want %q", rows[0][0], "a;b")
	}
	if rows[0][1] != `c"d` {
		t.Errorf("field 1 = %q, want %q", rows[0][1], `c"d`)
	}
}

func TestReaderKeepsLeadingSpace(t *testing.T) {
	rows, _ := readAll(t, testTable("A"), "A\r\n\" a \"\r\n")
	if rows[0][0] != " a " {
		t.Errorf("field = %q, want %q", rows[0][0], " a ")
	}
}

func TestReaderRejectsABareQuote(t *testing.T) {
	r, err := csvio.NewReader(strings.NewReader("A;B\r\n\"a\"b;c\r\n"), testTable("A", "B"))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if _, _, err = r.Read(); err == nil {
		t.Error("Read err = nil, want a parse error")
	}
}

func TestReaderReturnsEOF(t *testing.T) {
	r, err := csvio.NewReader(strings.NewReader("A\r\n"), testTable("A"))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	for range 2 {
		if _, _, err = r.Read(); !errors.Is(err, io.EOF) {
			t.Errorf("err = %v, want io.EOF", err)
		}
	}
}

func TestReaderTableIsACopy(t *testing.T) {
	tbl := testTable("A", "B")
	r, err := csvio.NewReader(strings.NewReader("A;B\r\n"), tbl)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	got := r.Table()
	got.Columns[0].Name = "TAMPERED"
	if r.Table().Columns[0].Name != "A" {
		t.Error("Table() shares its Columns backing array")
	}
}

func TestReaderRoundTripsTheWriter(t *testing.T) {
	tbl := testTable("A", "B", "C")
	rows := [][]string{
		{"plain", "", "Straße"},
		{"a;b", `c"d`, "e\nf"},
		{" lead", "trail ", "“typo”"},
	}
	out := writeAll(t, tbl, rows...)

	got, _ := readAll(t, tbl, out)
	if len(got) != len(rows) {
		t.Fatalf("read %d rows, want %d", len(got), len(rows))
	}
	for i := range rows {
		for j := range rows[i] {
			if got[i][j] != rows[i][j] {
				t.Errorf("row %d field %d = %q, want %q", i, j, got[i][j], rows[i][j])
			}
		}
	}
}
