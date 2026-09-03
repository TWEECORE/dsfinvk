package model_test

import (
	"errors"
	"testing"
	"time"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

var (
	testStart = time.Date(2019, time.January, 21, 15, 45, 5, 0, time.FixedZone("CET", 3600))
	testEnde  = time.Date(2019, time.January, 21, 15, 45, 55, 0, time.FixedZone("CET", 3600))
)

// euro returns a Decimal from an amount in cents.
func euro(cents int64) dsfinvk.Decimal { return dsfinvk.FromCents(cents) }

// closingWithBons returns a Kassenabschluss carrying one VAT key and the given Bons.
func closingWithBons(bons ...model.Bon) model.Kassenabschluss {
	c := minimalClosing()
	c.Kasse.Basiswaehrung = "EUR"
	c.USt = []model.USt{
		{Schluessel: 1, Satz: dsfinvk.FromCents(1900)},
		{Schluessel: 2, Satz: dsfinvk.FromCents(700)},
		{Schluessel: 5, Satz: dsfinvk.Zero},
	}
	c.Bons = bons
	return c
}

func TestBuildTransactionRow(t *testing.T) {
	t.Parallel()

	bon := model.Bon{
		ID: "40062", Nr: 62, Typ: schema.BonTypBeleg, Name: "Kassenbeleg",
		TerminalID: "T1", Storno: false, Start: testStart, Ende: testEnde,
		BedienerID: "B1", BedienerName: "Anna", UmsBrutto: euro(5000),
		Kunde: model.Kunde{
			Name: "Kunde A", ID: "C1", Typ: "Mitarbeiter", Strasse: "Weg 4",
			PLZ: "10115", Ort: "Berlin", Land: "DEU", UStID: "DE111222333",
		},
		Notiz: "Tisch 3",
	}

	row := onlyRow(t, buildOne(t, closingWithBons(bon)), "transactions.csv")

	want := map[string]string{
		"BON_ID":        "40062",
		"BON_NR":        "62",
		"BON_TYP":       "Beleg",
		"BON_NAME":      "Kassenbeleg",
		"TERMINAL_ID":   "T1",
		"BON_STORNO":    "0",
		"BON_START":     "2019-01-21T15:45:05+01:00",
		"BON_ENDE":      "2019-01-21T15:45:55+01:00",
		"BEDIENER_ID":   "B1",
		"BEDIENER_NAME": "Anna",
		"UMS_BRUTTO":    "50,00",
		"KUNDE_NAME":    "Kunde A",
		"KUNDE_ID":      "C1",
		"KUNDE_TYP":     "Mitarbeiter",
		"KUNDE_STRASSE": "Weg 4",
		"KUNDE_PLZ":     "10115",
		"KUNDE_ORT":     "Berlin",
		"KUNDE_LAND":    "DEU",
		"KUNDE_USTID":   "DE111222333",
		"BON_NOTIZ":     "Tisch 3",
	}
	for column, w := range want {
		if got := field(t, "transactions.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildTransactionWritesZeroUmsatzAndStorno(t *testing.T) {
	t.Parallel()

	bon := model.Bon{ID: "1", Nr: 1, Typ: schema.BonTypBeleg, Storno: true}

	row := onlyRow(t, buildOne(t, closingWithBons(bon)), "transactions.csv")

	if got := field(t, "transactions.csv", row, "UMS_BRUTTO"); got != "0,00" {
		t.Errorf("UMS_BRUTTO = %q, want 0,00", got)
	}
	if got := field(t, "transactions.csv", row, "BON_STORNO"); got != "1" {
		t.Errorf("BON_STORNO = %q, want 1", got)
	}
}

func TestBuildAllocationGroupRows(t *testing.T) {
	t.Parallel()

	bon := model.Bon{ID: "1", Nr: 1, Typ: schema.BonTypBeleg, Abrechnungskreise: []string{"Tisch 10", "Kellner 3"}}

	rows := buildOne(t, closingWithBons(bon))["allocation_groups.csv"]
	if len(rows) != 2 {
		t.Fatalf("allocation_groups.csv has %d rows, want 2", len(rows))
	}
	if got := field(t, "allocation_groups.csv", rows[0], "ABRECHNUNGSKREIS"); got != "Tisch 10" {
		t.Errorf("first ABRECHNUNGSKREIS = %q, want Tisch 10", got)
	}
	if got := field(t, "allocation_groups.csv", rows[1], "BON_ID"); got != "1" {
		t.Errorf("second BON_ID = %q, want 1", got)
	}
}

func TestBuildPaymentRowsOfBon(t *testing.T) {
	t.Parallel()

	bon := model.Bon{
		ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
		Zahlungen: []model.Zahlung{
			{Typ: schema.ZahlartTypBar, BetragBasis: euro(1000)},
			{Typ: schema.ZahlartTypKreditkarte, Name: "Mastercard", Waehrung: "CHF", BetragWaehrung: euro(950), BetragBasis: euro(1000)},
		},
	}

	rows := buildOne(t, closingWithBons(bon))["datapayment.csv"]
	if len(rows) != 2 {
		t.Fatalf("datapayment.csv has %d rows, want 2", len(rows))
	}

	base := map[string]string{
		"BON_ID":           "1",
		"ZAHLART_TYP":      "Bar",
		"ZAHLART_NAME":     "",
		"ZAHLWAEH_CODE":    "EUR",
		"ZAHLWAEH_BETRAG":  "",
		"BASISWAEH_BETRAG": "10,00",
	}
	for column, w := range base {
		if got := field(t, "datapayment.csv", rows[0], column); got != w {
			t.Errorf("base %s = %q, want %q", column, got, w)
		}
	}

	foreign := map[string]string{
		"ZAHLART_TYP":      "Kreditkarte",
		"ZAHLART_NAME":     "Mastercard",
		"ZAHLWAEH_CODE":    "CHF",
		"ZAHLWAEH_BETRAG":  "9,50",
		"BASISWAEH_BETRAG": "10,00",
	}
	for column, w := range foreign {
		if got := field(t, "datapayment.csv", rows[1], column); got != w {
			t.Errorf("foreign %s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildTransactionVATRows(t *testing.T) {
	t.Parallel()

	bon := model.Bon{
		ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
		USt: []model.BonUSt{
			{Schluessel: 1, Brutto: euro(250), Netto: euro(210), USt: euro(40)},
			{Schluessel: 2, Brutto: dsfinvk.Zero, Netto: dsfinvk.Zero, USt: dsfinvk.Zero},
		},
	}

	rows := buildOne(t, closingWithBons(bon))["transactions_vat.csv"]
	if len(rows) != 2 {
		t.Fatalf("transactions_vat.csv has %d rows, want 2", len(rows))
	}

	first := map[string]string{"UST_SCHLUESSEL": "1", "BON_BRUTTO": "2,50", "BON_NETTO": "2,10", "BON_UST": "0,40"}
	for column, w := range first {
		if got := field(t, "transactions_vat.csv", rows[0], column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
	for _, column := range []string{"BON_BRUTTO", "BON_NETTO", "BON_UST"} {
		if got := field(t, "transactions_vat.csv", rows[1], column); got != "0,00" {
			t.Errorf("zero %s = %q, want 0,00", column, got)
		}
	}
}

func TestBuildReferenceRows(t *testing.T) {
	t.Parallel()

	refDate := time.Date(2019, time.January, 20, 12, 0, 0, 0, time.UTC)
	bon := model.Bon{
		ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
		Referenzen: []model.Referenz{
			{PosZeile: "2", Typ: schema.RefTypTransaktion, Datum: refDate, KasseID: "K0", Nr: 3, BonID: "900"},
			{Typ: schema.RefTypExterneSonstige, Name: "Gutschrift", Datum: refDate, KasseID: "K0", Nr: 3, BonID: "RE-4711"},
		},
	}

	rows := buildOne(t, closingWithBons(bon))["references.csv"]
	if len(rows) != 2 {
		t.Fatalf("references.csv has %d rows, want 2", len(rows))
	}

	internal := map[string]string{
		"POS_ZEILE":      "2",
		"REF_TYP":        "Transaktion",
		"REF_NAME":       "",
		"REF_DATUM":      "2019-01-20T12:00:00Z",
		"REF_Z_KASSE_ID": "K0",
		"REF_Z_NR":       "3",
		"REF_BON_ID":     "900",
	}
	for column, w := range internal {
		if got := field(t, "references.csv", rows[0], column); got != w {
			t.Errorf("internal %s = %q, want %q", column, got, w)
		}
	}

	external := map[string]string{
		"POS_ZEILE":      "",
		"REF_TYP":        "ExterneSonstige",
		"REF_NAME":       "Gutschrift",
		"REF_DATUM":      "",
		"REF_Z_KASSE_ID": "",
		"REF_Z_NR":       "",
		"REF_BON_ID":     "RE-4711",
	}
	for column, w := range external {
		if got := field(t, "references.csv", rows[1], column); got != w {
			t.Errorf("external %s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildTSETransactionRow(t *testing.T) {
	t.Parallel()

	c := closingWithBons(model.Bon{
		ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
		TSE: &model.TSETransaktion{
			TSEID: 1, Nr: 13,
			Start:       time.Date(2018, time.March, 21, 7, 30, 54, 250_000_000, time.FixedZone("CET", 3600)),
			Ende:        time.Date(2018, time.March, 21, 7, 30, 55, 0, time.FixedZone("CET", 3600)),
			Vorgangsart: schema.TSEVorgangsartKassenbelegV1,
			SigZaehler:  13, Signatur: "BEYwRAIg", Fehler: "", Vorgangsdaten: "Beleg^1.00_0.00",
		},
	})
	c.TSEs = []model.TSE{testTSE()}

	row := onlyRow(t, buildOne(t, c), "transactions_tse.csv")

	want := map[string]string{
		"BON_ID":             "1",
		"TSE_ID":             "1",
		"TSE_TANR":           "13",
		"TSE_TA_START":       "2018-03-21T06:30:54.250Z",
		"TSE_TA_ENDE":        "2018-03-21T06:30:55.000Z",
		"TSE_TA_VORGANGSART": "Kassenbeleg-V1",
		"TSE_TA_SIGZ":        "13",
		"TSE_TA_SIG":         "BEYwRAIg",
		"TSE_TA_FEHLER":      "",
		"TSE_VORGANGSDATEN":  "Beleg^1.00_0.00",
	}
	for column, w := range want {
		if got := field(t, "transactions_tse.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildRejectsDuplicateBonID(t *testing.T) {
	t.Parallel()

	c := closingWithBons(
		model.Bon{ID: "1", Nr: 1, Typ: schema.BonTypBeleg},
		model.Bon{ID: "1", Nr: 2, Typ: schema.BonTypBeleg},
	)

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrDuplicateBon) {
		t.Fatalf("Build error = %v, want ErrDuplicateBon", err)
	}
}

func TestBuildAllowsSameBonIDInAnotherClosing(t *testing.T) {
	t.Parallel()

	first := closingWithBons(model.Bon{ID: "1", Nr: 1, Typ: schema.BonTypBeleg})
	second := closingWithBons(model.Bon{ID: "1", Nr: 1, Typ: schema.BonTypBeleg})
	second.Nr = 2

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{first, second}}); err != nil {
		t.Fatalf("Build: %v", err)
	}
}

func TestBuildRejectsUnknownBonEnums(t *testing.T) {
	t.Parallel()

	cases := map[string]model.Bon{
		"BON_TYP":     {ID: "1", Nr: 1, Typ: "Kassenbeleg"},
		"ZAHLART_TYP": {ID: "1", Nr: 1, Typ: schema.BonTypBeleg, Zahlungen: []model.Zahlung{{Typ: "bar"}}},
		"REF_TYP":     {ID: "1", Nr: 1, Typ: schema.BonTypBeleg, Referenzen: []model.Referenz{{Typ: "Beleg", BonID: "9"}}},
	}
	for name, bon := range cases {
		c := closingWithBons(bon)
		if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrEnumValue) {
			t.Errorf("%s: Build error = %v, want ErrEnumValue", name, err)
		}
	}
}

func TestBuildRejectsUnknownTSEReference(t *testing.T) {
	t.Parallel()

	c := closingWithBons(model.Bon{
		ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
		TSE: &model.TSETransaktion{TSEID: 9, Nr: 1, Vorgangsart: schema.TSEVorgangsartKassenbelegV1},
	})
	c.TSEs = []model.TSE{testTSE()}

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrUnknownTSE) {
		t.Fatalf("Build error = %v, want ErrUnknownTSE", err)
	}
}

func TestBuildRejectsUnknownVATKeyReference(t *testing.T) {
	t.Parallel()

	c := closingWithBons(model.Bon{
		ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
		USt: []model.BonUSt{{Schluessel: 4}},
	})

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrUnknownUSt) {
		t.Fatalf("Build error = %v, want ErrUnknownUSt", err)
	}
}
