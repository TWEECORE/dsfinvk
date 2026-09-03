package model_test

import (
	"reflect"
	"testing"

	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

// project reduces the rows of a file to the named columns, in row order.
func project(t *testing.T, rows model.Rows, file string, columns ...string) [][]string {
	t.Helper()

	out := make([][]string, 0, len(rows[file]))
	for _, row := range rows[file] {
		picked := make([]string, len(columns))
		for i, column := range columns {
			picked[i] = field(t, file, row, column)
		}
		out = append(out, picked)
	}
	return out
}

// gvColumns and payColumns are the aggregate columns the oracles pin.
var (
	gvColumns  = []string{"GV_TYP", "GV_NAME", "AGENTUR_ID", "UST_SCHLUESSEL", "Z_UMS_BRUTTO", "Z_UMS_NETTO", "Z_UST"}
	payColumns = []string{"ZAHLART_TYP", "ZAHLART_NAME", "Z_ZAHLART_BETRAG"}
	curColumns = []string{"ZAHLART_WAEH", "ZAHLART_BETRAG_WAEH"}
)

// beleg builds a Beleg with printed totals, payments and positions.
func beleg(id string, nr, ums int64, ust []model.BonUSt, zahlungen []model.Zahlung, positionen ...model.Position) model.Bon {
	return model.Bon{
		ID: id, Nr: nr, Typ: schema.BonTypBeleg, UmsBrutto: euro(ums),
		USt: ust, Zahlungen: zahlungen, Positionen: positionen,
	}
}

// pos builds a Position with one VAT share.
func pos(zeile string, typ schema.GVTyp, schluessel, brutto, netto, ust int64) model.Position {
	return model.Position{
		Zeile: zeile, GVTyp: typ,
		USt: []model.PosUSt{{Schluessel: schluessel, Brutto: euro(brutto), Netto: euro(netto), USt: euro(ust)}},
	}
}

// bar is a cash payment in the Basiswaehrung.
func bar(cents int64) []model.Zahlung {
	return []model.Zahlung{{Typ: schema.ZahlartTypBar, BetragBasis: euro(cents)}}
}

// keine is the placeholder payment of a Vorgang that moves no money.
func keine() []model.Zahlung {
	return []model.Zahlung{{Typ: schema.ZahlartTypKeine}}
}

// anhangF is the worked example of Spec 2.4 Anhang F p.106 to p.108.
func anhangF() model.Kassenabschluss {
	return closingWithBons(
		beleg("1", 1, 25000, []model.BonUSt{{Schluessel: 5, Brutto: euro(25000), Netto: euro(25000)}}, bar(25000),
			pos("1", schema.GVTypGeldtransit, 5, 25000, 25000, 0)),
		beleg("2", 2, 450, []model.BonUSt{
			{Schluessel: 1, Brutto: euro(250), Netto: euro(210), USt: euro(40)},
			{Schluessel: 2, Brutto: euro(200), Netto: euro(187), USt: euro(13)},
		}, bar(450),
			pos("1", schema.GVTypUmsatz, 1, 250, 210, 40),
			pos("2", schema.GVTypUmsatz, 2, 200, 187, 13)),
		beleg("3", 3, 450, []model.BonUSt{
			{Schluessel: 1, Brutto: euro(250), Netto: euro(210), USt: euro(40)},
			{Schluessel: 2, Brutto: euro(200), Netto: euro(187), USt: euro(13)},
		}, []model.Zahlung{{Typ: schema.ZahlartTypKreditkarte, Name: "Mastercard", BetragBasis: euro(450)}},
			pos("1", schema.GVTypUmsatz, 1, 250, 210, 40),
			pos("2", schema.GVTypUmsatz, 2, 200, 187, 13)),
		beleg("4", 4, -500, []model.BonUSt{{Schluessel: 5, Brutto: euro(-500), Netto: euro(-500)}}, bar(-500),
			pos("1", schema.GVTypPrivatentnahme, 5, -500, -500, 0)),
		model.Bon{
			ID: "5", Nr: 5, Typ: schema.BonTypAVBestellung, UmsBrutto: euro(1250),
			USt:        []model.BonUSt{{Schluessel: 2, Brutto: euro(1250), Netto: euro(1168), USt: euro(82)}},
			Zahlungen:  keine(),
			Positionen: []model.Position{pos("1", schema.GVTypUmsatz, 2, 1250, 1168, 82)},
		},
		beleg("6", 6, 250, []model.BonUSt{{Schluessel: 2, Brutto: euro(250), Netto: euro(234), USt: euro(16)}}, bar(250),
			pos("1", schema.GVTypAnzahlungseinstellung, 2, 250, 234, 16)),
		beleg("7", 7, 0, []model.BonUSt{{Schluessel: 2}}, keine(),
			pos("1", schema.GVTypUmsatz, 2, 250, 234, 16),
			pos("2", schema.GVTypAnzahlungsaufloesung, 2, -250, -234, -16)),
		model.Bon{
			ID: "8a", Nr: 8, Typ: schema.BonTypAVBestellung, UmsBrutto: euro(250),
			USt:        []model.BonUSt{{Schluessel: 2, Brutto: euro(250), Netto: euro(234), USt: euro(16)}},
			Zahlungen:  keine(),
			Positionen: []model.Position{pos("1", schema.GVTypUmsatz, 2, 250, 234, 16)},
		},
		beleg("8b", 9, 100, []model.BonUSt{{Schluessel: 2, Brutto: euro(100), Netto: euro(94), USt: euro(6)}}, bar(100),
			pos("1", schema.GVTypAnzahlungseinstellung, 2, 100, 94, 6)),
		beleg("9", 10, 150, []model.BonUSt{{Schluessel: 2, Brutto: euro(150), Netto: euro(140), USt: euro(10)}}, bar(150),
			pos("1", schema.GVTypUmsatz, 2, 250, 234, 16),
			pos("2", schema.GVTypAnzahlungsaufloesung, 2, -100, -94, -6)),
		beleg("10", 11, 100, []model.BonUSt{{Schluessel: 2, Brutto: euro(250), Netto: euro(234), USt: euro(16)}}, bar(100),
			pos("1", schema.GVTypUmsatz, 2, 250, 234, 16),
			pos("2", schema.GVTypForderungsentstehung, 2, -150, -140, -10)),
		beleg("11", 12, 150, []model.BonUSt{{Schluessel: 2}}, bar(150),
			pos("1", schema.GVTypForderungsaufloesung, 2, 150, 140, 10)),
	)
}

func TestAggregateAnhangFBusinesscases(t *testing.T) {
	t.Parallel()

	rows := buildOne(t, anhangF())

	// Spec 2.4 p.108: Umsatz 5,00 and 11,50, Geldtransit 250,00, Privatentnahme -5,00.
	// The netto and USt columns are summed from the Positionen, so key 2 gives
	// 10,76 and 0,74 where the illustration divides the total and prints 10,75 and 0,75.
	want := [][]string{
		{"Geldtransit", "", "0", "5", "250,00", "250,00", "0,00"},
		{"Umsatz", "", "0", "1", "5,00", "4,20", "0,80"},
		{"Umsatz", "", "0", "2", "11,50", "10,76", "0,74"},
		{"Privatentnahme", "", "0", "5", "-5,00", "-5,00", "0,00"},
		{"Anzahlungseinstellung", "", "0", "2", "3,50", "3,28", "0,22"},
		{"Anzahlungsaufloesung", "", "0", "2", "-3,50", "-3,28", "-0,22"},
		{"Forderungsentstehung", "", "0", "2", "-1,50", "-1,40", "-0,10"},
		{"Forderungsaufloesung", "", "0", "2", "1,50", "1,40", "0,10"},
	}
	if got := project(t, rows, "businesscases.csv", gvColumns...); !reflect.DeepEqual(got, want) {
		t.Errorf("businesscases.csv =\n%v\nwant\n%v", got, want)
	}
}

func TestAggregateAnhangFPayments(t *testing.T) {
	t.Parallel()

	rows := buildOne(t, anhangF())

	// Spec 2.4 p.108: Bar 257,00 and Kreditkarte Mastercard 4,50. The
	// illustration omits the Keine row of the Vorgang that moves no money.
	wantPay := [][]string{
		{"Bar", "", "257,00"},
		{"Kreditkarte", "Mastercard", "4,50"},
		{"Keine", "", "0,00"},
	}
	if got := project(t, rows, "payment.csv", payColumns...); !reflect.DeepEqual(got, wantPay) {
		t.Errorf("payment.csv =\n%v\nwant\n%v", got, wantPay)
	}

	// Spec 2.4 p.108: the cash balance is EUR 257,00.
	wantCur := [][]string{{"EUR", "257,00"}}
	if got := project(t, rows, "cash_per_currency.csv", curColumns...); !reflect.DeepEqual(got, wantCur) {
		t.Errorf("cash_per_currency.csv =\n%v\nwant\n%v", got, wantCur)
	}

	row := onlyRow(t, rows, "cashpointclosing.csv")
	if got := field(t, "cashpointclosing.csv", row, "Z_SE_ZAHLUNGEN"); got != "261,50" {
		t.Errorf("Z_SE_ZAHLUNGEN = %q, want 261,50", got)
	}
	if got := field(t, "cashpointclosing.csv", row, "Z_SE_BARZAHLUNGEN"); got != "257,00" {
		t.Errorf("Z_SE_BARZAHLUNGEN = %q, want 257,00", got)
	}
	if got := field(t, "cashpointclosing.csv", row, "Z_START_ID"); got != "1" {
		t.Errorf("Z_START_ID = %q, want 1", got)
	}
	if got := field(t, "cashpointclosing.csv", row, "Z_ENDE_ID"); got != "11" {
		t.Errorf("Z_ENDE_ID = %q, want 11", got)
	}
}

func TestAggregateEinzweckgutschein(t *testing.T) {
	t.Parallel()

	// 03_Beispiele V_2_3_03: buying a single purpose voucher for 50 EUR and
	// redeeming it. UMS_BRUTTO is 50,00 on the purchase and 0,00 on the redemption.
	c := closingWithBons(
		beleg("40062", 62, 5000, []model.BonUSt{{Schluessel: 1, Brutto: euro(5000), Netto: euro(4202), USt: euro(798)}}, bar(5000),
			pos("1", schema.GVTypEinzweckgutscheinKauf, 1, 5000, 4202, 798)),
		beleg("40185", 185, 0, []model.BonUSt{{Schluessel: 1}}, keine(),
			pos("1", schema.GVTypUmsatz, 1, 5000, 4202, 798),
			pos("2", schema.GVTypEinzweckgutscheinEinloesung, 1, -5000, -4202, -798)),
	)

	rows := buildOne(t, c)

	wantGV := [][]string{
		{"EinzweckgutscheinKauf", "", "0", "1", "50,00", "42,02", "7,98"},
		{"Umsatz", "", "0", "1", "50,00", "42,02", "7,98"},
		{"EinzweckgutscheinEinloesung", "", "0", "1", "-50,00", "-42,02", "-7,98"},
	}
	if got := project(t, rows, "businesscases.csv", gvColumns...); !reflect.DeepEqual(got, wantGV) {
		t.Errorf("businesscases.csv =\n%v\nwant\n%v", got, wantGV)
	}

	wantPay := [][]string{{"Bar", "", "50,00"}, {"Keine", "", "0,00"}}
	if got := project(t, rows, "payment.csv", payColumns...); !reflect.DeepEqual(got, wantPay) {
		t.Errorf("payment.csv =\n%v\nwant\n%v", got, wantPay)
	}

	transactions := project(t, rows, "transactions.csv", "BON_ID", "UMS_BRUTTO")
	wantTransactions := [][]string{{"40062", "50,00"}, {"40185", "0,00"}}
	if !reflect.DeepEqual(transactions, wantTransactions) {
		t.Errorf("transactions.csv =\n%v\nwant\n%v", transactions, wantTransactions)
	}
}

func TestAggregateGeldtransit(t *testing.T) {
	t.Parallel()

	// 03_Beispiele V_2_3_06: 250,00 into the drawer in the morning, 4.200,00 out
	// in the evening, both as Geldtransit against UST_SCHLUESSEL 5.
	c := closingWithBons(
		beleg("10230", 1, 25000, []model.BonUSt{{Schluessel: 5, Brutto: euro(25000), Netto: euro(25000)}}, bar(25000),
			pos("1", schema.GVTypGeldtransit, 5, 25000, 25000, 0)),
		beleg("10315", 2, -420000, []model.BonUSt{{Schluessel: 5, Brutto: euro(-420000), Netto: euro(-420000)}}, bar(-420000),
			pos("1", schema.GVTypGeldtransit, 5, -420000, -420000, 0)),
	)

	rows := buildOne(t, c)

	wantGV := [][]string{{"Geldtransit", "", "0", "5", "-3950,00", "-3950,00", "0,00"}}
	if got := project(t, rows, "businesscases.csv", gvColumns...); !reflect.DeepEqual(got, wantGV) {
		t.Errorf("businesscases.csv =\n%v\nwant\n%v", got, wantGV)
	}

	wantCur := [][]string{{"EUR", "-3950,00"}}
	if got := project(t, rows, "cash_per_currency.csv", curColumns...); !reflect.DeepEqual(got, wantCur) {
		t.Errorf("cash_per_currency.csv =\n%v\nwant\n%v", got, wantCur)
	}

	if got := field(t, "transactions.csv", rows["transactions.csv"][1], "UMS_BRUTTO"); got != "-4200,00" {
		t.Errorf("UMS_BRUTTO = %q, want -4200,00", got)
	}
}

func TestAggregateForderung(t *testing.T) {
	t.Parallel()

	// 03_Beispiele V_2_3_01 b: 23,00 consumed without payment, settled in cash
	// two days later against UST_SCHLUESSEL 1 and 2.
	c := closingWithBons(
		beleg("20025", 25, 0, []model.BonUSt{{Schluessel: 5}}, keine(),
			pos("1", schema.GVTypUmsatz, 5, 1000, 1000, 0),
			pos("2", schema.GVTypUmsatz, 5, 300, 300, 0),
			pos("3", schema.GVTypUmsatz, 5, 1000, 1000, 0),
			pos("4", schema.GVTypForderungsentstehung, 5, -2300, -2300, 0)),
		beleg("20080", 80, 2300, []model.BonUSt{
			{Schluessel: 1, Brutto: euro(1300), Netto: euro(1092), USt: euro(208)},
			{Schluessel: 2, Brutto: euro(1000), Netto: euro(935), USt: euro(65)},
		}, bar(2300),
			pos("1", schema.GVTypForderungsaufloesung, 1, 1300, 1092, 208),
			pos("2", schema.GVTypForderungsaufloesung, 2, 1000, 935, 65)),
	)

	rows := buildOne(t, c)

	wantGV := [][]string{
		{"Umsatz", "", "0", "5", "23,00", "23,00", "0,00"},
		{"Forderungsentstehung", "", "0", "5", "-23,00", "-23,00", "0,00"},
		{"Forderungsaufloesung", "", "0", "1", "13,00", "10,92", "2,08"},
		{"Forderungsaufloesung", "", "0", "2", "10,00", "9,35", "0,65"},
	}
	if got := project(t, rows, "businesscases.csv", gvColumns...); !reflect.DeepEqual(got, wantGV) {
		t.Errorf("businesscases.csv =\n%v\nwant\n%v", got, wantGV)
	}

	wantPay := [][]string{{"Keine", "", "0,00"}, {"Bar", "", "23,00"}}
	if got := project(t, rows, "payment.csv", payColumns...); !reflect.DeepEqual(got, wantPay) {
		t.Errorf("payment.csv =\n%v\nwant\n%v", got, wantPay)
	}
}

func TestAggregateSeparatesAgentur(t *testing.T) {
	t.Parallel()

	c := closingWithBons(beleg("1", 1, 2000, nil, bar(2000),
		pos("1", schema.GVTypUmsatz, 1, 1000, 840, 160),
		model.Position{
			Zeile: "2", GVTyp: schema.GVTypUmsatz, AgenturID: 3,
			USt: []model.PosUSt{{Schluessel: 1, Brutto: euro(1000), Netto: euro(840), USt: euro(160)}},
		}))
	c.Agenturen = []model.Agentur{{ID: 3, Address: model.Address{Name: "Lotto AG"}}}

	want := [][]string{
		{"Umsatz", "", "0", "1", "10,00", "8,40", "1,60"},
		{"Umsatz", "", "3", "1", "10,00", "8,40", "1,60"},
	}
	if got := project(t, buildOne(t, c), "businesscases.csv", gvColumns...); !reflect.DeepEqual(got, want) {
		t.Errorf("businesscases.csv =\n%v\nwant\n%v", got, want)
	}
}

func TestAggregateSeparatesGVName(t *testing.T) {
	t.Parallel()

	first := pos("1", schema.GVTypUmsatz, 1, 1000, 840, 160)
	second := pos("2", schema.GVTypUmsatz, 1, 500, 420, 80)
	second.GVName = "Theke"

	c := closingWithBons(beleg("1", 1, 1500, nil, bar(1500), first, second))

	want := [][]string{
		{"Umsatz", "", "0", "1", "10,00", "8,40", "1,60"},
		{"Umsatz", "Theke", "0", "1", "5,00", "4,20", "0,80"},
	}
	if got := project(t, buildOne(t, c), "businesscases.csv", gvColumns...); !reflect.DeepEqual(got, want) {
		t.Errorf("businesscases.csv =\n%v\nwant\n%v", got, want)
	}
}

func TestAggregateForeignCurrencyCash(t *testing.T) {
	t.Parallel()

	c := closingWithBons(beleg("1", 1, 2000, nil, []model.Zahlung{
		{Typ: schema.ZahlartTypBar, BetragBasis: euro(1000)},
		{Typ: schema.ZahlartTypBar, Waehrung: "CHF", BetragWaehrung: euro(950), BetragBasis: euro(1000)},
	}))

	rows := buildOne(t, c)

	wantCur := [][]string{{"EUR", "10,00"}, {"CHF", "9,50"}}
	if got := project(t, rows, "cash_per_currency.csv", curColumns...); !reflect.DeepEqual(got, wantCur) {
		t.Errorf("cash_per_currency.csv =\n%v\nwant\n%v", got, wantCur)
	}

	row := onlyRow(t, rows, "cashpointclosing.csv")
	if got := field(t, "cashpointclosing.csv", row, "Z_SE_BARZAHLUNGEN"); got != "20,00" {
		t.Errorf("Z_SE_BARZAHLUNGEN = %q, want 20,00", got)
	}
}

func TestAggregateIgnoresNonBelegBons(t *testing.T) {
	t.Parallel()

	c := closingWithBons(model.Bon{
		ID: "1", Nr: 1, Typ: schema.BonTypAVTraining, UmsBrutto: euro(1000),
		Zahlungen:  bar(1000),
		Positionen: []model.Position{pos("1", schema.GVTypUmsatz, 1, 1000, 840, 160)},
	})

	rows := buildOne(t, c)

	for _, file := range []string{"businesscases.csv", "payment.csv", "cash_per_currency.csv"} {
		if got := len(rows[file]); got != 0 {
			t.Errorf("%s has %d rows, want 0", file, got)
		}
	}
	if got := len(rows["lines.csv"]); got != 1 {
		t.Errorf("lines.csv has %d rows, want 1", got)
	}
	if got := len(rows["transactions.csv"]); got != 1 {
		t.Errorf("transactions.csv has %d rows, want 1", got)
	}
}
