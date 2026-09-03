package model_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tweecore/dsfinvk"

	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

// twoClosingExport is one Kasse closed twice: the Anhang F day, then a closing
// that fills every remaining column, with a certificate over 2000 characters.
func twoClosingExport() model.Export {
	first := anhangF()
	decorate(&first)

	second := fullClosing()

	return model.Export{Abschluesse: []model.Kassenabschluss{first, second}}
}

// decorate adds the master data every closing carries.
func decorate(c *model.Kassenabschluss) {
	c.Buchungstag = time.Date(2019, time.January, 22, 0, 0, 0, 0, time.UTC)
	c.Unternehmen = model.Address{
		Name: "Baecker B", Strasse: "Hauptstr. 1", PLZ: "10115", Ort: "Berlin",
		Land: "DEU", StNr: "12/345/67890", UStID: "DE123456789",
	}
	c.Standort = model.Location{
		Name: "Filiale Mitte", Strasse: "Nebenstr. 2", PLZ: "10117",
		Ort: "Berlin", Land: "DEU", UStID: "DE123456789",
	}
	c.Kasse = model.Kasse{
		Brand: "Acme", Modell: "X1", Seriennr: "SN-1", SWBrand: "AcmePOS",
		SWVersion: "3.2", Basiswaehrung: "EUR", KeineUStZuordnung: true,
	}
	c.Terminals = []model.Terminal{{
		ID: "T1", Brand: "Acme", Modell: "S1", Seriennr: "TS-1",
		SWBrand: "AcmeSlave", SWVersion: "1.0",
	}}
	c.Agenturen = []model.Agentur{{
		ID: 3,
		Address: model.Address{
			Name: "Lotto AG", Strasse: "Losweg 3", PLZ: "20095", Ort: "Hamburg",
			Land: "DEU", StNr: "22/333/44444", UStID: "DE987654321",
		},
	}}
	c.TSEs = []model.TSE{testTSE()}
	c.USt = []model.USt{
		{Schluessel: 1, Satz: dsfinvk.FromCents(1900), Beschreibung: "Allgemeiner Steuersatz"},
		{Schluessel: 2, Satz: dsfinvk.FromCents(700), Beschreibung: "Ermaessigter Steuersatz"},
		{Schluessel: 5, Satz: dsfinvk.Zero, Beschreibung: "Nicht steuerbar"},
	}
}

// fullClosing exercises every column of every table.
func fullClosing() model.Kassenabschluss {
	c := minimalClosing()
	c.Nr = 2
	decorate(&c)

	long := testTSE()
	long.Zertifikat = strings.Repeat("A", 1000) + strings.Repeat("B", 1000) + strings.Repeat("C", 42)
	c.TSEs = []model.TSE{long}

	position := model.Position{
		Zeile: "1", GutscheinNr: "4666", Artikeltext: "Fleisch", TerminalID: "T1",
		GVTyp: schema.GVTypUmsatz, GVName: "Theke", Inhaus: true, Storno: true,
		AgenturID: 3, ArtNr: "666431", GTIN: "4001234567890", WarengrID: "W1",
		Warengr: "Frisch", Menge: quantity(2000), Faktor: quantity(1500),
		Einheit: "kg", StkBrutto: euro(500),
		USt: []model.PosUSt{{Schluessel: 1, Brutto: euro(1000), Netto: euro(840), USt: euro(160)}},
		Preisfindung: []model.Preisfindung{
			{Typ: schema.PreisfindungBaseAmount, Schluessel: 1, Brutto: euro(1100), Netto: euro(924), USt: euro(176)},
		},
		Zusatzinfos: []model.Zusatzinfo{{
			ArtNr: "A1", GTIN: "4009", Name: "Cola", WarengrID: "W2", Warengr: "Getraenke",
			Menge: quantity(1000), Faktor: quantity(1000), Einheit: "Stueck", Schluessel: 1,
			BasispreisBrutto: euro(250), BasispreisNetto: euro(210), BasispreisUSt: euro(40),
		}},
	}

	c.Bons = []model.Bon{{
		ID: "100", Nr: 1, Typ: schema.BonTypBeleg, Name: "Kassenbeleg",
		TerminalID: "T1", Storno: true, Start: testStart, Ende: testEnde,
		BedienerID: "B1", BedienerName: "Anna", UmsBrutto: euro(1000),
		Kunde: model.Kunde{
			Name: "Kunde A", ID: "C1", Typ: "Mitarbeiter", Strasse: "Weg 4",
			PLZ: "10115", Ort: "Berlin", Land: "DEU", UStID: "DE111222333",
		},
		Notiz:             "Tisch 3",
		Abrechnungskreise: []string{"Tisch 10"},
		Zahlungen: []model.Zahlung{
			{Typ: schema.ZahlartTypBar, Name: "Kasse", BetragBasis: euro(500)},
			{Typ: schema.ZahlartTypBar, Name: "Devisen", Waehrung: "CHF", BetragWaehrung: euro(475), BetragBasis: euro(500)},
		},
		USt: []model.BonUSt{{Schluessel: 1, Brutto: euro(1000), Netto: euro(840), USt: euro(160)}},
		Referenzen: []model.Referenz{
			{PosZeile: "1", Typ: schema.RefTypTransaktion, Datum: testErstellung, KasseID: "K1", Nr: 1, BonID: "11"},
			{Typ: schema.RefTypExterneSonstige, Name: "Gutschrift", BonID: "RE-4711"},
		},
		TSE: &model.TSETransaktion{
			TSEID: 1, Nr: 13, Start: testStart, Ende: testEnde,
			Vorgangsart: schema.TSEVorgangsartKassenbelegV1, SigZaehler: 13,
			Signatur: "BEYwRAIg", Fehler: "keiner", Vorgangsdaten: "Beleg^10.00_0.00",
		},
		Positionen: []model.Position{position},
	}}
	return c
}

func TestWriteExportRoundTrip(t *testing.T) {
	t.Parallel()

	e := twoClosingExport()
	want, tables, err := model.Build(e)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	dir := filepath.Join(t.TempDir(), "export")
	sink, err := csvio.NewDirSink(dir, false)
	if err != nil {
		t.Fatalf("NewDirSink: %v", err)
	}
	if werr := model.WriteExport(e, sink, csvio.DataSupplier{Name: "Baecker B"}); werr != nil {
		t.Fatalf("WriteExport: %v", werr)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != len(schema.Tables())+2 {
		t.Errorf("export holds %d files, want %d", len(entries), len(schema.Tables())+2)
	}

	src, err := csvio.OpenDir(dir)
	if err != nil {
		t.Fatalf("OpenDir: %v", err)
	}
	readTables, supplier, _, err := csvio.ReadIndex(src)
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}
	if supplier.Name != "Baecker B" {
		t.Errorf("DataSupplier.Name = %q, want Baecker B", supplier.Name)
	}
	if len(readTables) != len(tables) {
		t.Fatalf("index.xml has %d tables, want %d", len(readTables), len(tables))
	}

	for i, table := range readTables {
		if table.File != tables[i].File {
			t.Fatalf("table %d is %s, want %s", i, table.File, tables[i].File)
		}
		if table.File == "tse.csv" {
			if _, ok := table.ColumnIndex("TSE_ZERTIFIKAT_III"); !ok {
				t.Errorf("index.xml tse.csv has no TSE_ZERTIFIKAT_III")
			}
		}
		if got := readAll(t, src, table); !reflect.DeepEqual(got, want[table.File]) {
			t.Errorf("%s read back as\n%v\nwant\n%v", table.File, got, want[table.File])
		}
	}
}

// readAll reads every data row of one CSV file of the export.
func readAll(t *testing.T, src csvio.Source, table schema.Table) [][]string {
	t.Helper()

	rc, err := src.Open(table.File)
	if err != nil {
		t.Fatalf("Open %s: %v", table.File, err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			t.Errorf("Close %s: %v", table.File, cerr)
		}
	}()

	reader, err := csvio.NewReader(rc, table)
	if err != nil {
		t.Fatalf("NewReader %s: %v", table.File, err)
	}

	var out [][]string
	for {
		record, _, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return out
		}
		if err != nil {
			t.Fatalf("Read %s: %v", table.File, err)
		}
		out = append(out, record)
	}
}

func TestWriteExportReportsBuildError(t *testing.T) {
	t.Parallel()

	e := model.Export{Abschluesse: []model.Kassenabschluss{{Nr: 1}}}

	sink, err := csvio.NewDirSink(filepath.Join(t.TempDir(), "export"), false)
	if err != nil {
		t.Fatalf("NewDirSink: %v", err)
	}
	if err := model.WriteExport(e, sink, csvio.DataSupplier{}); !errors.Is(err, model.ErrKasseID) {
		t.Fatalf("WriteExport error = %v, want ErrKasseID", err)
	}
}

func TestBuildFillsEveryColumn(t *testing.T) {
	t.Parallel()

	rows, tables, err := model.Build(twoClosingExport())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(rows) != len(tables) {
		t.Fatalf("Rows has %d files, want %d", len(rows), len(tables))
	}

	filled := make(map[string]map[string]bool, len(tables))
	for _, table := range tables {
		data, ok := rows[table.File]
		if !ok {
			t.Fatalf("Rows has no entry for %s", table.File)
		}
		filled[table.File] = make(map[string]bool, len(table.Columns))
		for _, row := range data {
			if len(row) != len(table.Columns) {
				t.Fatalf("%s row has %d fields, want %d", table.File, len(row), len(table.Columns))
			}
			for i, c := range table.Columns {
				if row[i] != "" {
					filled[table.File][c.Name] = true
				}
			}
		}
	}

	// Every column must be reachable, so the fixture has to fill each one at least once.
	for _, table := range tables {
		for _, c := range table.Columns {
			if !filled[table.File][c.Name] {
				t.Errorf("%s: column %s never filled", table.File, c.Name)
			}
		}
	}
}
