package schema_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk/schema"
)

var wantOrder = []string{
	"cashpointclosing.csv",
	"location.csv",
	"cashregister.csv",
	"slaves.csv",
	"pa.csv",
	"tse.csv",
	"vat.csv",
	"businesscases.csv",
	"payment.csv",
	"cash_per_currency.csv",
	"transactions.csv",
	"datapayment.csv",
	"lines.csv",
	"itemamounts.csv",
	"subitems.csv",
	"transactions_tse.csv",
	"transactions_vat.csv",
	"lines_vat.csv",
	"allocation_groups.csv",
	"references.csv",
}

func TestTableCount(t *testing.T) {
	if got, want := len(schema.Tables()), 20; got != want {
		t.Errorf("len(Tables) = %d, want %d", got, want)
	}
}

func TestTotalColumnCount(t *testing.T) {
	total := 0
	for _, tbl := range schema.Tables() {
		total += len(tbl.Columns)
	}
	if got, want := total, 219; got != want {
		t.Errorf("total columns = %d, want %d", got, want)
	}
}

func TestTableOrder(t *testing.T) {
	got := schema.Files()
	if !slices.Equal(got, wantOrder) {
		t.Errorf("Files() =\n%v\nwant\n%v", got, wantOrder)
	}
}

func TestTablesReturnsADeepCopy(t *testing.T) {
	a := schema.Tables()
	a[0].File = "tampered.csv"
	a[0].Columns[0].Name = "TAMPERED"
	b := schema.Tables()
	if b[0].File != wantOrder[0] {
		t.Errorf("Tables() shares its backing array: File = %q", b[0].File)
	}
	if b[0].Columns[0].Name != "Z_KASSE_ID" {
		t.Errorf("Tables() shares its Columns slices: Name = %q", b[0].Columns[0].Name)
	}
}

func TestTableByFileReturnsADeepCopy(t *testing.T) {
	a, ok := schema.TableByFile("transactions.csv")
	if !ok {
		t.Fatal("transactions.csv not found")
	}
	a.Columns[0].Name = "TAMPERED"
	b, _ := schema.TableByFile("transactions.csv")
	if b.Columns[0].Name != "Z_KASSE_ID" {
		t.Errorf("TableByFile shares its Columns slice: Name = %q", b.Columns[0].Name)
	}
}

func TestFilesReturnsACopy(t *testing.T) {
	a := schema.Files()
	a[0] = "tampered.csv"
	if b := schema.Files(); b[0] != wantOrder[0] {
		t.Errorf("Files() returns a shared slice: got %q after mutation", b[0])
	}
}

func TestEveryFileEndsInCSV(t *testing.T) {
	for _, tbl := range schema.Tables() {
		if !strings.HasSuffix(tbl.File, ".csv") {
			t.Errorf("table %q: File does not end in .csv", tbl.File)
		}
		if tbl.Name == "" {
			t.Errorf("table %q: empty Name", tbl.File)
		}
	}
}

func TestKeyColumnsComeFirst(t *testing.T) {
	want := []schema.Column{
		{Name: "Z_KASSE_ID", Type: schema.ColumnAlphaNumeric, MaxLength: 50},
		{Name: "Z_ERSTELLUNG", Type: schema.ColumnAlphaNumeric, MaxLength: 30},
		{Name: "Z_NR", Type: schema.ColumnNumeric},
	}
	for _, tbl := range schema.Tables() {
		if len(tbl.Columns) < 3 {
			t.Errorf("table %q has %d columns", tbl.File, len(tbl.Columns))
			continue
		}
		for i, w := range want {
			got := tbl.Columns[i]
			if got.Name != w.Name || got.Type != w.Type || got.MaxLength != w.MaxLength || got.Accuracy != 0 {
				t.Errorf("%s column %d = %+v, want Name %q Type %v MaxLength %d Accuracy 0",
					tbl.File, i, got, w.Name, w.Type, w.MaxLength)
			}
		}
	}
}

func TestColumnAccuracy(t *testing.T) {
	tests := []struct {
		file, column string
		accuracy     int
	}{
		{"cashpointclosing.csv", "Z_SE_ZAHLUNGEN", 2},
		{"cashpointclosing.csv", "Z_SE_BARZAHLUNGEN", 2},
		{"lines.csv", "STK_BR", 5},
		{"lines.csv", "MENGE", 3},
		{"cashpointclosing.csv", "Z_NR", 0},
	}
	for _, tc := range tests {
		t.Run(tc.file+"/"+tc.column, func(t *testing.T) {
			tbl, ok := schema.TableByFile(tc.file)
			if !ok {
				t.Fatalf("TableByFile(%q) not found", tc.file)
			}
			c, ok := tbl.Column(tc.column)
			if !ok {
				t.Fatalf("column %q not found", tc.column)
			}
			if c.Type != schema.ColumnNumeric {
				t.Errorf("Type = %v, want ColumnNumeric", c.Type)
			}
			if c.Accuracy != tc.accuracy {
				t.Errorf("Accuracy = %d, want %d", c.Accuracy, tc.accuracy)
			}
			if c.MaxLength != 0 {
				t.Errorf("MaxLength = %d, want 0 for a Numeric column", c.MaxLength)
			}
		})
	}
}

func TestColumnCountsPerTable(t *testing.T) {
	tests := map[string]int{
		"transactions.csv": 23,
		"references.csv":   11,
		"subitems.csv":     17,
	}
	for file, want := range tests {
		tbl, ok := schema.TableByFile(file)
		if !ok {
			t.Fatalf("TableByFile(%q) not found", file)
		}
		if got := len(tbl.Columns); got != want {
			t.Errorf("%s has %d columns, want %d", file, got, want)
		}
	}
}

// index.xml spells the column TSE_ZERTIFIKAT_I; Anhang G spells it TSE_Zertifikat_I.
func TestTSECertificateColumn(t *testing.T) {
	tbl, ok := schema.TableByFile("tse.csv")
	if !ok {
		t.Fatal("tse.csv not found")
	}
	c, ok := tbl.Column("TSE_ZERTIFIKAT_I")
	if !ok {
		t.Fatal("TSE_ZERTIFIKAT_I not found in tse.csv")
	}
	if c.Type != schema.ColumnAlphaNumeric {
		t.Errorf("Type = %v, want ColumnAlphaNumeric", c.Type)
	}
	if got, want := c.MaxLength, 1000; got != want {
		t.Errorf("MaxLength = %d, want %d", got, want)
	}
	if _, ok := tbl.Column("TSE_Zertifikat_I"); ok {
		t.Error("lookup is case-insensitive; it must not be")
	}
}

func TestTableByFile(t *testing.T) {
	tbl, ok := schema.TableByFile("transactions.csv")
	if !ok {
		t.Fatal("transactions.csv not found")
	}
	if got, want := tbl.Name, "Bonkopf"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if _, ok := schema.TableByFile("nope.csv"); ok {
		t.Error(`TableByFile("nope.csv") reported found`)
	}
	if _, ok := schema.TableByFile(""); ok {
		t.Error(`TableByFile("") reported found`)
	}
}

func TestColumnIndex(t *testing.T) {
	tbl, ok := schema.TableByFile("transactions.csv")
	if !ok {
		t.Fatal("transactions.csv not found")
	}
	for i, c := range tbl.Columns {
		got, ok := tbl.ColumnIndex(c.Name)
		if !ok {
			t.Errorf("ColumnIndex(%q) not found", c.Name)
			continue
		}
		if got != i {
			t.Errorf("ColumnIndex(%q) = %d, want %d", c.Name, got, i)
		}
	}
	if _, ok := tbl.ColumnIndex("NOT_A_COLUMN"); ok {
		t.Error(`ColumnIndex("NOT_A_COLUMN") reported found`)
	}
}

func TestColumn(t *testing.T) {
	tbl, ok := schema.TableByFile("lines.csv")
	if !ok {
		t.Fatal("lines.csv not found")
	}
	c, ok := tbl.Column("ART_NR")
	if !ok {
		t.Fatal("ART_NR not found")
	}
	if c.Name != "ART_NR" {
		t.Errorf("Name = %q", c.Name)
	}
	if _, ok := tbl.Column("NOT_A_COLUMN"); ok {
		t.Error(`Column("NOT_A_COLUMN") reported found`)
	}
}

func TestColumnTypeString(t *testing.T) {
	if got, want := schema.ColumnAlphaNumeric.String(), "AlphaNumeric"; got != want {
		t.Errorf("= %q, want %q", got, want)
	}
	if got, want := schema.ColumnNumeric.String(), "Numeric"; got != want {
		t.Errorf("= %q, want %q", got, want)
	}
	if got := schema.ColumnType(9).String(); got != "ColumnType(9)" {
		t.Errorf("= %q, want %q", got, "ColumnType(9)")
	}
}

func TestFormatConstants(t *testing.T) {
	if schema.ColumnDelimiter != ';' {
		t.Errorf("ColumnDelimiter = %q", schema.ColumnDelimiter)
	}
	if schema.TextEncapsulator != '"' {
		t.Errorf("TextEncapsulator = %q", schema.TextEncapsulator)
	}
	if schema.RecordDelimiter != "\r\n" {
		t.Errorf("RecordDelimiter = %q", schema.RecordDelimiter)
	}
	if schema.DecimalSymbol != ',' {
		t.Errorf("DecimalSymbol = %q", schema.DecimalSymbol)
	}
	if schema.DigitGroupingSymbol != '.' {
		t.Errorf("DigitGroupingSymbol = %q", schema.DigitGroupingSymbol)
	}
	if schema.Encoding != "UTF-8" {
		t.Errorf("Encoding = %q, want %q", schema.Encoding, "UTF-8")
	}
	if schema.HeaderRow != 1 {
		t.Errorf("HeaderRow = %d, want 1", schema.HeaderRow)
	}
	if schema.DataStartRow != 2 {
		t.Errorf("DataStartRow = %d", schema.DataStartRow)
	}
	if schema.HeaderRow >= schema.DataStartRow {
		t.Errorf("HeaderRow %d is not before DataStartRow %d", schema.HeaderRow, schema.DataStartRow)
	}
}

func TestSpecVersion(t *testing.T) {
	if got, want := schema.SpecVersion, "2.4"; got != want {
		t.Errorf("SpecVersion = %q, want %q", got, want)
	}
}

// TestTaxonomyVersions pins the writable versions (spec p.38) against the full
// Aenderungsnachweis table on spec p.1 to p.2.
func TestTaxonomyVersions(t *testing.T) {
	if got, want := schema.CurrentTaxonomyVersions(), []string{"2.3", "2.4"}; !slices.Equal(got, want) {
		t.Errorf("CurrentTaxonomyVersions() = %v, want %v", got, want)
	}
	want := []string{"1.0", "1.1", "2.0", "2.1", "2.2", "2.3", "2.4"}
	if got := schema.KnownTaxonomyVersions(); !slices.Equal(got, want) {
		t.Errorf("KnownTaxonomyVersions() = %v, want %v", got, want)
	}
	if !slices.Contains(schema.CurrentTaxonomyVersions(), schema.SpecVersion) {
		t.Error("SpecVersion is not among CurrentTaxonomyVersions()")
	}
	for _, v := range schema.CurrentTaxonomyVersions() {
		if !slices.Contains(schema.KnownTaxonomyVersions(), v) {
			t.Errorf("current version %q is not known", v)
		}
	}
}

func TestTaxonomyVersionsReturnCopies(t *testing.T) {
	a := schema.CurrentTaxonomyVersions()
	a[0] = "tampered"
	if b := schema.CurrentTaxonomyVersions(); b[0] != "2.3" {
		t.Errorf("CurrentTaxonomyVersions() returns a shared slice: got %q", b[0])
	}
	c := schema.KnownTaxonomyVersions()
	c[0] = "tampered"
	if d := schema.KnownTaxonomyVersions(); d[0] != "1.0" {
		t.Errorf("KnownTaxonomyVersions() returns a shared slice: got %q", d[0])
	}
}
