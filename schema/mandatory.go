package schema

import (
	"slices"
	"strconv"
)

// Requirement says how strictly a column must be filled. Source:
// Anhang_G_Uebersicht.xlsx, sheet "Uebersicht"; the marker fill maps to grey
// Mandatory, unfilled Conditional, yellow IfItem, green OneOf.
type Requirement uint8

const (
	RequirementConditional Requirement = iota
	RequirementMandatory
	RequirementIfItem
	RequirementOneOf
)

// String returns the name of the requirement class.
func (r Requirement) String() string {
	switch r {
	case RequirementConditional:
		return "Conditional"
	case RequirementMandatory:
		return "Mandatory"
	case RequirementIfItem:
		return "IfItem"
	case RequirementOneOf:
		return "OneOf"
	default:
		return "Requirement(" + strconv.Itoa(int(r)) + ")"
	}
}

// ColumnRequirement returns how strictly a column must be filled; ok is false
// when the Anhang G matrix does not classify the pair.
func ColumnRequirement(file, column string) (Requirement, bool) {
	r, ok := requirements[file][column]
	return r, ok
}

// Anhang G legend K231: "tax_number und vat_id_number: Es muss jeweils
// mindestens ein Feld gefüllt sein".
var oneOfGroups = map[string][][]string{
	"cashpointclosing.csv": {{"STNR", "USTID"}},
	"pa.csv":               {{"AGENTUR_STNR", "AGENTUR_USTID"}},
}

// OneOfGroups returns the column groups of a file of which at least one member
// must be filled; the result is a fresh copy.
func OneOfGroups(file string) [][]string {
	groups := oneOfGroups[file]
	out := make([][]string, len(groups))
	for i, g := range groups {
		out[i] = slices.Clone(g)
	}
	return out
}

// RequirementEntry is one classified column of the Anhang G matrix.
type RequirementEntry struct {
	File        string
	Column      string
	Requirement Requirement
}

// RequirementEntries returns the whole Anhang G matrix in index.xml order.
func RequirementEntries() []RequirementEntry {
	out := make([]RequirementEntry, 0, 219)
	for _, t := range tables {
		for _, c := range t.Columns {
			if r, ok := requirements[t.File][c.Name]; ok {
				out = append(out, RequirementEntry{File: t.File, Column: c.Name, Requirement: r})
			}
		}
	}
	return out
}

var requirements = map[string]map[string]Requirement{
	"cashpointclosing.csv": {
		"Z_KASSE_ID":        RequirementMandatory,
		"Z_ERSTELLUNG":      RequirementMandatory,
		"Z_NR":              RequirementMandatory,
		"Z_BUCHUNGSTAG":     RequirementConditional,
		"TAXONOMIE_VERSION": RequirementMandatory,
		"Z_START_ID":        RequirementMandatory,
		"Z_ENDE_ID":         RequirementMandatory,
		"NAME":              RequirementMandatory,
		"STRASSE":           RequirementMandatory,
		"PLZ":               RequirementMandatory,
		"ORT":               RequirementMandatory,
		"LAND":              RequirementMandatory,
		"STNR":              RequirementOneOf,
		"USTID":             RequirementOneOf,
		"Z_SE_ZAHLUNGEN":    RequirementMandatory,
		"Z_SE_BARZAHLUNGEN": RequirementMandatory,
	},
	"location.csv": {
		"Z_KASSE_ID":   RequirementMandatory,
		"Z_ERSTELLUNG": RequirementMandatory,
		"Z_NR":         RequirementMandatory,
		"LOC_NAME":     RequirementMandatory,
		"LOC_STRASSE":  RequirementMandatory,
		"LOC_PLZ":      RequirementMandatory,
		"LOC_ORT":      RequirementMandatory,
		"LOC_LAND":     RequirementMandatory,
		"LOC_USTID":    RequirementConditional,
	},
	"cashregister.csv": {
		"Z_KASSE_ID":           RequirementMandatory,
		"Z_ERSTELLUNG":         RequirementMandatory,
		"Z_NR":                 RequirementMandatory,
		"KASSE_BRAND":          RequirementConditional,
		"KASSE_MODELL":         RequirementConditional,
		"KASSE_SERIENNR":       RequirementMandatory,
		"KASSE_SW_BRAND":       RequirementConditional,
		"KASSE_SW_VERSION":     RequirementConditional,
		"KASSE_BASISWAEH_CODE": RequirementMandatory,
		"KEINE_UST_ZUORDNUNG":  RequirementConditional,
	},
	"slaves.csv": {
		"Z_KASSE_ID":          RequirementMandatory,
		"Z_ERSTELLUNG":        RequirementMandatory,
		"Z_NR":                RequirementMandatory,
		"TERMINAL_ID":         RequirementMandatory,
		"TERMINAL_BRAND":      RequirementConditional,
		"TERMINAL_MODELL":     RequirementConditional,
		"TERMINAL_SERIENNR":   RequirementMandatory,
		"TERMINAL_SW_BRAND":   RequirementConditional,
		"TERMINAL_SW_VERSION": RequirementConditional,
	},
	"pa.csv": {
		"Z_KASSE_ID":      RequirementMandatory,
		"Z_ERSTELLUNG":    RequirementMandatory,
		"Z_NR":            RequirementMandatory,
		"AGENTUR_ID":      RequirementMandatory,
		"AGENTUR_NAME":    RequirementMandatory,
		"AGENTUR_STRASSE": RequirementMandatory,
		"AGENTUR_PLZ":     RequirementMandatory,
		"AGENTUR_ORT":     RequirementMandatory,
		"AGENTUR_LAND":    RequirementMandatory,
		"AGENTUR_STNR":    RequirementOneOf,
		"AGENTUR_USTID":   RequirementOneOf,
	},
	"tse.csv": {
		"Z_KASSE_ID":        RequirementMandatory,
		"Z_ERSTELLUNG":      RequirementMandatory,
		"Z_NR":              RequirementMandatory,
		"TSE_ID":            RequirementMandatory,
		"TSE_SERIAL":        RequirementMandatory,
		"TSE_SIG_ALGO":      RequirementMandatory,
		"TSE_ZEITFORMAT":    RequirementMandatory,
		"TSE_PD_ENCODING":   RequirementMandatory,
		"TSE_PUBLIC_KEY":    RequirementMandatory,
		"TSE_ZERTIFIKAT_I":  RequirementMandatory,
		"TSE_ZERTIFIKAT_II": RequirementMandatory,
	},
	"vat.csv": {
		"Z_KASSE_ID":     RequirementMandatory,
		"Z_ERSTELLUNG":   RequirementMandatory,
		"Z_NR":           RequirementMandatory,
		"UST_SCHLUESSEL": RequirementMandatory,
		"UST_SATZ":       RequirementMandatory,
		"UST_BESCHR":     RequirementMandatory,
	},
	"businesscases.csv": {
		"Z_KASSE_ID":     RequirementMandatory,
		"Z_ERSTELLUNG":   RequirementMandatory,
		"Z_NR":           RequirementMandatory,
		"GV_TYP":         RequirementMandatory,
		"GV_NAME":        RequirementConditional,
		"AGENTUR_ID":     RequirementConditional,
		"UST_SCHLUESSEL": RequirementMandatory,
		"Z_UMS_BRUTTO":   RequirementMandatory,
		"Z_UMS_NETTO":    RequirementMandatory,
		"Z_UST":          RequirementMandatory,
	},
	"payment.csv": {
		"Z_KASSE_ID":       RequirementMandatory,
		"Z_ERSTELLUNG":     RequirementMandatory,
		"Z_NR":             RequirementMandatory,
		"ZAHLART_TYP":      RequirementMandatory,
		"ZAHLART_NAME":     RequirementConditional,
		"Z_ZAHLART_BETRAG": RequirementMandatory,
	},
	"cash_per_currency.csv": {
		"Z_KASSE_ID":          RequirementMandatory,
		"Z_ERSTELLUNG":        RequirementMandatory,
		"Z_NR":                RequirementMandatory,
		"ZAHLART_WAEH":        RequirementMandatory,
		"ZAHLART_BETRAG_WAEH": RequirementMandatory,
	},
	"transactions.csv": {
		"Z_KASSE_ID":    RequirementMandatory,
		"Z_ERSTELLUNG":  RequirementMandatory,
		"Z_NR":          RequirementMandatory,
		"BON_ID":        RequirementMandatory,
		"BON_NR":        RequirementMandatory,
		"BON_TYP":       RequirementMandatory,
		"BON_NAME":      RequirementConditional,
		"TERMINAL_ID":   RequirementConditional,
		"BON_STORNO":    RequirementConditional,
		"BON_START":     RequirementMandatory,
		"BON_ENDE":      RequirementMandatory,
		"BEDIENER_ID":   RequirementMandatory,
		"BEDIENER_NAME": RequirementConditional,
		"UMS_BRUTTO":    RequirementMandatory,
		"KUNDE_NAME":    RequirementMandatory,
		"KUNDE_ID":      RequirementMandatory,
		"KUNDE_TYP":     RequirementMandatory,
		"KUNDE_STRASSE": RequirementConditional,
		"KUNDE_PLZ":     RequirementConditional,
		"KUNDE_ORT":     RequirementConditional,
		"KUNDE_LAND":    RequirementConditional,
		"KUNDE_USTID":   RequirementConditional,
		"BON_NOTIZ":     RequirementConditional,
	},
	"datapayment.csv": {
		"Z_KASSE_ID":       RequirementMandatory,
		"Z_ERSTELLUNG":     RequirementMandatory,
		"Z_NR":             RequirementMandatory,
		"BON_ID":           RequirementMandatory,
		"ZAHLART_TYP":      RequirementMandatory,
		"ZAHLART_NAME":     RequirementConditional,
		"ZAHLWAEH_CODE":    RequirementMandatory,
		"ZAHLWAEH_BETRAG":  RequirementConditional,
		"BASISWAEH_BETRAG": RequirementMandatory,
	},
	"lines.csv": {
		"Z_KASSE_ID":      RequirementMandatory,
		"Z_ERSTELLUNG":    RequirementMandatory,
		"Z_NR":            RequirementMandatory,
		"BON_ID":          RequirementMandatory,
		"POS_ZEILE":       RequirementMandatory,
		"GUTSCHEIN_NR":    RequirementConditional,
		"ARTIKELTEXT":     RequirementMandatory,
		"POS_TERMINAL_ID": RequirementConditional,
		"GV_TYP":          RequirementMandatory,
		"GV_NAME":         RequirementConditional,
		"INHAUS":          RequirementConditional,
		"P_STORNO":        RequirementConditional,
		"AGENTUR_ID":      RequirementConditional,
		"ART_NR":          RequirementIfItem,
		"GTIN":            RequirementConditional,
		"WARENGR_ID":      RequirementConditional,
		"WARENGR":         RequirementConditional,
		"MENGE":           RequirementIfItem,
		"FAKTOR":          RequirementConditional,
		"EINHEIT":         RequirementConditional,
		"STK_BR":          RequirementIfItem,
	},
	"itemamounts.csv": {
		"Z_KASSE_ID":     RequirementMandatory,
		"Z_ERSTELLUNG":   RequirementMandatory,
		"Z_NR":           RequirementMandatory,
		"BON_ID":         RequirementMandatory,
		"POS_ZEILE":      RequirementMandatory,
		"TYP":            RequirementConditional,
		"UST_SCHLUESSEL": RequirementMandatory,
		"PF_BRUTTO":      RequirementConditional,
		"PF_NETTO":       RequirementConditional,
		"PF_UST":         RequirementConditional,
	},
	"subitems.csv": {
		"Z_KASSE_ID":           RequirementMandatory,
		"Z_ERSTELLUNG":         RequirementMandatory,
		"Z_NR":                 RequirementMandatory,
		"BON_ID":               RequirementMandatory,
		"POS_ZEILE":            RequirementMandatory,
		"ZI_ART_NR":            RequirementMandatory,
		"ZI_GTIN":              RequirementConditional,
		"ZI_NAME":              RequirementConditional,
		"ZI_WARENGR_ID":        RequirementConditional,
		"ZI_WARENGR":           RequirementConditional,
		"ZI_MENGE":             RequirementMandatory,
		"ZI_FAKTOR":            RequirementConditional,
		"ZI_EINHEIT":           RequirementConditional,
		"ZI_UST_SCHLUESSEL":    RequirementMandatory,
		"ZI_BASISPREIS_BRUTTO": RequirementConditional,
		"ZI_BASISPREIS_NETTO":  RequirementConditional,
		"ZI_BASISPREIS_UST":    RequirementConditional,
	},
	"transactions_tse.csv": {
		"Z_KASSE_ID":         RequirementMandatory,
		"Z_ERSTELLUNG":       RequirementMandatory,
		"Z_NR":               RequirementMandatory,
		"BON_ID":             RequirementMandatory,
		"TSE_ID":             RequirementMandatory,
		"TSE_TANR":           RequirementMandatory,
		"TSE_TA_START":       RequirementMandatory,
		"TSE_TA_ENDE":        RequirementMandatory,
		"TSE_TA_VORGANGSART": RequirementMandatory,
		"TSE_TA_SIGZ":        RequirementMandatory,
		"TSE_TA_SIG":         RequirementMandatory,
		"TSE_TA_FEHLER":      RequirementConditional,
		"TSE_VORGANGSDATEN":  RequirementConditional,
	},
	"transactions_vat.csv": {
		"Z_KASSE_ID":     RequirementMandatory,
		"Z_ERSTELLUNG":   RequirementMandatory,
		"Z_NR":           RequirementMandatory,
		"BON_ID":         RequirementMandatory,
		"UST_SCHLUESSEL": RequirementMandatory,
		"BON_BRUTTO":     RequirementMandatory,
		"BON_NETTO":      RequirementMandatory,
		"BON_UST":        RequirementMandatory,
	},
	"lines_vat.csv": {
		"Z_KASSE_ID":     RequirementMandatory,
		"Z_ERSTELLUNG":   RequirementMandatory,
		"Z_NR":           RequirementMandatory,
		"BON_ID":         RequirementMandatory,
		"POS_ZEILE":      RequirementMandatory,
		"UST_SCHLUESSEL": RequirementMandatory,
		"POS_BRUTTO":     RequirementConditional,
		"POS_NETTO":      RequirementConditional,
		"POS_UST":        RequirementConditional,
	},
	"allocation_groups.csv": {
		"Z_KASSE_ID":       RequirementMandatory,
		"Z_ERSTELLUNG":     RequirementMandatory,
		"Z_NR":             RequirementMandatory,
		"BON_ID":           RequirementMandatory,
		"ABRECHNUNGSKREIS": RequirementConditional,
	},
	"references.csv": {
		"Z_KASSE_ID":     RequirementMandatory,
		"Z_ERSTELLUNG":   RequirementMandatory,
		"Z_NR":           RequirementMandatory,
		"BON_ID":         RequirementMandatory,
		"POS_ZEILE":      RequirementConditional,
		"REF_TYP":        RequirementMandatory,
		"REF_NAME":       RequirementConditional,
		"REF_DATUM":      RequirementConditional,
		"REF_Z_KASSE_ID": RequirementConditional,
		"REF_Z_NR":       RequirementConditional,
		"REF_BON_ID":     RequirementMandatory,
	},
}
