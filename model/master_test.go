package model_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

// testTSE is a TSE with a short certificate.
func testTSE() model.TSE {
	return model.TSE{
		ID:         1,
		Serial:     "abc123",
		SigAlgo:    schema.TSESigAlgoECDSAPlainSHA256,
		Zeitformat: schema.TSEZeitformatUnixTime,
		PDEncoding: schema.TSEPDEncodingUTF8,
		PublicKey:  "cHVibGlj",
		Zertifikat: "Y2VydA==",
	}
}

func TestBuildLocationRow(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Standort = model.Location{
		Name: "Filiale Mitte", Strasse: "Nebenstr. 2", PLZ: "10117",
		Ort: "Berlin", Land: "DEU", UStID: "DE123456789",
	}

	row := onlyRow(t, buildOne(t, c), "location.csv")

	want := map[string]string{
		"Z_KASSE_ID":  "K1",
		"Z_NR":        "1",
		"LOC_NAME":    "Filiale Mitte",
		"LOC_STRASSE": "Nebenstr. 2",
		"LOC_PLZ":     "10117",
		"LOC_ORT":     "Berlin",
		"LOC_LAND":    "DEU",
		"LOC_USTID":   "DE123456789",
	}
	for column, w := range want {
		if got := field(t, "location.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildCashregisterRow(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Kasse = model.Kasse{
		Brand: "Acme", Modell: "X1", Seriennr: "SN-1", SWBrand: "AcmePOS",
		SWVersion: "3.2", Basiswaehrung: "EUR", KeineUStZuordnung: true,
	}

	row := onlyRow(t, buildOne(t, c), "cashregister.csv")

	want := map[string]string{
		"KASSE_BRAND":          "Acme",
		"KASSE_MODELL":         "X1",
		"KASSE_SERIENNR":       "SN-1",
		"KASSE_SW_BRAND":       "AcmePOS",
		"KASSE_SW_VERSION":     "3.2",
		"KASSE_BASISWAEH_CODE": "EUR",
		"KEINE_UST_ZUORDNUNG":  "1",
	}
	for column, w := range want {
		if got := field(t, "cashregister.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildCashregisterWritesFalseFlag(t *testing.T) {
	t.Parallel()

	row := onlyRow(t, buildOne(t, minimalClosing()), "cashregister.csv")

	if got := field(t, "cashregister.csv", row, "KEINE_UST_ZUORDNUNG"); got != "0" {
		t.Errorf("KEINE_UST_ZUORDNUNG = %q, want 0", got)
	}
}

func TestBuildTerminalRows(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Terminals = []model.Terminal{
		{ID: "T1", Brand: "Acme", Modell: "S1", Seriennr: "TS-1", SWBrand: "AcmeSlave", SWVersion: "1.0"},
		{ID: "T2"},
	}

	rows := buildOne(t, c)["slaves.csv"]
	if len(rows) != 2 {
		t.Fatalf("slaves.csv has %d rows, want 2", len(rows))
	}

	want := map[string]string{
		"TERMINAL_ID":         "T1",
		"TERMINAL_BRAND":      "Acme",
		"TERMINAL_MODELL":     "S1",
		"TERMINAL_SERIENNR":   "TS-1",
		"TERMINAL_SW_BRAND":   "AcmeSlave",
		"TERMINAL_SW_VERSION": "1.0",
	}
	for column, w := range want {
		if got := field(t, "slaves.csv", rows[0], column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
	if got := field(t, "slaves.csv", rows[1], "TERMINAL_ID"); got != "T2" {
		t.Errorf("second TERMINAL_ID = %q, want T2", got)
	}
}

func TestBuildAgenturRow(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Agenturen = []model.Agentur{{
		ID: 7,
		Address: model.Address{
			Name: "Lotto AG", Strasse: "Losweg 3", PLZ: "20095", Ort: "Hamburg",
			Land: "DEU", StNr: "22/333/44444", UStID: "DE987654321",
		},
	}}

	row := onlyRow(t, buildOne(t, c), "pa.csv")

	want := map[string]string{
		"AGENTUR_ID":      "7",
		"AGENTUR_NAME":    "Lotto AG",
		"AGENTUR_STRASSE": "Losweg 3",
		"AGENTUR_PLZ":     "20095",
		"AGENTUR_ORT":     "Hamburg",
		"AGENTUR_LAND":    "DEU",
		"AGENTUR_STNR":    "22/333/44444",
		"AGENTUR_USTID":   "DE987654321",
	}
	for column, w := range want {
		if got := field(t, "pa.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildRejectsAgenturIDBelowOne(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Agenturen = []model.Agentur{{ID: 0}}

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrAgenturID) {
		t.Fatalf("Build error = %v, want ErrAgenturID", err)
	}
}

func TestBuildVATRow(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.USt = []model.USt{{Schluessel: 1, Satz: dsfinvk.FromCents(1900), Beschreibung: "Allgemeiner Steuersatz"}}

	row := onlyRow(t, buildOne(t, c), "vat.csv")

	want := map[string]string{
		"UST_SCHLUESSEL": "1",
		"UST_SATZ":       "19,00",
		"UST_BESCHR":     "Allgemeiner Steuersatz",
	}
	for column, w := range want {
		if got := field(t, "vat.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildTSERow(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.TSEs = []model.TSE{testTSE()}

	row := onlyRow(t, buildOne(t, c), "tse.csv")

	want := map[string]string{
		"TSE_ID":            "1",
		"TSE_SERIAL":        "abc123",
		"TSE_SIG_ALGO":      "ecdsa-plain-SHA256",
		"TSE_ZEITFORMAT":    "unixTime",
		"TSE_PD_ENCODING":   "UTF-8",
		"TSE_PUBLIC_KEY":    "cHVibGlj",
		"TSE_ZERTIFIKAT_I":  "Y2VydA==",
		"TSE_ZERTIFIKAT_II": "",
	}
	for column, w := range want {
		if got := field(t, "tse.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildTSESplitsCertificate(t *testing.T) {
	t.Parallel()

	cert := strings.Repeat("A", 1000) + strings.Repeat("B", 1000) + strings.Repeat("C", 5)

	c := minimalClosing()
	tse := testTSE()
	tse.Zertifikat = cert
	c.TSEs = []model.TSE{tse}

	rows, tables, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var table schema.Table
	for _, tbl := range tables {
		if tbl.File == "tse.csv" {
			table = tbl
		}
	}
	if _, ok := table.ColumnIndex("TSE_ZERTIFIKAT_III"); !ok {
		t.Fatalf("tse.csv has no TSE_ZERTIFIKAT_III column")
	}
	if _, ok := table.ColumnIndex("TSE_ZERTIFIKAT_IV"); ok {
		t.Errorf("tse.csv unexpectedly has TSE_ZERTIFIKAT_IV")
	}

	row := rows["tse.csv"][0]
	if got := fieldOf(t, table, row, "TSE_ZERTIFIKAT_I"); got != strings.Repeat("A", 1000) {
		t.Errorf("TSE_ZERTIFIKAT_I = %d chars, want 1000 A", len(got))
	}
	if got := fieldOf(t, table, row, "TSE_ZERTIFIKAT_II"); got != strings.Repeat("B", 1000) {
		t.Errorf("TSE_ZERTIFIKAT_II = %d chars, want 1000 B", len(got))
	}
	if got := fieldOf(t, table, row, "TSE_ZERTIFIKAT_III"); got != "CCCCC" {
		t.Errorf("TSE_ZERTIFIKAT_III = %q, want CCCCC", got)
	}
}

func TestBuildRejectsUnknownTSEEnum(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	tse := testTSE()
	tse.SigAlgo = "rsa-plain-SHA256"
	c.TSEs = []model.TSE{tse}

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrEnumValue) {
		t.Fatalf("Build error = %v, want ErrEnumValue", err)
	}
}

func TestBuildRejectsOverlongField(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.Unternehmen.Name = strings.Repeat("x", 61)

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, csvio.ErrTooLong) {
		t.Fatalf("Build error = %v, want csvio.ErrTooLong", err)
	}
}

func TestBuildRejectsTooManyFractionDigits(t *testing.T) {
	t.Parallel()

	satz, err := dsfinvk.ParseComma("19,001")
	if err != nil {
		t.Fatalf("ParseComma: %v", err)
	}

	c := minimalClosing()
	c.USt = []model.USt{{Schluessel: 1, Satz: satz}}

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, csvio.ErrAccuracy) {
		t.Fatalf("Build error = %v, want csvio.ErrAccuracy", err)
	}
}

func TestBuildNumbersCertificateColumnsInRoman(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	tse := testTSE()
	tse.Zertifikat = strings.Repeat("A", 4001)
	c.TSEs = []model.TSE{tse}

	_, tables, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var table schema.Table
	for _, tbl := range tables {
		if tbl.File == "tse.csv" {
			table = tbl
		}
	}
	for _, column := range []string{"TSE_ZERTIFIKAT_III", "TSE_ZERTIFIKAT_IV", "TSE_ZERTIFIKAT_V"} {
		if _, ok := table.ColumnIndex(column); !ok {
			t.Errorf("tse.csv has no %s column", column)
		}
	}
	if got := len(table.Columns); got != 14 {
		t.Errorf("tse.csv has %d columns, want 14", got)
	}
}
