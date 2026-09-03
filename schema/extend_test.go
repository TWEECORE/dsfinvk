package schema_test

import (
	"errors"
	"testing"

	"github.com/tweecore/dsfinvk/schema"
)

func tseTable(t *testing.T) schema.Table {
	t.Helper()

	tbl, ok := schema.TableByFile("tse.csv")
	if !ok {
		t.Fatal("TableByFile(tse.csv) not found")
	}
	return tbl
}

func TestWithExtraColumnsAppends(t *testing.T) {
	base := tseTable(t)
	extra := schema.Column{
		Name:        "TSE_ZERTIFIKAT_III",
		Description: "Weiterer Rest des base64-codierten Zertifikats der TSE",
		Type:        schema.ColumnAlphaNumeric,
		MaxLength:   1000,
	}

	got, err := base.WithExtraColumns(extra)
	if err != nil {
		t.Fatalf("WithExtraColumns: %v", err)
	}
	if want := len(base.Columns) + 1; len(got.Columns) != want {
		t.Fatalf("len(Columns) = %d, want %d", len(got.Columns), want)
	}
	if last := got.Columns[len(got.Columns)-1]; last != extra {
		t.Errorf("last column = %+v, want %+v", last, extra)
	}
	for i, c := range base.Columns {
		if got.Columns[i] != c {
			t.Errorf("column %d = %+v, want %+v", i, got.Columns[i], c)
		}
	}
}

func TestWithExtraColumnsIsADeepCopy(t *testing.T) {
	base := tseTable(t)
	got, err := base.WithExtraColumns(schema.Column{Name: "X", Type: schema.ColumnNumeric})
	if err != nil {
		t.Fatalf("WithExtraColumns: %v", err)
	}

	got.Columns[0].Name = "TAMPERED"
	if base.Columns[0].Name != "Z_KASSE_ID" {
		t.Errorf("receiver column 0 = %q, want Z_KASSE_ID", base.Columns[0].Name)
	}
	fresh := tseTable(t)
	if fresh.Columns[0].Name != "Z_KASSE_ID" {
		t.Errorf("package table column 0 = %q, want Z_KASSE_ID", fresh.Columns[0].Name)
	}
}

func TestWithExtraColumnsWithoutArgumentsCopies(t *testing.T) {
	base := tseTable(t)
	got, err := base.WithExtraColumns()
	if err != nil {
		t.Fatalf("WithExtraColumns: %v", err)
	}
	if len(got.Columns) != len(base.Columns) {
		t.Fatalf("len(Columns) = %d, want %d", len(got.Columns), len(base.Columns))
	}
	got.Columns[1].Name = "TAMPERED"
	if base.Columns[1].Name == "TAMPERED" {
		t.Error("WithExtraColumns() shares the Columns backing array")
	}
}

func TestWithExtraColumnsMultiple(t *testing.T) {
	base := tseTable(t)
	got, err := base.WithExtraColumns(
		schema.Column{Name: "TSE_ZERTIFIKAT_III", Type: schema.ColumnAlphaNumeric, MaxLength: 1000},
		schema.Column{Name: "TSE_ZERTIFIKAT_IV", Type: schema.ColumnAlphaNumeric, MaxLength: 1000},
	)
	if err != nil {
		t.Fatalf("WithExtraColumns: %v", err)
	}
	if _, ok := got.ColumnIndex("TSE_ZERTIFIKAT_IV"); !ok {
		t.Error("TSE_ZERTIFIKAT_IV is missing")
	}
	if i, _ := got.ColumnIndex("TSE_ZERTIFIKAT_III"); i != len(base.Columns) {
		t.Errorf("TSE_ZERTIFIKAT_III index = %d, want %d", i, len(base.Columns))
	}
}

func TestWithExtraColumnsRejectsDuplicates(t *testing.T) {
	base := tseTable(t)

	tests := map[string][]schema.Column{
		"duplicate of an existing column": {{Name: "TSE_SERIAL", Type: schema.ColumnAlphaNumeric, MaxLength: 68}},
		"duplicate among the new columns": {
			{Name: "EXTRA", Type: schema.ColumnNumeric},
			{Name: "EXTRA", Type: schema.ColumnNumeric},
		},
	}
	for name, cols := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := base.WithExtraColumns(cols...); !errors.Is(err, schema.ErrDuplicateColumn) {
				t.Errorf("err = %v, want ErrDuplicateColumn", err)
			}
		})
	}
}

func TestWithExtraColumnsRejectsAnEmptyName(t *testing.T) {
	base := tseTable(t)
	if _, err := base.WithExtraColumns(schema.Column{Type: schema.ColumnNumeric}); !errors.Is(err, schema.ErrEmptyColumnName) {
		t.Errorf("err = %v, want ErrEmptyColumnName", err)
	}
}

func TestWithExtraColumnsLeavesTheReceiverUnchanged(t *testing.T) {
	base := tseTable(t)
	before := len(base.Columns)
	if _, err := base.WithExtraColumns(schema.Column{Name: "EXTRA", Type: schema.ColumnNumeric}); err != nil {
		t.Fatalf("WithExtraColumns: %v", err)
	}
	if len(base.Columns) != before {
		t.Errorf("len(receiver.Columns) = %d, want %d", len(base.Columns), before)
	}
}
