package schema

import (
	"errors"
	"fmt"
	"slices"
)

var (
	// ErrDuplicateColumn reports an extra column whose name a table already carries.
	ErrDuplicateColumn = errors.New("duplicate column name")
	// ErrEmptyColumnName reports an extra column without a name.
	ErrEmptyColumnName = errors.New("empty column name")
)

// WithExtraColumns returns a copy of t with cols appended. Spec 2.4 p.10, p.79.
func (t Table) WithExtraColumns(cols ...Column) (Table, error) {
	out := t.clone()
	seen := make(map[string]struct{}, len(out.Columns)+len(cols))
	for _, c := range out.Columns {
		seen[c.Name] = struct{}{}
	}

	for _, c := range cols {
		if c.Name == "" {
			return Table{}, fmt.Errorf("schema: table %s: %w", t.File, ErrEmptyColumnName)
		}
		if _, dup := seen[c.Name]; dup {
			return Table{}, fmt.Errorf("schema: table %s: %w %q", t.File, ErrDuplicateColumn, c.Name)
		}
		seen[c.Name] = struct{}{}
		out.Columns = append(out.Columns, c)
	}
	out.Columns = slices.Clip(out.Columns)
	return out, nil
}
