package schema

import (
	"slices"
	"strconv"
)

// VATKeysRevision is the publication date of the Anlage 2 VAT key table.
const VATKeysRevision = "2024-12-05"

// VATKey is one entry of the UST_SCHLUESSEL table. Anlage 2, 2024-12-05.
type VATKey struct {
	ID int
	// Rate is the decimal-comma percentage, empty for the current-rate keys.
	Rate        string
	Description string
	// ValidFrom is the "gueltig ab" date in ISO 8601, empty for the current-rate keys.
	ValidFrom  string
	Historical bool
}

// vatKeys is the UST_SCHLUESSEL table of Anlage 2, in ascending ID order.
var vatKeys = []VATKey{
	{
		ID:          1,
		Description: "Zum Zeitpunkt der Erfassung des Geschäftsvorfalls geltender allgemeiner Steuersatz nach § 12 Abs. 1 UStG",
	},
	{
		ID:          2,
		Description: "Zum Zeitpunkt der Erfassung des Geschäftsvorfalls geltender ermäßigter Steuersatz nach § 12 Abs. 2 UStG",
	},
	{
		ID:          3,
		Description: "Zum Zeitpunkt der Erfassung des Geschäftsvorfalls geltender Durchschnittssatz nach § 24 Abs. 1 Nr. 3 UStG (übrige Fälle)",
	},
	{
		ID:          4,
		Description: "Zum Zeitpunkt der Erfassung des Geschäftsvorfalls geltender Durchschnittssatz nach § 24 Abs. 1 Nr. 1 UStG",
	},
	{
		ID:          5,
		Rate:        "0,00",
		Description: "Nicht Steuerbar",
	},
	{
		ID:          6,
		Rate:        "0,00",
		Description: "Umsatzsteuerfrei",
	},
	{
		ID:          7,
		Rate:        "0,00",
		Description: "UmsatzsteuerNichtErmittelbar",
	},
	{
		ID:          8,
		Description: "Zum Zeitpunkt der Erfassung des Geschäftsvorfalls geltender ermäßigter Steuersatz nach § 12 Abs. 3 UStG",
	},
	{
		ID:          11,
		Rate:        "19,00",
		Description: "Historischer allgemeiner Steuersatz nach § 12 Abs. 1 UStG",
		ValidFrom:   "2020-07-01",
		Historical:  true,
	},
	{
		ID:          12,
		Rate:        "7,00",
		Description: "Historischer ermäßigter Steuersatz nach § 12 Abs. 2 UStG",
		ValidFrom:   "2020-07-01",
		Historical:  true,
	},
	{
		ID:          13,
		Rate:        "10,70",
		Description: "Historischer Durchschnittssatz nach § 24 Abs. 1 Nr. 3 UStG",
		ValidFrom:   "2022-01-01",
		Historical:  true,
	},
	{
		ID:          21,
		Rate:        "16,00",
		Description: "Historischer allgemeiner Steuersatz nach § 12 Abs. 1 UStG",
		ValidFrom:   "2021-01-01",
		Historical:  true,
	},
	{
		ID:          22,
		Rate:        "5,00",
		Description: "Historischer ermäßigter Steuersatz nach § 12 Abs. 2 UStG",
		ValidFrom:   "2021-01-01",
		Historical:  true,
	},
	{
		ID:          23,
		Rate:        "9,50",
		Description: "Historischer Durchschnittssatz nach § 24 Abs. 1 Nr. 3 UStG",
		ValidFrom:   "2023-01-01",
		Historical:  true,
	},
	{
		ID:          33,
		Rate:        "9,00",
		Description: "Historischer Durchschnittssatz nach § 24 Abs. 1 Nr. 3 UStG",
		ValidFrom:   "2024-12-06",
		Historical:  true,
	},
	{
		ID:          43,
		Rate:        "8,40",
		Description: "Historischer Durchschnittssatz nach § 24 Abs. 1 Nr. 3 UStG",
		ValidFrom:   "2025-01-01",
		Historical:  true,
	},
}

// VATKeys returns a copy of the UST_SCHLUESSEL table, in ascending ID order.
func VATKeys() []VATKey { return slices.Clone(vatKeys) }

var vatKeysByID = func() map[int]VATKey {
	m := make(map[int]VATKey, len(vatKeys))
	for _, k := range vatKeys {
		m[k.ID] = k
	}
	return m
}()

// VATKeyByID returns the Anlage 2 entry for the given UST_SCHLUESSEL.
func VATKeyByID(id int) (VATKey, bool) {
	k, ok := vatKeysByID[id]
	return k, ok
}

// VATKeyClass says which Anlage 2 range an ID falls in: up to 999 reserved, 1000+ individual.
type VATKeyClass uint8

const (
	VATKeyDefined VATKeyClass = iota
	VATKeyReserved
	VATKeyIndividual
	// VATKeyInvalid covers zero and negative IDs; Anlage 2 counts from 1.
	VATKeyInvalid
)

// String returns the name of the class.
func (c VATKeyClass) String() string {
	switch c {
	case VATKeyDefined:
		return "Defined"
	case VATKeyReserved:
		return "Reserved"
	case VATKeyIndividual:
		return "Individual"
	case VATKeyInvalid:
		return "Invalid"
	default:
		return "VATKeyClass(" + strconv.Itoa(int(c)) + ")"
	}
}

// ClassifyVATKey reports which range of Anlage 2 the given UST_SCHLUESSEL falls into.
func ClassifyVATKey(id int) VATKeyClass {
	switch {
	case id <= 0:
		return VATKeyInvalid
	case isDefinedVATKey(id):
		return VATKeyDefined
	case id >= individualVATKeyFrom:
		return VATKeyIndividual
	default:
		return VATKeyReserved
	}
}

const individualVATKeyFrom = 1000

func isDefinedVATKey(id int) bool {
	_, ok := vatKeysByID[id]
	return ok
}

// Spec 2.4 p.114, Anlage 2 p.2: the tax container of each base key, index 0 unused.
var baseVATKeyTaxSlots = [9]int{0, 1, 2, 3, 4, 5, 5, 5, 5}

// VATKeyTaxSlot returns the 1-based TSE processData tax container of an
// UST_SCHLUESSEL. Spec 2.4 p.26: a historical key is two digits, the first a
// sequence number and the second a reference to the original ID.
func VATKeyTaxSlot(id int) (int, bool) {
	base := id
	if id >= 10 && id <= 99 {
		base = id % 10
	}
	if base < 1 || base >= len(baseVATKeyTaxSlots) {
		return 0, false
	}
	return baseVATKeyTaxSlots[base], true
}
