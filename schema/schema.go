package schema

import "slices"

// SpecVersion is the DSFinV-K version this package describes.
const SpecVersion = "2.4"

// Spec 2.4 p.38: 2.4 has no changes of substance, so 2.3 stays writable.
var currentTaxonomyVersions = []string{"2.3", "2.4"}

// Spec 2.4 p.1 to p.2, Aenderungsnachweis: every published version.
var knownTaxonomyVersions = []string{"1.0", "1.1", "2.0", "2.1", "2.2", "2.3", "2.4"}

// CurrentTaxonomyVersions returns the TAXONOMIE_VERSION values a new export may carry.
func CurrentTaxonomyVersions() []string { return slices.Clone(currentTaxonomyVersions) }

// KnownTaxonomyVersions returns every published TAXONOMIE_VERSION, for reading archives.
func KnownTaxonomyVersions() []string { return slices.Clone(knownTaxonomyVersions) }

var tablesByFile = func() map[string]int {
	m := make(map[string]int, len(tables))
	for i, t := range tables {
		m[t.File] = i
	}
	return m
}()

var files = func() []string {
	out := make([]string, len(tables))
	for i, t := range tables {
		out[i] = t.File
	}
	return out
}()

// clone returns t with a copy of its Columns slice.
func (t Table) clone() Table {
	t.Columns = slices.Clone(t.Columns)
	return t
}

// Tables returns the DSFinV-K tables in index.xml order, as a deep copy.
func Tables() []Table {
	out := make([]Table, len(tables))
	for i, t := range tables {
		out[i] = t.clone()
	}
	return out
}

// TableByFile returns a copy of the table written to the given CSV file name.
func TableByFile(file string) (Table, bool) {
	i, ok := tablesByFile[file]
	if !ok {
		return Table{}, false
	}
	return tables[i].clone(), true
}

// Files returns the CSV file names of all tables in index.xml order.
func Files() []string {
	return slices.Clone(files)
}

// ColumnIndex returns the 0-based position of the named column.
func (t Table) ColumnIndex(name string) (int, bool) {
	for i, c := range t.Columns {
		if c.Name == name {
			return i, true
		}
	}
	return 0, false
}

// Column returns the named column.
func (t Table) Column(name string) (Column, bool) {
	i, ok := t.ColumnIndex(name)
	if !ok {
		return Column{}, false
	}
	return t.Columns[i], true
}
