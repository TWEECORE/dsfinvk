# dsfinvk

Go toolkit for DSFinV-K 2.4: generate, parse, and validate German cash register exports.

## Usage

Generating an export takes three steps: describe the Kassenabschluss with the types in
`model`, pick a target (`csvio.NewDirSink` or `csvio.NewZipSink`), and call
`model.WriteExport`. One call writes all 20 CSV files, `index.xml`, and the DTD.

```go
package main

import (
	"log"
	"time"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

func main() {
	berlin, _ := time.LoadLocation("Europe/Berlin")
	start := time.Date(2026, 9, 3, 12, 15, 0, 0, berlin)

	bon := model.Bon{
		ID:         "1001",
		Nr:         1,
		Typ:        schema.BonTypBeleg,
		Start:      start,
		Ende:       start.Add(40 * time.Second),
		BedienerID: "42",
		UmsBrutto:  dsfinvk.FromCents(1190),
		Positionen: []model.Position{{
			Zeile:       "1",
			Artikeltext: "Kaffee",
			GVTyp:       schema.GVTypUmsatz,
			Menge:       dsfinvk.FromInt(2),
			StkBrutto:   dsfinvk.FromCents(595),
			USt: []model.PosUSt{{
				Schluessel: 1,
				Brutto:     dsfinvk.FromCents(1190),
				Netto:      dsfinvk.FromCents(1000),
				USt:        dsfinvk.FromCents(190),
			}},
		}},
		Zahlungen: []model.Zahlung{{
			Typ: schema.ZahlartTypBar, Waehrung: "EUR", BetragBasis: dsfinvk.FromCents(1190),
		}},
		USt: []model.BonUSt{{
			Schluessel: 1,
			Brutto:     dsfinvk.FromCents(1190),
			Netto:      dsfinvk.FromCents(1000),
			USt:        dsfinvk.FromCents(190),
		}},
		TSE: &model.TSETransaktion{
			TSEID:       1,
			Nr:          17,
			Start:       start.UTC(),
			Ende:        start.Add(40 * time.Second).UTC(),
			Vorgangsart: schema.TSEVorgangsartKassenbelegV1,
			SigZaehler:  53,
			Signatur:    "MEQCIB...base64...",
		},
	}

	abschluss := model.Kassenabschluss{
		KasseID:    "KASSE-01",
		Erstellung: time.Date(2026, 9, 3, 22, 0, 0, 0, berlin),
		Nr:         1,
		Unternehmen: model.Address{
			Name: "Beispiel GmbH", Strasse: "Musterweg 1", PLZ: "10115", Ort: "Berlin",
			Land: "DEU", StNr: "1234567890",
		},
		Standort: model.Location{
			Name: "Filiale Mitte", Strasse: "Musterweg 1", PLZ: "10115", Ort: "Berlin", Land: "DEU",
		},
		Kasse: model.Kasse{Seriennr: "SN-0001", Basiswaehrung: "EUR"},
		TSEs: []model.TSE{{
			ID:         1,
			Serial:     "a1b2c3",
			SigAlgo:    schema.TSESigAlgoECDSAPlainSHA256,
			Zeitformat: schema.TSEZeitformatUnixTime,
			PDEncoding: schema.TSEPDEncodingUTF8,
			PublicKey:  "MFkw...base64...",
			Zertifikat: "MIIB...base64...",
		}},
		USt:  []model.USt{{Schluessel: 1, Satz: dsfinvk.FromInt(19), Beschreibung: "Allgemeiner Steuersatz"}},
		Bons: []model.Bon{bon},
	}

	sink, err := csvio.NewDirSink("export-2026-09-03", false)
	if err != nil {
		log.Fatal(err)
	}
	export := model.Export{Abschluesse: []model.Kassenabschluss{abschluss}}
	if err := model.WriteExport(export, sink, csvio.DataSupplier{Name: "Beispiel GmbH"}); err != nil {
		log.Fatal(err)
	}
}
```

What to know when filling the model:

- **Amounts are `dsfinvk.Decimal`**, never floats. Build them with `dsfinvk.FromCents`, `dsfinvk.FromInt`,
  `dsfinvk.FromScaled`, or `dsfinvk.ParseComma("11,90")`.
- **`Bon.UmsBrutto` and `Bon.USt` are the values printed on the receipt.** Store them when the receipt is
  issued; the builder copies them verbatim and never derives them from positions (Spec 2.4 p.19, p.90).
- **`Bon.TSE` holds what your TSE returned** when the transaction was finished: transaction number,
  signature counter, signature, start and end time. Persist them at that moment; they cannot be
  reconstructed later. Times are written as UTC with milliseconds.
- **Closing aggregates are computed for you**: `payment.csv`, `businesscases.csv`,
  `cash_per_currency.csv`, `Z_SE_ZAHLUNGEN`, `Z_SE_BARZAHLUNGEN`, `Z_START_ID`, `Z_ENDE_ID`. Only Bons of
  type `Beleg` enter them; `AV*` types are still written to `transactions.csv`.
- **Booleans and sentinels** follow the spec: `Storno`, `Inhaus`, `KeineUStZuordnung` become `0`/`1`;
  `AgenturID` 0 means "own company". Long TSE certificates are split into `TSE_ZERTIFIKAT_I`, `_II`,
  `_III`, ... and `index.xml` is extended automatically.
- **Errors** from `WriteExport` are `errors.Is`-comparable: `model.ErrDuplicateBon`, `model.ErrUnknownUSt`,
  `model.ErrUnknownTSE`, `model.ErrUnknownAgentur`, `model.ErrEnumValue`, `csvio.ErrTooLong`,
  `csvio.ErrAccuracy`, `dsfinvk.ErrOverflow`.
- Use `csvio.NewZipSink(w)` instead of `NewDirSink` to produce a single archive. `csvio.OpenDir` and
  `csvio.OpenZip` read an export back; `csvio.ReadIndex` returns its table set.
