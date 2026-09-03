package model_test

import (
	"errors"
	"testing"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

// quantity returns a Decimal from thousandths.
func quantity(milli int64) dsfinvk.Decimal {
	v, err := dsfinvk.FromScaled(milli, 3)
	if err != nil {
		panic(err)
	}
	return v
}

// bonWith wraps positions in one Beleg.
func bonWith(positions ...model.Position) model.Bon {
	return model.Bon{ID: "1", Nr: 1, Typ: schema.BonTypBeleg, Positionen: positions}
}

func TestBuildLineRow(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{
		Zeile: "1", GutscheinNr: "4666", Artikeltext: "Fleisch", TerminalID: "T1",
		GVTyp: schema.GVTypUmsatz, GVName: "Theke", Inhaus: true, Storno: false,
		AgenturID: 0, ArtNr: "666431", GTIN: "4001234567890", WarengrID: "W1", Warengr: "Frisch",
		Menge: quantity(2000), Faktor: quantity(1500), Einheit: "kg", StkBrutto: euro(500),
	}))

	row := onlyRow(t, buildOne(t, c), "lines.csv")

	want := map[string]string{
		"BON_ID":          "1",
		"POS_ZEILE":       "1",
		"GUTSCHEIN_NR":    "4666",
		"ARTIKELTEXT":     "Fleisch",
		"POS_TERMINAL_ID": "T1",
		"GV_TYP":          "Umsatz",
		"GV_NAME":         "Theke",
		"INHAUS":          "1",
		"P_STORNO":        "0",
		"AGENTUR_ID":      "0",
		"ART_NR":          "666431",
		"GTIN":            "4001234567890",
		"WARENGR_ID":      "W1",
		"WARENGR":         "Frisch",
		"MENGE":           "2,000",
		"FAKTOR":          "1,500",
		"EINHEIT":         "kg",
		"STK_BR":          "5,00",
	}
	for column, w := range want {
		if got := field(t, "lines.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildLineLeavesOptionalNumbersEmpty(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz}))

	row := onlyRow(t, buildOne(t, c), "lines.csv")

	for _, column := range []string{"MENGE", "FAKTOR", "STK_BR"} {
		if got := field(t, "lines.csv", row, column); got != "" {
			t.Errorf("%s = %q, want empty", column, got)
		}
	}
	if got := field(t, "lines.csv", row, "AGENTUR_ID"); got != "0" {
		t.Errorf("AGENTUR_ID = %q, want 0", got)
	}
}

func TestBuildLineVATRows(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{
		Zeile: "1", GVTyp: schema.GVTypUmsatz,
		USt: []model.PosUSt{
			{Schluessel: 1, Brutto: euro(250), Netto: euro(210), USt: euro(40)},
			{Schluessel: 2, Brutto: euro(200)},
		},
	}))

	rows := buildOne(t, c)["lines_vat.csv"]
	if len(rows) != 2 {
		t.Fatalf("lines_vat.csv has %d rows, want 2", len(rows))
	}

	first := map[string]string{
		"BON_ID": "1", "POS_ZEILE": "1", "UST_SCHLUESSEL": "1",
		"POS_BRUTTO": "2,50", "POS_NETTO": "2,10", "POS_UST": "0,40",
	}
	for column, w := range first {
		if got := field(t, "lines_vat.csv", rows[0], column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}

	second := map[string]string{"UST_SCHLUESSEL": "2", "POS_BRUTTO": "2,00", "POS_NETTO": "", "POS_UST": ""}
	for column, w := range second {
		if got := field(t, "lines_vat.csv", rows[1], column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildItemamountRows(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{
		Zeile: "1", GVTyp: schema.GVTypUmsatz,
		Preisfindung: []model.Preisfindung{
			{Typ: schema.PreisfindungBaseAmount, Schluessel: 1, Brutto: euro(5000)},
			{Typ: schema.PreisfindungRabatt, Schluessel: 1, Brutto: euro(-500), Netto: euro(-420), USt: euro(-80)},
		},
	}))

	rows := buildOne(t, c)["itemamounts.csv"]
	if len(rows) != 2 {
		t.Fatalf("itemamounts.csv has %d rows, want 2", len(rows))
	}

	first := map[string]string{
		"BON_ID": "1", "POS_ZEILE": "1", "TYP": "base_amount", "UST_SCHLUESSEL": "1",
		"PF_BRUTTO": "50,00", "PF_NETTO": "", "PF_UST": "",
	}
	for column, w := range first {
		if got := field(t, "itemamounts.csv", rows[0], column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}

	second := map[string]string{"TYP": "Rabatt", "PF_BRUTTO": "-5,00", "PF_NETTO": "-4,20", "PF_UST": "-0,80"}
	for column, w := range second {
		if got := field(t, "itemamounts.csv", rows[1], column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildSubitemRow(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{
		Zeile: "1", GVTyp: schema.GVTypUmsatz,
		Zusatzinfos: []model.Zusatzinfo{{
			ArtNr: "A1", GTIN: "4009", Name: "Cola", WarengrID: "W2", Warengr: "Getraenke",
			Menge: quantity(1000), Faktor: quantity(1000), Einheit: "Stueck", Schluessel: 1,
			BasispreisBrutto: euro(250), BasispreisNetto: euro(210), BasispreisUSt: euro(40),
		}},
	}))

	row := onlyRow(t, buildOne(t, c), "subitems.csv")

	want := map[string]string{
		"BON_ID":               "1",
		"POS_ZEILE":            "1",
		"ZI_ART_NR":            "A1",
		"ZI_GTIN":              "4009",
		"ZI_NAME":              "Cola",
		"ZI_WARENGR_ID":        "W2",
		"ZI_WARENGR":           "Getraenke",
		"ZI_MENGE":             "1,000",
		"ZI_FAKTOR":            "1,000",
		"ZI_EINHEIT":           "Stueck",
		"ZI_UST_SCHLUESSEL":    "1",
		"ZI_BASISPREIS_BRUTTO": "2,50",
		"ZI_BASISPREIS_NETTO":  "2,10",
		"ZI_BASISPREIS_UST":    "0,40",
	}
	for column, w := range want {
		if got := field(t, "subitems.csv", row, column); got != w {
			t.Errorf("%s = %q, want %q", column, got, w)
		}
	}
}

func TestBuildSubitemWritesZeroMenge(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{
		Zeile: "1", GVTyp: schema.GVTypUmsatz,
		Zusatzinfos: []model.Zusatzinfo{{ArtNr: "A1", Schluessel: 1}},
	}))

	row := onlyRow(t, buildOne(t, c), "subitems.csv")

	if got := field(t, "subitems.csv", row, "ZI_MENGE"); got != "0,000" {
		t.Errorf("ZI_MENGE = %q, want 0,000", got)
	}
	if got := field(t, "subitems.csv", row, "ZI_FAKTOR"); got != "" {
		t.Errorf("ZI_FAKTOR = %q, want empty", got)
	}
}

func TestBuildRejectsDuplicatePositionZeile(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(
		model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz},
		model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz},
	))

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrDuplicatePosition) {
		t.Fatalf("Build error = %v, want ErrDuplicatePosition", err)
	}
}

func TestBuildRejectsUnknownAgentur(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz, AgenturID: 3}))

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrUnknownAgentur) {
		t.Fatalf("Build error = %v, want ErrUnknownAgentur", err)
	}
}

func TestBuildAcceptsDeclaredAgentur(t *testing.T) {
	t.Parallel()

	c := closingWithBons(bonWith(model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz, AgenturID: 3}))
	c.Agenturen = []model.Agentur{{ID: 3, Address: model.Address{Name: "Lotto AG"}}}

	row := onlyRow(t, buildOne(t, c), "lines.csv")
	if got := field(t, "lines.csv", row, "AGENTUR_ID"); got != "3" {
		t.Errorf("AGENTUR_ID = %q, want 3", got)
	}
}

func TestBuildRejectsUnknownPositionEnums(t *testing.T) {
	t.Parallel()

	cases := map[string]model.Position{
		"GV_TYP": {Zeile: "1", GVTyp: "Verkauf"},
		"TYP": {
			Zeile: "1", GVTyp: schema.GVTypUmsatz,
			Preisfindung: []model.Preisfindung{{Typ: "Grundbetrag", Schluessel: 1}},
		},
	}
	for name, p := range cases {
		c := closingWithBons(bonWith(p))
		if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrEnumValue) {
			t.Errorf("%s: Build error = %v, want ErrEnumValue", name, err)
		}
	}
}
