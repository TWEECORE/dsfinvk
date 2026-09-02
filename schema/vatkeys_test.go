package schema_test

import (
	"regexp"
	"slices"
	"testing"

	"github.com/tweecore/dsfinvk/schema"
)

func TestVATKeysRevision(t *testing.T) {
	if got, want := schema.VATKeysRevision, "2024-12-05"; got != want {
		t.Errorf("VATKeysRevision = %q, want %q", got, want)
	}
	if schema.VATKeysRevision == schema.SpecVersion {
		t.Error("VATKeysRevision must be tracked separately from SpecVersion")
	}
}

func TestVATKeyIDs(t *testing.T) {
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 11, 12, 13, 21, 22, 23, 33, 43}
	got := make([]int, len(schema.VATKeys()))
	for i, k := range schema.VATKeys() {
		got[i] = k.ID
	}
	if !slices.Equal(got, want) {
		t.Errorf("VAT key IDs =\n%v\nwant\n%v", got, want)
	}
	if len(schema.VATKeys()) != 16 {
		t.Errorf("%d defined keys, want 16", len(schema.VATKeys()))
	}
}

func TestVATKeyByID(t *testing.T) {
	for _, k := range schema.VATKeys() {
		got, ok := schema.VATKeyByID(k.ID)
		if !ok {
			t.Errorf("VATKeyByID(%d): ok = false", k.ID)
			continue
		}
		if got != k {
			t.Errorf("VATKeyByID(%d) = %+v, want %+v", k.ID, got, k)
		}
	}
	for _, id := range []int{0, -1, 9, 10, 14, 99, 1000} {
		if _, ok := schema.VATKeyByID(id); ok {
			t.Errorf("VATKeyByID(%d): ok = true", id)
		}
	}
}

func TestVATKeyRates(t *testing.T) {
	tests := map[int]string{
		// Keys 1-4 and 8 mean the rate in force at recording time; Anlage 2 states none.
		1: "", 2: "", 3: "", 4: "", 8: "",
		5: "0,00", 6: "0,00", 7: "0,00",
		11: "19,00", 12: "7,00", 13: "10,70",
		21: "16,00", 22: "5,00", 23: "9,50",
		33: "9,00", 43: "8,40",
	}
	for id, want := range tests {
		k, ok := schema.VATKeyByID(id)
		if !ok {
			t.Errorf("VATKeyByID(%d) not found", id)
			continue
		}
		if k.Rate != want {
			t.Errorf("VATKeyByID(%d).Rate = %q, want %q", id, k.Rate, want)
		}
	}
	re := regexp.MustCompile(`^\d{1,2},\d{2}$`)
	for _, k := range schema.VATKeys() {
		if k.Rate != "" && !re.MatchString(k.Rate) {
			t.Errorf("key %d: Rate %q is not of the form NN,NN", k.ID, k.Rate)
		}
	}
}

func TestVATKeyHistoricalAndValidFrom(t *testing.T) {
	wantValidFrom := map[int]string{
		1: "", 2: "", 3: "", 4: "", 5: "", 6: "", 7: "", 8: "",
		11: "2020-07-01", 12: "2020-07-01", 13: "2022-01-01",
		21: "2021-01-01", 22: "2021-01-01", 23: "2023-01-01",
		33: "2024-12-06", 43: "2025-01-01",
	}
	for _, k := range schema.VATKeys() {
		if got, want := k.ValidFrom, wantValidFrom[k.ID]; got != want {
			t.Errorf("key %d: ValidFrom = %q, want %q", k.ID, got, want)
		}
		wantHistorical := wantValidFrom[k.ID] != ""
		if k.Historical != wantHistorical {
			t.Errorf("key %d: Historical = %v, want %v", k.ID, k.Historical, wantHistorical)
		}
	}
	n := 0
	for _, k := range schema.VATKeys() {
		if k.Historical {
			n++
		}
	}
	if n != 8 {
		t.Errorf("%d historical keys, want 8", n)
	}
}

func TestVATKeyDescriptions(t *testing.T) {
	tests := map[int]string{
		5:  "Nicht Steuerbar",
		6:  "Umsatzsteuerfrei",
		7:  "UmsatzsteuerNichtErmittelbar",
		11: "Historischer allgemeiner Steuersatz nach § 12 Abs. 1 UStG",
		43: "Historischer Durchschnittssatz nach § 24 Abs. 1 Nr. 3 UStG",
	}
	for id, want := range tests {
		k, ok := schema.VATKeyByID(id)
		if !ok {
			t.Errorf("VATKeyByID(%d) not found", id)
			continue
		}
		if k.Description != want {
			t.Errorf("key %d: Description = %q, want %q", id, k.Description, want)
		}
	}
	for _, k := range schema.VATKeys() {
		if k.Description == "" {
			t.Errorf("key %d has no description", k.ID)
		}
	}
	tbl, ok := schema.TableByFile("vat.csv")
	if !ok {
		t.Fatal("vat.csv not found")
	}
	col, ok := tbl.Column("UST_BESCHR")
	if !ok {
		t.Fatal("UST_BESCHR not found")
	}
	for _, k := range schema.VATKeys() {
		if n := len([]rune(k.Description)); n > col.MaxLength {
			t.Logf("key %d: description is %d characters, UST_BESCHR allows %d", k.ID, n, col.MaxLength)
		}
	}
}

func TestClassifyVATKey(t *testing.T) {
	for _, k := range schema.VATKeys() {
		if got := schema.ClassifyVATKey(k.ID); got != schema.VATKeyDefined {
			t.Errorf("ClassifyVATKey(%d) = %v, want VATKeyDefined", k.ID, got)
		}
	}
	for _, id := range []int{9, 10, 14, 20, 100, 999} {
		if got := schema.ClassifyVATKey(id); got != schema.VATKeyReserved {
			t.Errorf("ClassifyVATKey(%d) = %v, want VATKeyReserved", id, got)
		}
	}
	for _, id := range []int{1000, 1001, 5000} {
		if got := schema.ClassifyVATKey(id); got != schema.VATKeyIndividual {
			t.Errorf("ClassifyVATKey(%d) = %v, want VATKeyIndividual", id, got)
		}
	}
	// Anlage 2 counts from 1, so zero and negative IDs are not a range at all.
	for _, id := range []int{0, -1, -42} {
		if got := schema.ClassifyVATKey(id); got != schema.VATKeyInvalid {
			t.Errorf("ClassifyVATKey(%d) = %v, want VATKeyInvalid", id, got)
		}
	}
}

func TestVATKeyClassString(t *testing.T) {
	tests := map[schema.VATKeyClass]string{
		schema.VATKeyDefined:    "Defined",
		schema.VATKeyReserved:   "Reserved",
		schema.VATKeyIndividual: "Individual",
		schema.VATKeyInvalid:    "Invalid",
		schema.VATKeyClass(9):   "VATKeyClass(9)",
	}
	for c, want := range tests {
		if got := c.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", uint8(c), got, want)
		}
	}
}

func TestVATKeyTaxSlot(t *testing.T) {
	want := map[int]int{
		1: 1, 11: 1, 21: 1,
		2: 2, 12: 2, 22: 2,
		3: 3, 13: 3, 23: 3, 33: 3, 43: 3,
		4: 4,
		5: 5, 6: 5, 7: 5, 8: 5,
	}
	if len(want) != len(schema.VATKeys()) {
		t.Fatalf("test data covers %d keys, VATKeys has %d", len(want), len(schema.VATKeys()))
	}
	for _, k := range schema.VATKeys() {
		wantSlot, ok := want[k.ID]
		if !ok {
			t.Errorf("key %d has no expected slot", k.ID)
			continue
		}
		got, ok := schema.VATKeyTaxSlot(k.ID)
		if !ok {
			t.Errorf("VATKeyTaxSlot(%d): ok = false", k.ID)
			continue
		}
		if got != wantSlot {
			t.Errorf("VATKeyTaxSlot(%d) = %d, want %d", k.ID, got, wantSlot)
		}
	}
	seen := map[int]bool{}
	for _, k := range schema.VATKeys() {
		slot, _ := schema.VATKeyTaxSlot(k.ID)
		seen[slot] = true
	}
	for slot := 1; slot <= 5; slot++ {
		if !seen[slot] {
			t.Errorf("no VAT key maps to tax slot %d", slot)
		}
	}
	if len(seen) != 5 {
		t.Errorf("%d distinct slots, want 5", len(seen))
	}
}

// TestVATKeyTaxSlotIsGenerative pins the rule of spec p.26: a historical key is
// two digits, the second of which references the original ID.
func TestVATKeyTaxSlotIsGenerative(t *testing.T) {
	tests := map[int]int{
		53: 3, 51: 1, 62: 2, 74: 4, 95: 5, 96: 5, 97: 5, 98: 5,
		14: 4, 18: 5,
	}
	for id, wantSlot := range tests {
		got, ok := schema.VATKeyTaxSlot(id)
		if !ok {
			t.Errorf("VATKeyTaxSlot(%d): ok = false, want slot %d", id, wantSlot)
			continue
		}
		if got != wantSlot {
			t.Errorf("VATKeyTaxSlot(%d) = %d, want %d", id, got, wantSlot)
		}
	}
	for _, id := range []int{0, -1, 9, 10, 19, 20, 90, 99, 100, 109, 999, 1000, 1001} {
		if slot, ok := schema.VATKeyTaxSlot(id); ok {
			t.Errorf("VATKeyTaxSlot(%d) = %d, true; want ok = false", id, slot)
		}
	}
}

func TestVATKeysAreSortedByID(t *testing.T) {
	keys := schema.VATKeys()
	for i := 1; i < len(keys); i++ {
		if keys[i-1].ID >= keys[i].ID {
			t.Errorf("VATKeys() is not strictly ascending at %d: %d then %d", i, keys[i-1].ID, keys[i].ID)
		}
	}
}

func TestVATKeysReturnsACopy(t *testing.T) {
	a := schema.VATKeys()
	a[0].Description = "tampered"
	if b := schema.VATKeys(); b[0].Description == "tampered" {
		t.Error("VATKeys() returns a shared slice")
	}
}
