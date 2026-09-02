package schema_test

import (
	"slices"
	"testing"

	"github.com/tweecore/dsfinvk/schema"
)

func TestBonTypValues(t *testing.T) {
	want := []string{
		"Beleg", "AVRechnung", "AVTransfer", "AVBestellung", "AVTraining",
		"AVBelegstorno", "AVBelegabbruch", "AVSachbezug", "AVSonstige",
	}
	if got := schema.BonTypValues(); !slices.Equal(got, want) {
		t.Errorf("BonTypValues() =\n%v\nwant\n%v", got, want)
	}
	if got, want := len(schema.BonTypValues()), 9; got != want {
		t.Errorf("len = %d, want %d", got, want)
	}
	if got := string(schema.BonTypBeleg); got != "Beleg" {
		t.Errorf("BonTypBeleg = %q", got)
	}
	if got := string(schema.BonTypAVBelegstorno); got != "AVBelegstorno" {
		t.Errorf("BonTypAVBelegstorno = %q", got)
	}
}

func TestValidBonTyp(t *testing.T) {
	for _, v := range schema.BonTypValues() {
		if !schema.ValidBonTyp(v) {
			t.Errorf("ValidBonTyp(%q) = false", v)
		}
	}
	for _, v := range []string{"", "beleg", "BELEG", "Beleg ", "AVSonstiges", "Rechnung"} {
		if schema.ValidBonTyp(v) {
			t.Errorf("ValidBonTyp(%q) = true", v)
		}
	}
}

func TestGVTypValues(t *testing.T) {
	want := []string{
		"Umsatz", "Pfand", "PfandRueckzahlung", "Rabatt", "Aufschlag",
		"ZuschussEcht", "ZuschussUnecht", "TrinkgeldAG", "TrinkgeldAN",
		"EinzweckgutscheinKauf", "EinzweckgutscheinEinloesung",
		"MehrzweckgutscheinKauf", "MehrzweckgutscheinEinloesung",
		"Forderungsentstehung", "Forderungsaufloesung",
		"Anzahlungseinstellung", "Anzahlungsaufloesung",
		"Anfangsbestand", "Privatentnahme", "Privateinlage", "Geldtransit",
		"Lohnzahlung", "Einzahlung", "Auszahlung", "DifferenzSollIst",
	}
	got := schema.GVTypValues()
	if !slices.Equal(got, want) {
		t.Errorf("GVTypValues() =\n%v\nwant\n%v", got, want)
	}
	if len(got) != 25 {
		t.Errorf("len = %d, want 25", len(got))
	}
}

func TestValidGVTyp(t *testing.T) {
	for _, v := range schema.GVTypValues() {
		if !schema.ValidGVTyp(v) {
			t.Errorf("ValidGVTyp(%q) = false", v)
		}
	}
	for _, v := range []string{"", "umsatz", "PfandRückzahlung", "Trinkgeld", "Differenz"} {
		if schema.ValidGVTyp(v) {
			t.Errorf("ValidGVTyp(%q) = true", v)
		}
	}
}

func TestGVTypAffectsCashOnly(t *testing.T) {
	want := []string{
		"Anfangsbestand", "Privatentnahme", "Privateinlage", "Geldtransit",
		"Lohnzahlung", "Einzahlung", "Auszahlung", "DifferenzSollIst",
	}
	wantSet := make(map[string]bool, len(want))
	for _, v := range want {
		wantSet[v] = true
	}
	n := 0
	for _, v := range schema.GVTypValues() {
		got := schema.GVTyp(v).AffectsCashOnly()
		if got != wantSet[v] {
			t.Errorf("GVTyp(%q).AffectsCashOnly() = %v, want %v", v, got, wantSet[v])
		}
		if got {
			n++
		}
	}
	if n != 8 {
		t.Errorf("%d cash-only types, want 8", n)
	}
	if schema.GVTyp("Nonsense").AffectsCashOnly() {
		t.Error(`GVTyp("Nonsense").AffectsCashOnly() = true`)
	}
}

func TestZahlartTypValues(t *testing.T) {
	want := []string{"Bar", "Unbar", "Keine", "ECKarte", "Kreditkarte", "ElZahlungsdienstleister", "Guthabenkarte"}
	if got := schema.ZahlartTypValues(); !slices.Equal(got, want) {
		t.Errorf("ZahlartTypValues() =\n%v\nwant\n%v", got, want)
	}
	for _, v := range want {
		if !schema.ValidZahlartTyp(v) {
			t.Errorf("ValidZahlartTyp(%q) = false", v)
		}
	}
	for _, v := range []string{"", "bar", "EC-Karte", "Karte"} {
		if schema.ValidZahlartTyp(v) {
			t.Errorf("ValidZahlartTyp(%q) = true", v)
		}
	}
}

func TestRefTypValues(t *testing.T) {
	want := []string{"ExterneRechnung", "ExternerLieferschein", "Transaktion", "ExterneSonstige"}
	if got := schema.RefTypValues(); !slices.Equal(got, want) {
		t.Errorf("RefTypValues() =\n%v\nwant\n%v", got, want)
	}
	for _, v := range want {
		if !schema.ValidRefTyp(v) {
			t.Errorf("ValidRefTyp(%q) = false", v)
		}
	}
	for _, v := range []string{"", "transaktion", "ExternerRechnung"} {
		if schema.ValidRefTyp(v) {
			t.Errorf("ValidRefTyp(%q) = true", v)
		}
	}
}

func TestPreisfindungTyp(t *testing.T) {
	want := []string{
		"base_amount", "Grundpreis",
		"discount", "Rabatt",
		"extra_amount", "Zuschlag",
	}
	if got := schema.PreisfindungTypValues(); !slices.Equal(got, want) {
		t.Errorf("PreisfindungTypValues() =\n%v\nwant\n%v", got, want)
	}
	wantCanonical := []string{"base_amount", "discount", "extra_amount"}
	if got := schema.PreisfindungTypCanonicalValues(); !slices.Equal(got, wantCanonical) {
		t.Errorf("PreisfindungTypCanonicalValues() = %v, want %v", got, wantCanonical)
	}
	for _, v := range want {
		if !schema.ValidPreisfindungTyp(v) {
			t.Errorf("ValidPreisfindungTyp(%q) = false", v)
		}
	}
	for _, v := range []string{"", "grundpreis", "BASE_AMOUNT", "Discount", "surcharge"} {
		if schema.ValidPreisfindungTyp(v) {
			t.Errorf("ValidPreisfindungTyp(%q) = true", v)
		}
	}
}

func TestNormalizePreisfindungTyp(t *testing.T) {
	tests := map[string]schema.PreisfindungTyp{
		"base_amount":  schema.PreisfindungBaseAmount,
		"Grundpreis":   schema.PreisfindungBaseAmount,
		"discount":     schema.PreisfindungDiscount,
		"Rabatt":       schema.PreisfindungDiscount,
		"extra_amount": schema.PreisfindungExtraAmount,
		"Zuschlag":     schema.PreisfindungExtraAmount,
	}
	for in, want := range tests {
		got, ok := schema.NormalizePreisfindungTyp(in)
		if !ok {
			t.Errorf("NormalizePreisfindungTyp(%q): ok = false", in)
			continue
		}
		if got != want {
			t.Errorf("NormalizePreisfindungTyp(%q) = %q, want %q", in, got, want)
		}
	}
	if _, ok := schema.NormalizePreisfindungTyp("nope"); ok {
		t.Error(`NormalizePreisfindungTyp("nope"): ok = true`)
	}
	for _, v := range schema.PreisfindungTypCanonicalValues() {
		got, ok := schema.NormalizePreisfindungTyp(v)
		if !ok || string(got) != v {
			t.Errorf("NormalizePreisfindungTyp(%q) = %q, %v", v, got, ok)
		}
	}
}

func TestTSESigAlgoValues(t *testing.T) {
	want := []string{
		"ecdsa-plain-SHA224", "ecdsa-plain-SHA256", "ecdsa-plain-SHA384", "ecdsa-plain-SHA512",
		"ecdsa-plain-SHA3-224", "ecdsa-plain-SHA3-256", "ecdsa-plain-SHA3-384", "ecdsa-plain-SHA3-512",
		"ecsdsa-plain-SHA224", "ecsdsa-plain-SHA256", "ecsdsa-plain-SHA384", "ecsdsa-plain-SHA512",
		"ecsdsa-plain-SHA3-224", "ecsdsa-plain-SHA3-256", "ecsdsa-plain-SHA3-384", "ecsdsa-plain-SHA3-512",
	}
	got := schema.TSESigAlgoValues()
	if !slices.Equal(got, want) {
		t.Errorf("TSESigAlgoValues() =\n%v\nwant\n%v", got, want)
	}
	if len(got) != 16 {
		t.Errorf("len = %d, want 16", len(got))
	}
	// "ecsdsa" is the spec's own spelling and must not be "corrected".
	if !schema.ValidTSESigAlgo("ecsdsa-plain-SHA256") {
		t.Error(`ValidTSESigAlgo("ecsdsa-plain-SHA256") = false`)
	}
	tbl, ok := schema.TableByFile("tse.csv")
	if !ok {
		t.Fatal("tse.csv not found")
	}
	col, ok := tbl.Column("TSE_SIG_ALGO")
	if !ok {
		t.Fatal("TSE_SIG_ALGO not found")
	}
	for _, v := range got {
		if len(v) > col.MaxLength {
			t.Errorf("%q is %d characters, TSE_SIG_ALGO allows %d", v, len(v), col.MaxLength)
		}
	}
}

func TestValidTSESigAlgo(t *testing.T) {
	for _, v := range schema.TSESigAlgoValues() {
		if !schema.ValidTSESigAlgo(v) {
			t.Errorf("ValidTSESigAlgo(%q) = false", v)
		}
	}
	for _, v := range []string{"", "ECDSA-plain-SHA256", "ecdsa-plain-sha256", "ecdsa-plain-SHA1", "eddsa-plain-SHA256"} {
		if schema.ValidTSESigAlgo(v) {
			t.Errorf("ValidTSESigAlgo(%q) = true", v)
		}
	}
}

func TestTSEZeitformatValues(t *testing.T) {
	want := []string{"unixTime", "utcTime", "utcTimeWithSeconds", "generalizedTime", "generalizedTimeWithMilliseconds"}
	if got := schema.TSEZeitformatValues(); !slices.Equal(got, want) {
		t.Errorf("TSEZeitformatValues() =\n%v\nwant\n%v", got, want)
	}
	for _, v := range want {
		if !schema.ValidTSEZeitformat(v) {
			t.Errorf("ValidTSEZeitformat(%q) = false", v)
		}
	}
	for _, v := range []string{"", "unixtime", "UnixTime", "utcTimeWithMilliseconds"} {
		if schema.ValidTSEZeitformat(v) {
			t.Errorf("ValidTSEZeitformat(%q) = true", v)
		}
	}
}

func TestTSEPDEncodingValues(t *testing.T) {
	want := []string{"UTF-8", "ASCII"}
	if got := schema.TSEPDEncodingValues(); !slices.Equal(got, want) {
		t.Errorf("TSEPDEncodingValues() = %v, want %v", got, want)
	}
	for _, v := range want {
		if !schema.ValidTSEPDEncoding(v) {
			t.Errorf("ValidTSEPDEncoding(%q) = false", v)
		}
	}
	for _, v := range []string{"", "utf-8", "UTF8", "ascii", "Latin-1"} {
		if schema.ValidTSEPDEncoding(v) {
			t.Errorf("ValidTSEPDEncoding(%q) = true", v)
		}
	}
}

func TestTSEVorgangsartValues(t *testing.T) {
	want := []string{"Kassenbeleg-V1", "Bestellung-V1", "SonstigerVorgang"}
	if got := schema.TSEVorgangsartValues(); !slices.Equal(got, want) {
		t.Errorf("TSEVorgangsartValues() = %v, want %v", got, want)
	}
	for _, v := range want {
		if !schema.ValidTSEVorgangsart(v) {
			t.Errorf("ValidTSEVorgangsart(%q) = false", v)
		}
	}
	for _, v := range []string{"", "Kassenbeleg", "kassenbeleg-v1", "Bestellung", "SonstigeVorgang"} {
		if schema.ValidTSEVorgangsart(v) {
			t.Errorf("ValidTSEVorgangsart(%q) = true", v)
		}
	}
}

func TestEnumValuesReturnCopies(t *testing.T) {
	tests := map[string]func() []string{
		"BonTyp":          schema.BonTypValues,
		"GVTyp":           schema.GVTypValues,
		"ZahlartTyp":      schema.ZahlartTypValues,
		"RefTyp":          schema.RefTypValues,
		"PreisfindungTyp": schema.PreisfindungTypValues,
		"TSESigAlgo":      schema.TSESigAlgoValues,
		"TSEZeitformat":   schema.TSEZeitformatValues,
		"TSEPDEncoding":   schema.TSEPDEncodingValues,
		"TSEVorgangsart":  schema.TSEVorgangsartValues,
	}
	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			first := fn()
			orig := first[0]
			first[0] = "tampered"
			if again := fn(); again[0] != orig {
				t.Errorf("%sValues() shares its backing array: got %q", name, again[0])
			}
		})
	}
}

func TestEnumValuesFitTheirColumns(t *testing.T) {
	tests := []struct {
		file, column string
		values       []string
	}{
		{"transactions.csv", "BON_TYP", schema.BonTypValues()},
		{"businesscases.csv", "GV_TYP", schema.GVTypValues()},
		{"payment.csv", "ZAHLART_TYP", schema.ZahlartTypValues()},
		{"references.csv", "REF_TYP", schema.RefTypValues()},
		{"itemamounts.csv", "TYP", schema.PreisfindungTypValues()},
		{"tse.csv", "TSE_SIG_ALGO", schema.TSESigAlgoValues()},
		{"tse.csv", "TSE_ZEITFORMAT", schema.TSEZeitformatValues()},
		{"tse.csv", "TSE_PD_ENCODING", schema.TSEPDEncodingValues()},
		{"transactions_tse.csv", "TSE_TA_VORGANGSART", schema.TSEVorgangsartValues()},
	}
	for _, tc := range tests {
		t.Run(tc.file+"/"+tc.column, func(t *testing.T) {
			tbl, ok := schema.TableByFile(tc.file)
			if !ok {
				t.Fatalf("table %q not found", tc.file)
			}
			col, ok := tbl.Column(tc.column)
			if !ok {
				t.Fatalf("column %q not found", tc.column)
			}
			if col.Type != schema.ColumnAlphaNumeric {
				t.Fatalf("column %q is %v, expected AlphaNumeric", tc.column, col.Type)
			}
			for _, v := range tc.values {
				if len([]rune(v)) > col.MaxLength {
					t.Errorf("%q is %d characters, %s allows %d", v, len([]rune(v)), tc.column, col.MaxLength)
				}
			}
		})
	}
}

// TestBool01 pins Anhang G legend K228: "Boolean-Felder werden mit 0 oder 1
// exportiert".
func TestBool01(t *testing.T) {
	if schema.Bool01False != "0" || schema.Bool01True != "1" {
		t.Errorf("Bool01False/Bool01True = %q/%q, want \"0\"/\"1\"", schema.Bool01False, schema.Bool01True)
	}
	for _, s := range []string{"0", "1"} {
		if !schema.ValidBool01(s) {
			t.Errorf("ValidBool01(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", " ", "2", "-1", "01", "true", "TRUE", "ja", "00"} {
		if schema.ValidBool01(s) {
			t.Errorf("ValidBool01(%q) = true, want false", s)
		}
	}
}

// TestIsBool01Column checks the boolean columns against Anhang E, where exactly
// four fields carry "Feldlänge: 1". Spec 2.4 p.72, p.85, p.94.
func TestIsBool01Column(t *testing.T) {
	want := map[string]bool{
		"cashregister.csv/KEINE_UST_ZUORDNUNG": true,
		"transactions.csv/BON_STORNO":          true,
		"lines.csv/INHAUS":                     true,
		"lines.csv/P_STORNO":                   true,
	}
	for _, tbl := range schema.Tables() {
		for _, c := range tbl.Columns {
			key := tbl.File + "/" + c.Name
			if got := schema.IsBool01Column(tbl.File, c.Name); got != want[key] {
				t.Errorf("IsBool01Column(%q, %q) = %t, want %t", tbl.File, c.Name, got, want[key])
			}
			// index.xml corroborates Anhang E: a boolean column is exactly a
			// one-character AlphaNumeric column.
			single := c.Type == schema.ColumnAlphaNumeric && c.MaxLength == 1
			if single != want[key] {
				t.Errorf("%s: MaxLength 1 = %t but boolean = %t", key, single, want[key])
			}
		}
	}
	if schema.IsBool01Column("nope.csv", "INHAUS") {
		t.Error(`IsBool01Column("nope.csv", "INHAUS") = true, want false`)
	}
	if schema.IsBool01Column("lines.csv", "NOPE") {
		t.Error(`IsBool01Column("lines.csv", "NOPE") = true, want false`)
	}
}

// TestBonTypAllowedInProcessData pins spec 2.4 p.113: AVBelegstorno cannot be
// used by systems secured with a TSE.
func TestBonTypAllowedInProcessData(t *testing.T) {
	for _, v := range schema.BonTypValues() {
		want := v != string(schema.BonTypAVBelegstorno)
		if got := schema.BonTyp(v).AllowedInProcessData(); got != want {
			t.Errorf("BonTyp(%q).AllowedInProcessData() = %t, want %t", v, got, want)
		}
	}
	n := 0
	for _, v := range schema.BonTypValues() {
		if schema.BonTyp(v).AllowedInProcessData() {
			n++
		}
	}
	if n != 8 {
		t.Errorf("%d types allowed in processData, want 8", n)
	}
}

// TestProcessDataClass pins spec 2.4 p.114 and Anhang D p.61 to p.63: processData
// knows only "Bar" and "Unbar".
func TestProcessDataClass(t *testing.T) {
	tests := map[schema.ZahlartTyp]schema.ProcessDataPaymentClass{
		schema.ZahlartTypBar:                     schema.PaymentClassBar,
		schema.ZahlartTypUnbar:                   schema.PaymentClassUnbar,
		schema.ZahlartTypECKarte:                 schema.PaymentClassUnbar,
		schema.ZahlartTypKreditkarte:             schema.PaymentClassUnbar,
		schema.ZahlartTypElZahlungsdienstleister: schema.PaymentClassUnbar,
		schema.ZahlartTypGuthabenkarte:           schema.PaymentClassUnbar,
	}
	for typ, want := range tests {
		got, ok := typ.ProcessDataClass()
		if !ok {
			t.Errorf("%q.ProcessDataClass(): ok = false", typ)
			continue
		}
		if got != want {
			t.Errorf("%q.ProcessDataClass() = %v, want %v", typ, got, want)
		}
	}
	for _, typ := range []schema.ZahlartTyp{schema.ZahlartTypKeine, "", "Bitcoin", "bar"} {
		if class, ok := typ.ProcessDataClass(); ok {
			t.Errorf("%q.ProcessDataClass() = (%v, true), want ok = false", typ, class)
		}
	}
	for _, v := range schema.ZahlartTypValues() {
		if _, ok := schema.ZahlartTyp(v).ProcessDataClass(); !ok && v != string(schema.ZahlartTypKeine) {
			t.Errorf("%q has no processData class", v)
		}
	}
}

func TestProcessDataPaymentClassString(t *testing.T) {
	tests := map[schema.ProcessDataPaymentClass]string{
		schema.PaymentClassBar:            "Bar",
		schema.PaymentClassUnbar:          "Unbar",
		schema.ProcessDataPaymentClass(9): "ProcessDataPaymentClass(9)",
	}
	for c, want := range tests {
		if got := c.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", uint8(c), got, want)
		}
	}
}

// TestValidISOAlpha3 covers the shape of the currency columns (spec 2.4 p.72,
// p.83, p.92) and the country columns (p.68, p.70, p.75, p.89).
func TestValidISOAlpha3(t *testing.T) {
	for _, s := range []string{"EUR", "CHF", "USD", "DEU", "AUT", "ZZZ"} {
		if !schema.ValidISOAlpha3(s) {
			t.Errorf("ValidISOAlpha3(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "E", "EU", "EURO", "eur", "Eur", "EU1", "E R", "DÄN", "€UR"} {
		if schema.ValidISOAlpha3(s) {
			t.Errorf("ValidISOAlpha3(%q) = true, want false", s)
		}
	}
	for _, tc := range [][2]string{
		{"cashregister.csv", "KASSE_BASISWAEH_CODE"},
		{"cash_per_currency.csv", "ZAHLART_WAEH"},
		{"datapayment.csv", "ZAHLWAEH_CODE"},
		{"cashpointclosing.csv", "LAND"},
		{"location.csv", "LOC_LAND"},
		{"pa.csv", "AGENTUR_LAND"},
		{"transactions.csv", "KUNDE_LAND"},
	} {
		tbl, ok := schema.TableByFile(tc[0])
		if !ok {
			t.Fatalf("TableByFile(%q) not found", tc[0])
		}
		c, ok := tbl.Column(tc[1])
		if !ok {
			t.Fatalf("%s: column %q not found", tc[0], tc[1])
		}
		if c.MaxLength != 3 {
			t.Errorf("%s.%s: MaxLength = %d, want 3", tc[0], tc[1], c.MaxLength)
		}
	}
}
