package model_test

import (
	"errors"
	"io"
	"testing"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

// huge is half of the largest representable Decimal, so two of them overflow.
var huge = dsfinvk.MaxDecimal

func TestBuildPropagatesOverflow(t *testing.T) {
	t.Parallel()

	cases := map[string]model.Bon{
		"POS_BRUTTO": bonWith(
			model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz, USt: []model.PosUSt{{Schluessel: 1, Brutto: huge}}},
			model.Position{Zeile: "2", GVTyp: schema.GVTypUmsatz, USt: []model.PosUSt{{Schluessel: 1, Brutto: huge}}},
		),
		"POS_NETTO": bonWith(
			model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz, USt: []model.PosUSt{{Schluessel: 1, Netto: huge}}},
			model.Position{Zeile: "2", GVTyp: schema.GVTypUmsatz, USt: []model.PosUSt{{Schluessel: 1, Netto: huge}}},
		),
		"POS_UST": bonWith(
			model.Position{Zeile: "1", GVTyp: schema.GVTypUmsatz, USt: []model.PosUSt{{Schluessel: 1, USt: huge}}},
			model.Position{Zeile: "2", GVTyp: schema.GVTypUmsatz, USt: []model.PosUSt{{Schluessel: 1, USt: huge}}},
		),
		"BASISWAEH_BETRAG": {
			ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
			Zahlungen: []model.Zahlung{
				{Typ: schema.ZahlartTypUnbar, BetragBasis: huge},
				{Typ: schema.ZahlartTypUnbar, BetragBasis: huge},
			},
		},
		"Z_SE_ZAHLUNGEN": {
			ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
			Zahlungen: []model.Zahlung{
				{Typ: schema.ZahlartTypUnbar, Name: "a", BetragBasis: huge},
				{Typ: schema.ZahlartTypUnbar, Name: "b", BetragBasis: huge},
			},
		},
		"Z_SE_BARZAHLUNGEN": {
			ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
			Zahlungen: []model.Zahlung{
				{Typ: schema.ZahlartTypBar, Name: "a", BetragBasis: huge},
				{Typ: schema.ZahlartTypBar, Name: "b", BetragBasis: huge},
			},
		},
		"ZAHLART_BETRAG_WAEH": {
			ID: "1", Nr: 1, Typ: schema.BonTypBeleg,
			Zahlungen: []model.Zahlung{
				{Typ: schema.ZahlartTypBar, Name: "a", Waehrung: "CHF", BetragWaehrung: huge},
				{Typ: schema.ZahlartTypBar, Name: "b", Waehrung: "CHF", BetragWaehrung: huge},
			},
		},
	}
	for name, bon := range cases {
		c := closingWithBons(bon)
		if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, dsfinvk.ErrOverflow) {
			t.Errorf("%s: Build error = %v, want dsfinvk.ErrOverflow", name, err)
		}
	}
}

func TestBuildEmptyCertificate(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	tse := testTSE()
	tse.Zertifikat = ""
	c.TSEs = []model.TSE{tse}

	row := onlyRow(t, buildOne(t, c), "tse.csv")

	for _, column := range []string{"TSE_ZERTIFIKAT_I", "TSE_ZERTIFIKAT_II"} {
		if got := field(t, "tse.csv", row, column); got != "" {
			t.Errorf("%s = %q, want empty", column, got)
		}
	}
}

// failingSink refuses to create one named file.
type failingSink struct{ fail string }

var errCreate = errors.New("create refused")

func (s failingSink) Create(name string) (io.WriteCloser, error) {
	if name == s.fail {
		return nil, errCreate
	}
	return discardCloser{}, nil
}

func (s failingSink) Close() error { return nil }

// discardCloser throws away everything written to it.
type discardCloser struct{}

func (discardCloser) Write(p []byte) (int, error) { return len(p), nil }
func (discardCloser) Close() error                { return nil }

func TestWriteExportReportsSinkError(t *testing.T) {
	t.Parallel()

	e := model.Export{Abschluesse: []model.Kassenabschluss{minimalClosing()}}

	err := model.WriteExport(e, failingSink{fail: "cashpointclosing.csv"}, csvio.DataSupplier{})
	if !errors.Is(err, errCreate) {
		t.Fatalf("WriteExport error = %v, want errCreate", err)
	}
}
