package csvio_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/schema"
)

func numericColumn(name string, accuracy int) schema.Column {
	return schema.Column{Name: name, Type: schema.ColumnNumeric, Accuracy: accuracy}
}

func alphaColumn(name string, maxLength int) schema.Column {
	return schema.Column{Name: name, Type: schema.ColumnAlphaNumeric, MaxLength: maxLength}
}

func mustParse(t *testing.T, s string) dsfinvk.Decimal {
	t.Helper()

	d, err := dsfinvk.ParseComma(s)
	if err != nil {
		t.Fatalf("ParseComma(%q): %v", s, err)
	}
	return d
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		accuracy int
		in       string
		want     string
	}{
		// Accuracy 0 is a plain integer.
		{"accuracy 0 integer", 0, "42", "42"},
		{"accuracy 0 zero", 0, "0", "0"},
		{"accuracy 0 negative", 0, "-7", "-7"},

		// Accuracy 2 amounts always carry two digits. Spec 2.4 p.31.
		{"accuracy 2 whole", 2, "12", "12,00"},
		{"accuracy 2 one digit", 2, "12,5", "12,50"},
		{"accuracy 2 two digits", 2, "12,55", "12,55"},
		{"accuracy 2 negative", 2, "-0,01", "-0,01"},
		{"accuracy 2 zero", 2, "0", "0,00"},

		// Accuracy 3 quantities carry three digits. Spec 2.4 p.96.
		{"accuracy 3 whole", 3, "2", "2,000"},
		{"accuracy 3 one digit", 3, "1,5", "1,500"},
		{"accuracy 3 three digits", 3, "0,125", "0,125"},
		{"accuracy 3 negative", 3, "-2", "-2,000"},

		// Accuracy 5 amounts fall back to two digits. Spec 2.4 p.31, p.90.
		{"accuracy 5 whole", 5, "12", "12,00"},
		{"accuracy 5 one digit", 5, "12,5", "12,50"},
		{"accuracy 5 two digits", 5, "12,55", "12,55"},
		{"accuracy 5 three digits", 5, "12,555", "12,555"},
		{"accuracy 5 four digits", 5, "12,5555", "12,5555"},
		{"accuracy 5 five digits", 5, "12,55555", "12,55555"},
		{"accuracy 5 trailing zero kept minimal", 5, "12,50500", "12,505"},
		{"accuracy 5 negative", 5, "-0,00001", "-0,00001"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			col := numericColumn("VAL", tc.accuracy)
			got, err := csvio.FormatValue(col, mustParse(t, tc.in))
			if err != nil {
				t.Fatalf("FormatValue: %v", err)
			}
			if got != tc.want {
				t.Errorf("FormatValue(%s) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatValueRejectsTooManyFractionDigits(t *testing.T) {
	tests := []struct {
		name     string
		accuracy int
		in       string
	}{
		{"accuracy 0 with a fraction", 0, "1,5"},
		{"accuracy 2 with three digits", 2, "1,005"},
		{"accuracy 3 with four digits", 3, "1,0005"},
		{"accuracy 4 with five digits", 4, "0,00001"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := csvio.FormatValue(numericColumn("VAL", tc.accuracy), mustParse(t, tc.in))
			if !errors.Is(err, csvio.ErrAccuracy) {
				t.Errorf("err = %v, want ErrAccuracy", err)
			}
			if err != nil && !strings.Contains(err.Error(), "VAL") {
				t.Errorf("err = %v, want it to name the column", err)
			}
		})
	}
}

func TestFormatValueRejectsAnAlphaNumericColumn(t *testing.T) {
	if _, err := csvio.FormatValue(alphaColumn("NAME", 10), dsfinvk.Zero); !errors.Is(err, csvio.ErrColumnType) {
		t.Errorf("err = %v, want ErrColumnType", err)
	}
}

func TestFormatValueRejectsANegativeAccuracy(t *testing.T) {
	if _, err := csvio.FormatValue(numericColumn("VAL", -1), dsfinvk.Zero); !errors.Is(err, csvio.ErrAccuracy) {
		t.Errorf("err = %v, want ErrAccuracy", err)
	}
}

func TestFormatValueUsesTheRealSchemaColumns(t *testing.T) {
	lines, ok := schema.TableByFile("lines.csv")
	if !ok {
		t.Fatal("lines.csv not found")
	}
	menge, ok := lines.Column("MENGE")
	if !ok {
		t.Fatal("MENGE not found")
	}
	if got, err := csvio.FormatValue(menge, mustParse(t, "2")); err != nil || got != "2,000" {
		t.Errorf("FormatValue(MENGE, 2) = %q, %v, want \"2,000\", nil", got, err)
	}

	stkBr, ok := lines.Column("STK_BR")
	if !ok {
		t.Fatal("STK_BR not found")
	}
	if got, err := csvio.FormatValue(stkBr, mustParse(t, "5")); err != nil || got != "5,00" {
		t.Errorf("FormatValue(STK_BR, 5) = %q, %v, want \"5,00\", nil", got, err)
	}
	if got, err := csvio.FormatValue(stkBr, mustParse(t, "3,33333")); err != nil || got != "3,33333" {
		t.Errorf("FormatValue(STK_BR, 3,33333) = %q, %v, want \"3,33333\", nil", got, err)
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name     string
		accuracy int
		in       string
		want     string
	}{
		{"two digits", 2, "12,50", "12,50"},
		{"no fraction", 2, "12", "12,00"},
		{"negative", 2, "-0,01", "-0,01"},
		{"integer column", 0, "42", "42"},
		{"five digits", 5, "0,00001", "0,00001"},
		{"three digits", 3, "1,500", "1,500"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			col := numericColumn("VAL", tc.accuracy)
			got, ok, err := csvio.ParseValue(col, tc.in)
			if err != nil {
				t.Fatalf("ParseValue: %v", err)
			}
			if !ok {
				t.Fatal("ok = false, want true")
			}
			if !got.Equal(mustParse(t, tc.want)) {
				t.Errorf("ParseValue(%q) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseValueAcceptsAnEmptyField(t *testing.T) {
	got, ok, err := csvio.ParseValue(numericColumn("VAL", 2), "")
	if err != nil {
		t.Fatalf("ParseValue: %v", err)
	}
	if ok {
		t.Error("ok = true, want false for an empty field")
	}
	if !got.IsZero() {
		t.Errorf("value = %s, want 0", got)
	}
}

func TestParseValueRejects(t *testing.T) {
	tests := []struct {
		name     string
		accuracy int
		in       string
		want     error
	}{
		{"digit grouping", 2, "1.234,50", dsfinvk.ErrSyntax},
		{"decimal point", 2, "12.50", dsfinvk.ErrSyntax},
		{"letters", 2, "abc", dsfinvk.ErrSyntax},
		{"trailing space", 2, "12,50 ", dsfinvk.ErrSyntax},
		{"leading plus", 2, "+12,50", dsfinvk.ErrSyntax},
		{"more digits than accuracy", 2, "12,505", csvio.ErrAccuracy},
		{"padded beyond accuracy", 2, "12,500", csvio.ErrAccuracy},
		{"fraction in an integer column", 0, "12,0", csvio.ErrAccuracy},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := csvio.ParseValue(numericColumn("VAL", tc.accuracy), tc.in)
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestParseValueRejectsAnAlphaNumericColumn(t *testing.T) {
	if _, _, err := csvio.ParseValue(alphaColumn("NAME", 10), "1"); !errors.Is(err, csvio.ErrColumnType) {
		t.Errorf("err = %v, want ErrColumnType", err)
	}
}

func TestParseValueRoundTripsFormatValue(t *testing.T) {
	for _, accuracy := range []int{0, 2, 3, 5} {
		col := numericColumn("VAL", accuracy)
		for _, in := range []string{"0", "1", "-1", "123456"} {
			v := mustParse(t, in)
			s, err := csvio.FormatValue(col, v)
			if err != nil {
				t.Fatalf("FormatValue(%d, %s): %v", accuracy, in, err)
			}
			back, ok, err := csvio.ParseValue(col, s)
			if err != nil || !ok {
				t.Fatalf("ParseValue(%d, %q) = %v, %v", accuracy, s, ok, err)
			}
			if !back.Equal(v) {
				t.Errorf("round trip accuracy %d: %s became %s", accuracy, v, back)
			}
		}
	}
}

func TestCheckLength(t *testing.T) {
	tests := []struct {
		name    string
		col     schema.Column
		in      string
		wantErr bool
	}{
		{"exact fit", alphaColumn("NAME", 3), "abc", false},
		{"shorter", alphaColumn("NAME", 3), "a", false},
		{"empty", alphaColumn("NAME", 3), "", false},
		{"too long", alphaColumn("NAME", 3), "abcd", true},
		{"runes not bytes", alphaColumn("NAME", 3), "äöü", false},
		{"runes not bytes too long", alphaColumn("NAME", 3), "äöüß", true},
		{"typographic quotes", alphaColumn("NAME", 3), "“a”", false},
		{"numeric column is unlimited", numericColumn("VAL", 2), strings.Repeat("9", 100), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := csvio.CheckLength(tc.col, tc.in)
			if tc.wantErr {
				if !errors.Is(err, csvio.ErrTooLong) {
					t.Errorf("err = %v, want ErrTooLong", err)
				}
				if err != nil && !strings.Contains(err.Error(), tc.col.Name) {
					t.Errorf("err = %v, want it to name the column", err)
				}
				return
			}
			if err != nil {
				t.Errorf("err = %v, want nil", err)
			}
		})
	}
}
