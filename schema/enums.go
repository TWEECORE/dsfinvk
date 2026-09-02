package schema

import (
	"slices"
	"strconv"
)

// Bool01False and Bool01True are the only values a DSFinV-K boolean column may
// carry. Anhang G legend K228: "Boolean-Felder werden mit 0 oder 1 exportiert".
const (
	Bool01False = "0"
	Bool01True  = "1"
)

// ValidBool01 reports whether s is a permitted boolean column value.
func ValidBool01(s string) bool { return s == Bool01False || s == Bool01True }

// The four boolean columns of Anhang E. Spec 2.4 p.72, p.85, p.94.
var bool01Columns = map[string]map[string]bool{
	"cashregister.csv": {"KEINE_UST_ZUORDNUNG": true},
	"transactions.csv": {"BON_STORNO": true},
	"lines.csv":        {"INHAUS": true, "P_STORNO": true},
}

// IsBool01Column reports whether the column takes only Bool01False or Bool01True.
func IsBool01Column(file, column string) bool { return bool01Columns[file][column] }

// ValidISOAlpha3 reports whether s has the shape of an ISO alpha-3 code: exactly
// three ASCII uppercase letters. Currency spec 2.4 p.72, p.83, p.92; country
// p.68, p.70, p.75, p.89. The code lists themselves are not checked.
func ValidISOAlpha3(s string) bool {
	if len(s) != 3 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 'A' || s[i] > 'Z' {
			return false
		}
	}
	return true
}

// BonTyp is the transaction type, written to transactions.BON_TYP. Spec 2.4 p.43.
type BonTyp string

const (
	BonTypBeleg          BonTyp = "Beleg"
	BonTypAVRechnung     BonTyp = "AVRechnung"
	BonTypAVTransfer     BonTyp = "AVTransfer"
	BonTypAVBestellung   BonTyp = "AVBestellung"
	BonTypAVTraining     BonTyp = "AVTraining"
	BonTypAVBelegstorno  BonTyp = "AVBelegstorno"
	BonTypAVBelegabbruch BonTyp = "AVBelegabbruch"
	BonTypAVSachbezug    BonTyp = "AVSachbezug"
	BonTypAVSonstige     BonTyp = "AVSonstige"
)

var bonTypValues = []string{
	string(BonTypBeleg),
	string(BonTypAVRechnung),
	string(BonTypAVTransfer),
	string(BonTypAVBestellung),
	string(BonTypAVTraining),
	string(BonTypAVBelegstorno),
	string(BonTypAVBelegabbruch),
	string(BonTypAVSachbezug),
	string(BonTypAVSonstige),
}

// BonTypValues returns the permitted BON_TYP values in specification order.
func BonTypValues() []string { return slices.Clone(bonTypValues) }

// ValidBonTyp reports whether s is a permitted BON_TYP.
func ValidBonTyp(s string) bool { return slices.Contains(bonTypValues, s) }

// AllowedInProcessData reports whether the type may appear as the processData
// <Vorgangstyp>. Spec 2.4 p.113: AVBelegstorno cannot be used with a TSE.
func (b BonTyp) AllowedInProcessData() bool { return b != BonTypAVBelegstorno }

// GVTyp is the business case type, written to businesscases.GV_TYP. Spec 2.4 p.48-49.
type GVTyp string

const (
	GVTypUmsatz                       GVTyp = "Umsatz"
	GVTypPfand                        GVTyp = "Pfand"
	GVTypPfandRueckzahlung            GVTyp = "PfandRueckzahlung"
	GVTypRabatt                       GVTyp = "Rabatt"
	GVTypAufschlag                    GVTyp = "Aufschlag"
	GVTypZuschussEcht                 GVTyp = "ZuschussEcht"
	GVTypZuschussUnecht               GVTyp = "ZuschussUnecht"
	GVTypTrinkgeldAG                  GVTyp = "TrinkgeldAG"
	GVTypTrinkgeldAN                  GVTyp = "TrinkgeldAN"
	GVTypEinzweckgutscheinKauf        GVTyp = "EinzweckgutscheinKauf"
	GVTypEinzweckgutscheinEinloesung  GVTyp = "EinzweckgutscheinEinloesung"
	GVTypMehrzweckgutscheinKauf       GVTyp = "MehrzweckgutscheinKauf"
	GVTypMehrzweckgutscheinEinloesung GVTyp = "MehrzweckgutscheinEinloesung"
	GVTypForderungsentstehung         GVTyp = "Forderungsentstehung"
	GVTypForderungsaufloesung         GVTyp = "Forderungsaufloesung"
	GVTypAnzahlungseinstellung        GVTyp = "Anzahlungseinstellung"
	GVTypAnzahlungsaufloesung         GVTyp = "Anzahlungsaufloesung"
	GVTypAnfangsbestand               GVTyp = "Anfangsbestand"
	GVTypPrivatentnahme               GVTyp = "Privatentnahme"
	GVTypPrivateinlage                GVTyp = "Privateinlage"
	GVTypGeldtransit                  GVTyp = "Geldtransit"
	GVTypLohnzahlung                  GVTyp = "Lohnzahlung"
	GVTypEinzahlung                   GVTyp = "Einzahlung"
	GVTypAuszahlung                   GVTyp = "Auszahlung"
	GVTypDifferenzSollIst             GVTyp = "DifferenzSollIst"
)

var gvTypValues = []string{
	string(GVTypUmsatz),
	string(GVTypPfand),
	string(GVTypPfandRueckzahlung),
	string(GVTypRabatt),
	string(GVTypAufschlag),
	string(GVTypZuschussEcht),
	string(GVTypZuschussUnecht),
	string(GVTypTrinkgeldAG),
	string(GVTypTrinkgeldAN),
	string(GVTypEinzweckgutscheinKauf),
	string(GVTypEinzweckgutscheinEinloesung),
	string(GVTypMehrzweckgutscheinKauf),
	string(GVTypMehrzweckgutscheinEinloesung),
	string(GVTypForderungsentstehung),
	string(GVTypForderungsaufloesung),
	string(GVTypAnzahlungseinstellung),
	string(GVTypAnzahlungsaufloesung),
	string(GVTypAnfangsbestand),
	string(GVTypPrivatentnahme),
	string(GVTypPrivateinlage),
	string(GVTypGeldtransit),
	string(GVTypLohnzahlung),
	string(GVTypEinzahlung),
	string(GVTypAuszahlung),
	string(GVTypDifferenzSollIst),
}

// GVTypValues returns the permitted GV_TYP values in specification order.
func GVTypValues() []string { return slices.Clone(gvTypValues) }

// ValidGVTyp reports whether s is a permitted GV_TYP.
func ValidGVTyp(s string) bool { return slices.Contains(gvTypValues, s) }

var cashOnlyGVTypen = []string{
	string(GVTypAnfangsbestand),
	string(GVTypPrivatentnahme),
	string(GVTypPrivateinlage),
	string(GVTypGeldtransit),
	string(GVTypLohnzahlung),
	string(GVTypEinzahlung),
	string(GVTypAuszahlung),
	string(GVTypDifferenzSollIst),
}

// AffectsCashOnly reports whether the type affects only the cash holding. Spec 2.4 p.34.
func (g GVTyp) AffectsCashOnly() bool { return slices.Contains(cashOnlyGVTypen, string(g)) }

// ZahlartTyp is the payment type, written to payment.ZAHLART_TYP. Spec 2.4 p.61.
type ZahlartTyp string

const (
	ZahlartTypBar                     ZahlartTyp = "Bar"
	ZahlartTypUnbar                   ZahlartTyp = "Unbar"
	ZahlartTypKeine                   ZahlartTyp = "Keine"
	ZahlartTypECKarte                 ZahlartTyp = "ECKarte"
	ZahlartTypKreditkarte             ZahlartTyp = "Kreditkarte"
	ZahlartTypElZahlungsdienstleister ZahlartTyp = "ElZahlungsdienstleister"
	ZahlartTypGuthabenkarte           ZahlartTyp = "Guthabenkarte"
)

var zahlartTypValues = []string{
	string(ZahlartTypBar),
	string(ZahlartTypUnbar),
	string(ZahlartTypKeine),
	string(ZahlartTypECKarte),
	string(ZahlartTypKreditkarte),
	string(ZahlartTypElZahlungsdienstleister),
	string(ZahlartTypGuthabenkarte),
}

// ZahlartTypValues returns the permitted ZAHLART_TYP values in specification order.
func ZahlartTypValues() []string { return slices.Clone(zahlartTypValues) }

// ValidZahlartTyp reports whether s is a permitted ZAHLART_TYP.
func ValidZahlartTyp(s string) bool { return slices.Contains(zahlartTypValues, s) }

// ProcessDataPaymentClass is the payment split of the processData <Zahlungen>
// field, which knows only "Bar" and "Unbar". Spec 2.4 p.114.
type ProcessDataPaymentClass uint8

const (
	PaymentClassBar ProcessDataPaymentClass = iota
	PaymentClassUnbar
)

// String returns the processData spelling of the class.
func (c ProcessDataPaymentClass) String() string {
	switch c {
	case PaymentClassBar:
		return "Bar"
	case PaymentClassUnbar:
		return "Unbar"
	default:
		return "ProcessDataPaymentClass(" + strconv.Itoa(int(c)) + ")"
	}
}

// Anhang D p.61 to p.63: every card and electronic payment type is a form of
// "Unbar"; "Keine" stands for a transaction closed without any payment.
var processDataPaymentClasses = map[ZahlartTyp]ProcessDataPaymentClass{
	ZahlartTypBar:                     PaymentClassBar,
	ZahlartTypUnbar:                   PaymentClassUnbar,
	ZahlartTypECKarte:                 PaymentClassUnbar,
	ZahlartTypKreditkarte:             PaymentClassUnbar,
	ZahlartTypElZahlungsdienstleister: PaymentClassUnbar,
	ZahlartTypGuthabenkarte:           PaymentClassUnbar,
}

// ProcessDataClass returns the processData payment class of the type; ok is
// false for ZahlartTypKeine and for any value outside the enumeration.
func (z ZahlartTyp) ProcessDataClass() (ProcessDataPaymentClass, bool) {
	c, ok := processDataPaymentClasses[z]
	return c, ok
}

// RefTyp is the reference type, written to references.REF_TYP. Spec 2.4 p.102.
type RefTyp string

const (
	RefTypExterneRechnung      RefTyp = "ExterneRechnung"
	RefTypExternerLieferschein RefTyp = "ExternerLieferschein"
	RefTypTransaktion          RefTyp = "Transaktion"
	RefTypExterneSonstige      RefTyp = "ExterneSonstige"
)

var refTypValues = []string{
	string(RefTypExterneRechnung),
	string(RefTypExternerLieferschein),
	string(RefTypTransaktion),
	string(RefTypExterneSonstige),
}

// RefTypValues returns the permitted REF_TYP values in specification order.
func RefTypValues() []string { return slices.Clone(refTypValues) }

// ValidRefTyp reports whether s is a permitted REF_TYP.
func ValidRefTyp(s string) bool { return slices.Contains(refTypValues, s) }

// PreisfindungTyp is the price-finding type, written to itemamounts.TYP.
// Spec 2.4 p.98 accepts an English and a German spelling; the English is canonical.
type PreisfindungTyp string

// Spec 2.4 p.98, canonical English spellings.
const (
	PreisfindungBaseAmount  PreisfindungTyp = "base_amount"
	PreisfindungDiscount    PreisfindungTyp = "discount"
	PreisfindungExtraAmount PreisfindungTyp = "extra_amount"
)

// Spec 2.4 p.98, German alternative spellings.
const (
	PreisfindungGrundpreis PreisfindungTyp = "Grundpreis"
	PreisfindungRabatt     PreisfindungTyp = "Rabatt"
	PreisfindungZuschlag   PreisfindungTyp = "Zuschlag"
)

var preisfindungTypValues = []string{
	string(PreisfindungBaseAmount), string(PreisfindungGrundpreis),
	string(PreisfindungDiscount), string(PreisfindungRabatt),
	string(PreisfindungExtraAmount), string(PreisfindungZuschlag),
}

var preisfindungTypCanonicalValues = []string{
	string(PreisfindungBaseAmount),
	string(PreisfindungDiscount),
	string(PreisfindungExtraAmount),
}

var preisfindungTypNormalized = map[string]PreisfindungTyp{
	string(PreisfindungBaseAmount):  PreisfindungBaseAmount,
	string(PreisfindungGrundpreis):  PreisfindungBaseAmount,
	string(PreisfindungDiscount):    PreisfindungDiscount,
	string(PreisfindungRabatt):      PreisfindungDiscount,
	string(PreisfindungExtraAmount): PreisfindungExtraAmount,
	string(PreisfindungZuschlag):    PreisfindungExtraAmount,
}

// PreisfindungTypValues returns every accepted itemamounts.TYP value, both spellings.
func PreisfindungTypValues() []string { return slices.Clone(preisfindungTypValues) }

// PreisfindungTypCanonicalValues returns only the canonical English spellings.
func PreisfindungTypCanonicalValues() []string { return slices.Clone(preisfindungTypCanonicalValues) }

// ValidPreisfindungTyp reports whether s is an accepted itemamounts.TYP.
func ValidPreisfindungTyp(s string) bool {
	_, ok := preisfindungTypNormalized[s]
	return ok
}

// NormalizePreisfindungTyp maps an accepted itemamounts.TYP onto its English form.
func NormalizePreisfindungTyp(s string) (PreisfindungTyp, bool) {
	v, ok := preisfindungTypNormalized[s]
	return v, ok
}

// TSESigAlgo is the TSE signature algorithm, written to tse.TSE_SIG_ALGO. Spec 2.4 p.77-78.
type TSESigAlgo string

const (
	TSESigAlgoECDSAPlainSHA224  TSESigAlgo = "ecdsa-plain-SHA224"
	TSESigAlgoECDSAPlainSHA256  TSESigAlgo = "ecdsa-plain-SHA256"
	TSESigAlgoECDSAPlainSHA384  TSESigAlgo = "ecdsa-plain-SHA384"
	TSESigAlgoECDSAPlainSHA512  TSESigAlgo = "ecdsa-plain-SHA512"
	TSESigAlgoECDSAPlainSHA3224 TSESigAlgo = "ecdsa-plain-SHA3-224"
	TSESigAlgoECDSAPlainSHA3256 TSESigAlgo = "ecdsa-plain-SHA3-256"
	TSESigAlgoECDSAPlainSHA3384 TSESigAlgo = "ecdsa-plain-SHA3-384"
	TSESigAlgoECDSAPlainSHA3512 TSESigAlgo = "ecdsa-plain-SHA3-512"

	// "ecsdsa" is the spec's spelling, not a typo.
	TSESigAlgoECSDSAPlainSHA224  TSESigAlgo = "ecsdsa-plain-SHA224"
	TSESigAlgoECSDSAPlainSHA256  TSESigAlgo = "ecsdsa-plain-SHA256"
	TSESigAlgoECSDSAPlainSHA384  TSESigAlgo = "ecsdsa-plain-SHA384"
	TSESigAlgoECSDSAPlainSHA512  TSESigAlgo = "ecsdsa-plain-SHA512"
	TSESigAlgoECSDSAPlainSHA3224 TSESigAlgo = "ecsdsa-plain-SHA3-224"
	TSESigAlgoECSDSAPlainSHA3256 TSESigAlgo = "ecsdsa-plain-SHA3-256"
	TSESigAlgoECSDSAPlainSHA3384 TSESigAlgo = "ecsdsa-plain-SHA3-384"
	TSESigAlgoECSDSAPlainSHA3512 TSESigAlgo = "ecsdsa-plain-SHA3-512"
)

var tseSigAlgoValues = []string{
	string(TSESigAlgoECDSAPlainSHA224),
	string(TSESigAlgoECDSAPlainSHA256),
	string(TSESigAlgoECDSAPlainSHA384),
	string(TSESigAlgoECDSAPlainSHA512),
	string(TSESigAlgoECDSAPlainSHA3224),
	string(TSESigAlgoECDSAPlainSHA3256),
	string(TSESigAlgoECDSAPlainSHA3384),
	string(TSESigAlgoECDSAPlainSHA3512),
	string(TSESigAlgoECSDSAPlainSHA224),
	string(TSESigAlgoECSDSAPlainSHA256),
	string(TSESigAlgoECSDSAPlainSHA384),
	string(TSESigAlgoECSDSAPlainSHA512),
	string(TSESigAlgoECSDSAPlainSHA3224),
	string(TSESigAlgoECSDSAPlainSHA3256),
	string(TSESigAlgoECSDSAPlainSHA3384),
	string(TSESigAlgoECSDSAPlainSHA3512),
}

// TSESigAlgoValues returns the permitted TSE_SIG_ALGO values in specification order.
func TSESigAlgoValues() []string { return slices.Clone(tseSigAlgoValues) }

// ValidTSESigAlgo reports whether s is a permitted TSE_SIG_ALGO.
func ValidTSESigAlgo(s string) bool { return slices.Contains(tseSigAlgoValues, s) }

// TSEZeitformat is the TSE log time format, written to tse.TSE_ZEITFORMAT. Spec 2.4 p.78.
type TSEZeitformat string

const (
	TSEZeitformatUnixTime                        TSEZeitformat = "unixTime"
	TSEZeitformatUTCTime                         TSEZeitformat = "utcTime"
	TSEZeitformatUTCTimeWithSeconds              TSEZeitformat = "utcTimeWithSeconds"
	TSEZeitformatGeneralizedTime                 TSEZeitformat = "generalizedTime"
	TSEZeitformatGeneralizedTimeWithMilliseconds TSEZeitformat = "generalizedTimeWithMilliseconds"
)

var tseZeitformatValues = []string{
	string(TSEZeitformatUnixTime),
	string(TSEZeitformatUTCTime),
	string(TSEZeitformatUTCTimeWithSeconds),
	string(TSEZeitformatGeneralizedTime),
	string(TSEZeitformatGeneralizedTimeWithMilliseconds),
}

// TSEZeitformatValues returns the permitted TSE_ZEITFORMAT values in specification order.
func TSEZeitformatValues() []string { return slices.Clone(tseZeitformatValues) }

// ValidTSEZeitformat reports whether s is a permitted TSE_ZEITFORMAT.
func ValidTSEZeitformat(s string) bool { return slices.Contains(tseZeitformatValues, s) }

// TSEPDEncoding is the TSE processData encoding, written to tse.TSE_PD_ENCODING. Spec 2.4 p.79.
type TSEPDEncoding string

const (
	TSEPDEncodingUTF8  TSEPDEncoding = "UTF-8"
	TSEPDEncodingASCII TSEPDEncoding = "ASCII"
)

var tsePDEncodingValues = []string{
	string(TSEPDEncodingUTF8),
	string(TSEPDEncodingASCII),
}

// TSEPDEncodingValues returns the permitted TSE_PD_ENCODING values in specification order.
func TSEPDEncodingValues() []string { return slices.Clone(tsePDEncodingValues) }

// ValidTSEPDEncoding reports whether s is a permitted TSE_PD_ENCODING.
func ValidTSEPDEncoding(s string) bool { return slices.Contains(tsePDEncodingValues, s) }

// TSEVorgangsart is the TSE processType in transactions_tse.TSE_TA_VORGANGSART. Spec 2.4 Anhang I.
type TSEVorgangsart string

const (
	TSEVorgangsartKassenbelegV1    TSEVorgangsart = "Kassenbeleg-V1"
	TSEVorgangsartBestellungV1     TSEVorgangsart = "Bestellung-V1"
	TSEVorgangsartSonstigerVorgang TSEVorgangsart = "SonstigerVorgang"
)

var tseVorgangsartValues = []string{
	string(TSEVorgangsartKassenbelegV1),
	string(TSEVorgangsartBestellungV1),
	string(TSEVorgangsartSonstigerVorgang),
}

// TSEVorgangsartValues returns the permitted TSE_TA_VORGANGSART values in specification order.
func TSEVorgangsartValues() []string { return slices.Clone(tseVorgangsartValues) }

// ValidTSEVorgangsart reports whether s is a permitted TSE_TA_VORGANGSART.
func ValidTSEVorgangsart(s string) bool { return slices.Contains(tseVorgangsartValues, s) }
