package schema_test

import (
	"slices"
	"testing"

	"github.com/tweecore/dsfinvk/schema"
)

// TestRequirementCounts pins the Anhang G totals: 233 markers dedupe to the 219
// (file, column) pairs of index.xml.
func TestRequirementCounts(t *testing.T) {
	counts := map[schema.Requirement]int{}
	for _, tbl := range schema.Tables() {
		for _, c := range tbl.Columns {
			r, ok := schema.ColumnRequirement(tbl.File, c.Name)
			if !ok {
				t.Errorf("%s.%s has no entry in the Anhang G matrix", tbl.File, c.Name)
				continue
			}
			counts[r]++
		}
	}
	want := map[schema.Requirement]int{
		schema.RequirementMandatory:   151,
		schema.RequirementConditional: 61,
		schema.RequirementOneOf:       4,
		schema.RequirementIfItem:      3,
	}
	total := 0
	for req, n := range want {
		if got := counts[req]; got != n {
			t.Errorf("%v: %d columns, want %d", req, got, n)
		}
		total += n
	}
	if total != 219 {
		t.Errorf("test data sums to %d, want 219", total)
	}
	sum := 0
	for _, n := range counts {
		sum += n
	}
	if sum != 219 {
		t.Errorf("classified %d columns, want 219", sum)
	}
}

func TestEveryColumnIsClassified(t *testing.T) {
	for _, tbl := range schema.Tables() {
		for _, c := range tbl.Columns {
			if _, ok := schema.ColumnRequirement(tbl.File, c.Name); !ok {
				t.Errorf("%s.%s has no entry in the Anhang G matrix", tbl.File, c.Name)
			}
		}
	}
}

func TestNoStaleRequirementEntries(t *testing.T) {
	for _, e := range schema.RequirementEntries() {
		tbl, ok := schema.TableByFile(e.File)
		if !ok {
			t.Errorf("matrix names unknown file %q", e.File)
			continue
		}
		if _, ok := tbl.Column(e.Column); !ok {
			t.Errorf("matrix names unknown column %s.%s", e.File, e.Column)
		}
	}
	if got, want := len(schema.RequirementEntries()), 219; got != want {
		t.Errorf("matrix has %d entries, want %d", got, want)
	}
}

func TestColumnRequirementSpotChecks(t *testing.T) {
	tests := []struct {
		file, column string
		want         schema.Requirement
	}{
		{"cashpointclosing.csv", "Z_BUCHUNGSTAG", schema.RequirementConditional},
		// Anhang G leaves BON_STORNO conditional, not mandatory.
		{"transactions.csv", "BON_STORNO", schema.RequirementConditional},
		{"transactions.csv", "BON_ID", schema.RequirementMandatory},
		{"transactions.csv", "BON_TYP", schema.RequirementMandatory},
		{"transactions.csv", "UMS_BRUTTO", schema.RequirementMandatory},
		{"cashpointclosing.csv", "STNR", schema.RequirementOneOf},
		{"cashpointclosing.csv", "USTID", schema.RequirementOneOf},
		{"pa.csv", "AGENTUR_STNR", schema.RequirementOneOf},
		{"pa.csv", "AGENTUR_USTID", schema.RequirementOneOf},
		{"lines.csv", "ART_NR", schema.RequirementIfItem},
		{"lines.csv", "MENGE", schema.RequirementIfItem},
		{"lines.csv", "STK_BR", schema.RequirementIfItem},
	}
	for _, tc := range tests {
		t.Run(tc.file+"/"+tc.column, func(t *testing.T) {
			got, ok := schema.ColumnRequirement(tc.file, tc.column)
			if !ok {
				t.Fatal("ok = false, want true")
			}
			if got != tc.want {
				t.Errorf("= %v, want %v", got, tc.want)
			}
		})
	}
}

func TestKeyColumnsAreMandatoryEverywhere(t *testing.T) {
	for _, tbl := range schema.Tables() {
		for _, name := range []string{"Z_KASSE_ID", "Z_ERSTELLUNG", "Z_NR"} {
			got, ok := schema.ColumnRequirement(tbl.File, name)
			if !ok || got != schema.RequirementMandatory {
				t.Errorf("%s.%s = (%v, %t), want (RequirementMandatory, true)", tbl.File, name, got, ok)
			}
		}
	}
}

func TestExactIfItemAndOneOfSets(t *testing.T) {
	wantIfItem := map[string]bool{
		"lines.csv/ART_NR": true,
		"lines.csv/MENGE":  true,
		"lines.csv/STK_BR": true,
	}
	wantOneOf := map[string]bool{
		"cashpointclosing.csv/STNR":  true,
		"cashpointclosing.csv/USTID": true,
		"pa.csv/AGENTUR_STNR":        true,
		"pa.csv/AGENTUR_USTID":       true,
	}
	gotIfItem := map[string]bool{}
	gotOneOf := map[string]bool{}
	for _, tbl := range schema.Tables() {
		for _, c := range tbl.Columns {
			r, _ := schema.ColumnRequirement(tbl.File, c.Name)
			switch r {
			case schema.RequirementIfItem:
				gotIfItem[tbl.File+"/"+c.Name] = true
			case schema.RequirementOneOf:
				gotOneOf[tbl.File+"/"+c.Name] = true
			default:
				// Covered by the count tests.
			}
		}
	}
	for k := range wantIfItem {
		if !gotIfItem[k] {
			t.Errorf("%s is not RequirementIfItem", k)
		}
	}
	for k := range gotIfItem {
		if !wantIfItem[k] {
			t.Errorf("unexpected RequirementIfItem: %s", k)
		}
	}
	for k := range wantOneOf {
		if !gotOneOf[k] {
			t.Errorf("%s is not RequirementOneOf", k)
		}
	}
	for k := range gotOneOf {
		if !wantOneOf[k] {
			t.Errorf("unexpected RequirementOneOf: %s", k)
		}
	}
}

func TestColumnRequirementUnknown(t *testing.T) {
	for _, tc := range [][2]string{{"nope.csv", "NOPE"}, {"transactions.csv", "NOPE"}} {
		if got, ok := schema.ColumnRequirement(tc[0], tc[1]); ok {
			t.Errorf("ColumnRequirement(%q, %q) = (%v, true), want ok = false", tc[0], tc[1], got)
		}
	}
}

// TestOneOfGroups pins the Anhang G legend K231: tax_number and vat_id_number
// pair up, at least one of the two must be filled.
func TestOneOfGroups(t *testing.T) {
	tests := map[string][][]string{
		"cashpointclosing.csv": {{"STNR", "USTID"}},
		"pa.csv":               {{"AGENTUR_STNR", "AGENTUR_USTID"}},
	}
	for file, want := range tests {
		got := schema.OneOfGroups(file)
		if len(got) != len(want) {
			t.Fatalf("OneOfGroups(%q) = %v, want %v", file, got, want)
		}
		for i := range want {
			if !slices.Equal(got[i], want[i]) {
				t.Errorf("OneOfGroups(%q)[%d] = %v, want %v", file, i, got[i], want[i])
			}
		}
	}
	for _, tbl := range schema.Tables() {
		if _, paired := tests[tbl.File]; paired {
			continue
		}
		if got := schema.OneOfGroups(tbl.File); len(got) != 0 {
			t.Errorf("OneOfGroups(%q) = %v, want none", tbl.File, got)
		}
	}
	if got := schema.OneOfGroups("nope.csv"); len(got) != 0 {
		t.Errorf(`OneOfGroups("nope.csv") = %v, want none`, got)
	}
}

// TestOneOfGroupsCoverEveryOneOfColumn keeps the groups and the matrix in step.
func TestOneOfGroupsCoverEveryOneOfColumn(t *testing.T) {
	grouped := map[string]bool{}
	for _, tbl := range schema.Tables() {
		for _, group := range schema.OneOfGroups(tbl.File) {
			if len(group) < 2 {
				t.Errorf("%s: group %v has fewer than two columns", tbl.File, group)
			}
			for _, name := range group {
				if _, ok := tbl.Column(name); !ok {
					t.Errorf("%s: group names unknown column %q", tbl.File, name)
				}
				grouped[tbl.File+"/"+name] = true
			}
		}
	}
	for _, tbl := range schema.Tables() {
		for _, c := range tbl.Columns {
			r, _ := schema.ColumnRequirement(tbl.File, c.Name)
			key := tbl.File + "/" + c.Name
			if (r == schema.RequirementOneOf) != grouped[key] {
				t.Errorf("%s: requirement %v but grouped = %t", key, r, grouped[key])
			}
		}
	}
}

func TestOneOfGroupsReturnACopy(t *testing.T) {
	a := schema.OneOfGroups("pa.csv")
	a[0][0] = "TAMPERED"
	if b := schema.OneOfGroups("pa.csv"); b[0][0] != "AGENTUR_STNR" {
		t.Errorf("OneOfGroups returns shared slices: got %q after mutation", b[0][0])
	}
}

func TestRequirementString(t *testing.T) {
	tests := map[schema.Requirement]string{
		schema.RequirementConditional: "Conditional",
		schema.RequirementMandatory:   "Mandatory",
		schema.RequirementIfItem:      "IfItem",
		schema.RequirementOneOf:       "OneOf",
		schema.Requirement(9):         "Requirement(9)",
	}
	for req, want := range tests {
		if got := req.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", uint8(req), got, want)
		}
	}
}
