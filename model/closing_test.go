package model_test

import (
	"testing"
	"time"

	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

// buildOne renders a single Kassenabschluss and fails the test on error.
func buildOne(t *testing.T, c model.Kassenabschluss) model.Rows {
	t.Helper()

	rows, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return rows
}

// onlyRow returns the single row of file and fails the test if there is not exactly one.
func onlyRow(t *testing.T, rows model.Rows, file string) []string {
	t.Helper()

	got := rows[file]
	if len(got) != 1 {
		t.Fatalf("%s has %d rows, want 1", file, len(got))
	}
	return got[0]
}

// field returns the named column of a row of a standard table.
func field(t *testing.T, file string, row []string, column string) string {
	t.Helper()

	table, ok := schema.TableByFile(file)
	if !ok {
		t.Fatalf("unknown file %s", file)
	}
	return fieldOf(t, table, row, column)
}

// fieldOf returns the named column of a row of table.
func fieldOf(t *testing.T, table schema.Table, row []string, column string) string {
	t.Helper()

	i, ok := table.ColumnIndex(column)
	if !ok {
		t.Fatalf("%s has no column %s", table.File, column)
	}
	if len(row) != len(table.Columns) {
		t.Fatalf("%s row has %d fields, want %d", table.File, len(row), len(table.Columns))
	}
	return row[i]
}

func TestBuildClosingRow(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Unternehmen = model.Address{
		Name: "Baecker B", Strasse: "Hauptstr. 1", PLZ: "10115", Ort: "Berlin",
		Land: "DEU", StNr: "12/345/67890",
	}
	c.Bons = []model.Bon{
		{ID: "A1", Nr: 1, Typ: schema.BonTypBeleg},
		{ID: "A2", Nr: 2, Typ: schema.BonTypAVTraining},
	}

	row := onlyRow(t, buildOne(t, c), "cashpointclosing.csv")

	want := map[string]string{
		"Z_KASSE_ID":        "K1",
		"Z_ERSTELLUNG":      "2019-01-21T18:30:55+01:00",
		"Z_NR":              "1",
		"Z_BUCHUNGSTAG":     "",
		"TAXONOMIE_VERSION": "2.4",
		"Z_START_ID":        "A1",
		"Z_ENDE_ID":         "A2",
		"NAME":              "Baecker B",
		"STRASSE":           "Hauptstr. 1",
		"PLZ":               "10115",
		"ORT":               "Berlin",
		"LAND":              "DEU",
		"STNR":              "12/345/67890",
		"USTID":             "",
		"Z_SE_ZAHLUNGEN":    "0,00",
		"Z_SE_BARZAHLUNGEN": "0,00",
	}
	for column, w := range want {
		if got := field(t, "cashpointclosing.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildClosingBuchungstagAndVersion(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Buchungstag = time.Date(2019, time.January, 22, 0, 0, 0, 0, time.UTC)
	c.TaxonomieVersion = "2.3"

	row := onlyRow(t, buildOne(t, c), "cashpointclosing.csv")

	if got := field(t, "cashpointclosing.csv", row, "Z_BUCHUNGSTAG"); got != "2019-01-22" {
		t.Errorf("Z_BUCHUNGSTAG = %q, want 2019-01-22", got)
	}
	if got := field(t, "cashpointclosing.csv", row, "TAXONOMIE_VERSION"); got != "2.3" {
		t.Errorf("TAXONOMIE_VERSION = %q, want 2.3", got)
	}
}

func TestBuildClosingRowPerAbschluss(t *testing.T) {
	t.Parallel()

	first := minimalClosing()
	second := minimalClosing()
	second.Nr = 2

	rows, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{first, second}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got := len(rows["cashpointclosing.csv"]); got != 2 {
		t.Fatalf("cashpointclosing.csv has %d rows, want 2", got)
	}
}
