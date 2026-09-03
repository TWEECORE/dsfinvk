// Package model holds the DSFinV-K domain types and the builder that renders
// them to the CSV rows of an export.
package model

import (
	"time"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/schema"
)

// Address is the postal and tax identity of a company. Spec 2.4 p.67, p.68, p.74 to p.76.
type Address struct {
	Name    string
	Strasse string
	PLZ     string
	Ort     string
	Land    string
	StNr    string
	UStID   string
}

// Location is the settlement place of the Kasse, written to location.csv. Spec 2.4 p.69, p.70.
type Location struct {
	Name    string
	Strasse string
	PLZ     string
	Ort     string
	Land    string
	UStID   string
}

// Kasse is the cash register itself, written to cashregister.csv. Spec 2.4 p.71, p.72.
type Kasse struct {
	Brand             string
	Modell            string
	Seriennr          string
	SWBrand           string
	SWVersion         string
	Basiswaehrung     string
	KeineUStZuordnung bool
}

// Terminal is a slave cash register, written to slaves.csv. Spec 2.4 p.73, p.74.
type Terminal struct {
	ID        string
	Brand     string
	Modell    string
	Seriennr  string
	SWBrand   string
	SWVersion string
}

// Agentur is a principal the Kasse collects for, written to pa.csv. Spec 2.4 p.74 to p.76.
type Agentur struct {
	ID int64
	Address
}

// TSE is a technical security element, written to tse.csv. Spec 2.4 p.77 to p.79.
type TSE struct {
	ID         int64
	Serial     string
	SigAlgo    schema.TSESigAlgo
	Zeitformat schema.TSEZeitformat
	PDEncoding schema.TSEPDEncoding
	PublicKey  string
	Zertifikat string
}

// USt is one VAT rate of the Kassenabschluss, written to vat.csv. Spec 2.4 p.76, p.77.
type USt struct {
	Schluessel   int64
	Satz         dsfinvk.Decimal
	Beschreibung string
}

// Kunde is the recipient of the supply, written to transactions.csv. Spec 2.4 p.87 to p.89.
type Kunde struct {
	Name    string
	ID      string
	Typ     string
	Strasse string
	PLZ     string
	Ort     string
	Land    string
	UStID   string
}

// Zahlung is one payment of a Bon, written to datapayment.csv. Spec 2.4 p.91, p.92.
type Zahlung struct {
	Typ            schema.ZahlartTyp
	Name           string
	Waehrung       string
	BetragWaehrung dsfinvk.Decimal
	BetragBasis    dsfinvk.Decimal
}

// BonUSt is one printed VAT total of a Bon, written to transactions_vat.csv. Spec 2.4 p.90, p.91.
type BonUSt struct {
	Schluessel int64
	Brutto     dsfinvk.Decimal
	Netto      dsfinvk.Decimal
	USt        dsfinvk.Decimal
}

// PosUSt is one VAT share of a Position, written to lines_vat.csv. Spec 2.4 p.97.
type PosUSt struct {
	Schluessel int64
	Brutto     dsfinvk.Decimal
	Netto      dsfinvk.Decimal
	USt        dsfinvk.Decimal
}

// Preisfindung is a base price, discount or surcharge, written to itemamounts.csv. Spec 2.4 p.98, p.99.
type Preisfindung struct {
	Typ        schema.PreisfindungTyp
	Schluessel int64
	Brutto     dsfinvk.Decimal
	Netto      dsfinvk.Decimal
	USt        dsfinvk.Decimal
}

// Zusatzinfo is a sub item of a Position, written to subitems.csv. Spec 2.4 p.99 to p.101.
type Zusatzinfo struct {
	ArtNr            string
	GTIN             string
	Name             string
	WarengrID        string
	Warengr          string
	Menge            dsfinvk.Decimal
	Faktor           dsfinvk.Decimal
	Einheit          string
	Schluessel       int64
	BasispreisBrutto dsfinvk.Decimal
	BasispreisNetto  dsfinvk.Decimal
	BasispreisUSt    dsfinvk.Decimal
}

// Referenz points at another Vorgang or an external document, written to references.csv. Spec 2.4 p.102, p.103.
type Referenz struct {
	PosZeile string
	Typ      schema.RefTyp
	Name     string
	Datum    time.Time
	KasseID  string
	Nr       int64
	BonID    string
}

// TSETransaktion is the TSE signature of a Bon, written to transactions_tse.csv. Spec 2.4 p.104, p.105.
type TSETransaktion struct {
	TSEID         int64
	Nr            int64
	Start         time.Time
	Ende          time.Time
	Vorgangsart   schema.TSEVorgangsart
	SigZaehler    int64
	Signatur      string
	Fehler        string
	Vorgangsdaten string
}

// Position is one line of a Bon, written to lines.csv. Spec 2.4 p.93 to p.96.
type Position struct {
	Zeile        string
	GutscheinNr  string
	Artikeltext  string
	TerminalID   string
	GVTyp        schema.GVTyp
	GVName       string
	Inhaus       bool
	Storno       bool
	AgenturID    int64
	ArtNr        string
	GTIN         string
	WarengrID    string
	Warengr      string
	Menge        dsfinvk.Decimal
	Faktor       dsfinvk.Decimal
	Einheit      string
	StkBrutto    dsfinvk.Decimal
	USt          []PosUSt
	Preisfindung []Preisfindung
	Zusatzinfos  []Zusatzinfo
}

// Bon is one Vorgang, written to transactions.csv and its child tables. Spec 2.4 p.84 to p.89.
type Bon struct {
	ID                string
	Nr                int64
	Typ               schema.BonTyp
	Name              string
	TerminalID        string
	Storno            bool
	Start             time.Time
	Ende              time.Time
	BedienerID        string
	BedienerName      string
	UmsBrutto         dsfinvk.Decimal
	Kunde             Kunde
	Notiz             string
	Positionen        []Position
	Zahlungen         []Zahlung
	USt               []BonUSt
	Abrechnungskreise []string
	Referenzen        []Referenz
	TSE               *TSETransaktion
}

// Kassenabschluss is one cash register closing with its master data and Bons. Spec 2.4 p.64 to p.69.
type Kassenabschluss struct {
	KasseID          string
	Erstellung       time.Time
	Nr               int64
	Buchungstag      time.Time
	TaxonomieVersion string
	Unternehmen      Address
	Standort         Location
	Kasse            Kasse
	Terminals        []Terminal
	Agenturen        []Agentur
	TSEs             []TSE
	USt              []USt
	Bons             []Bon
}

// Export is the set of Kassenabschluesse one DSFinV-K export carries.
type Export struct {
	Abschluesse []Kassenabschluss
}
