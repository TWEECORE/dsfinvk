package csvio

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/tweecore/dsfinvk/schema"
)

// ErrHeader reports a header row that does not match the table definition.
var ErrHeader = errors.New("header does not match the schema")

// bom is the UTF-8 byte order mark a foreign export may carry.
const bom = "\ufeff"

// Reader reads the CSV records of one DSFinV-K table.
type Reader struct {
	csv   *csv.Reader
	table schema.Table
}

// NewReader returns a Reader for t, skipping an optional BOM and checking the header.
func NewReader(r io.Reader, t schema.Table) (*Reader, error) {
	table := cloneTable(t)

	br := bufio.NewReader(r)
	if prefix, err := br.Peek(len(bom)); err == nil && string(prefix) == bom {
		if _, err = br.Discard(len(bom)); err != nil {
			return nil, fmt.Errorf("csvio: %s: %w", table.File, err)
		}
	}

	cr := csv.NewReader(br)
	cr.Comma = schema.ColumnDelimiter
	cr.LazyQuotes = false
	cr.TrimLeadingSpace = false
	cr.ReuseRecord = false
	cr.FieldsPerRecord = -1

	rd := &Reader{csv: cr, table: table}
	if err := rd.readHeader(); err != nil {
		return nil, err
	}
	cr.FieldsPerRecord = len(table.Columns)
	return rd, nil
}

// Table returns the table the Reader was built for.
func (r *Reader) Table() schema.Table { return cloneTable(r.table) }

func (r *Reader) readHeader() error {
	header, err := r.csv.Read()
	if err != nil {
		return fmt.Errorf("csvio: %s: %w: cannot read the header row: %w", r.table.File, ErrHeader, err)
	}

	line := r.line()
	if len(header) != len(r.table.Columns) {
		return fmt.Errorf("csvio: %s: %w: line %d has %d columns, want %d",
			r.table.File, ErrHeader, line, len(header), len(r.table.Columns))
	}
	for i, c := range r.table.Columns {
		if header[i] != c.Name {
			return fmt.Errorf("csvio: %s: %w: line %d position %d: want %q, got %q",
				r.table.File, ErrHeader, line, i+1, c.Name, header[i])
		}
	}
	return nil
}

// Read returns the next record and its 1-based physical line number.
func (r *Reader) Read() ([]string, int, error) {
	record, err := r.csv.Read()
	if err == nil {
		return record, r.line(), nil
	}
	if errors.Is(err, io.EOF) {
		return nil, 0, io.EOF
	}

	line := 0
	var parseErr *csv.ParseError
	if errors.As(err, &parseErr) {
		line = parseErr.StartLine
	}
	if errors.Is(err, csv.ErrFieldCount) {
		return nil, line, fmt.Errorf("csvio: %s line %d: %w: %d fields, want %d",
			r.table.File, line, ErrFieldCount, len(record), len(r.table.Columns))
	}
	return nil, line, fmt.Errorf("csvio: %s line %d: %w", r.table.File, line, err)
}

// line is the physical line the most recent record starts on.
func (r *Reader) line() int {
	line, _ := r.csv.FieldPos(0)
	return line
}
