package csvio

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/tweecore/dsfinvk/schema"
)

var (
	// ErrHeaderWritten reports a second call to WriteHeader.
	ErrHeaderWritten = errors.New("header already written")
	// ErrFieldCount reports a record whose field count differs from the table's.
	ErrFieldCount = errors.New("wrong field count")
)

// specialChars force a field to be encapsulated.
var specialChars = string([]rune{schema.ColumnDelimiter, schema.TextEncapsulator, '\r', '\n'})

// Writer writes the CSV records of one DSFinV-K table.
type Writer struct {
	w      *bufio.Writer
	table  schema.Table
	buf    []byte
	header bool
	err    error
}

// NewWriter returns a Writer for t that emits the DSFinV-K CSV dialect without a BOM.
func NewWriter(w io.Writer, t schema.Table) *Writer {
	return &Writer{w: bufio.NewWriter(w), table: cloneTable(t)}
}

// cloneTable returns t with a copy of its Columns slice.
func cloneTable(t schema.Table) schema.Table {
	t.Columns = slices.Clone(t.Columns)
	return t
}

// Table returns the table the Writer was built for.
func (w *Writer) Table() schema.Table { return cloneTable(w.table) }

// WriteHeader writes the column name row; it may be called only once.
func (w *Writer) WriteHeader() error {
	if w.err != nil {
		return w.err
	}
	if w.header {
		return fmt.Errorf("csvio: %s: %w", w.table.File, ErrHeaderWritten)
	}

	names := make([]string, len(w.table.Columns))
	for i, c := range w.table.Columns {
		names[i] = c.Name
	}
	w.header = true
	return w.writeRecord(names)
}

// Write writes one record; len(fields) must match the table's column count.
func (w *Writer) Write(fields []string) error {
	if w.err != nil {
		return w.err
	}
	if len(fields) != len(w.table.Columns) {
		return fmt.Errorf("csvio: %s: %w: %d fields, want %d", w.table.File, ErrFieldCount, len(fields), len(w.table.Columns))
	}
	if !w.header {
		if err := w.WriteHeader(); err != nil {
			return err
		}
	}
	return w.writeRecord(fields)
}

// Flush writes any buffered output to the underlying writer.
func (w *Writer) Flush() {
	if err := w.w.Flush(); err != nil && w.err == nil {
		w.err = fmt.Errorf("csvio: %s: %w", w.table.File, err)
	}
}

// Error returns the first error the Writer hit.
func (w *Writer) Error() error { return w.err }

func (w *Writer) writeRecord(fields []string) error {
	w.buf = w.buf[:0]
	for i, f := range fields {
		if i > 0 {
			w.buf = append(w.buf, schema.ColumnDelimiter)
		}
		w.buf = appendField(w.buf, f)
	}
	w.buf = append(w.buf, schema.RecordDelimiter...)

	if _, err := w.w.Write(w.buf); err != nil {
		w.err = fmt.Errorf("csvio: %s: %w", w.table.File, err)
		return w.err
	}
	return nil
}

// appendField appends s, encapsulated and with doubled quotes when needed.
func appendField(dst []byte, s string) []byte {
	if !needsQuotes(s) {
		return append(dst, s...)
	}

	dst = append(dst, schema.TextEncapsulator)
	for i := 0; i < len(s); i++ {
		if s[i] == schema.TextEncapsulator {
			dst = append(dst, schema.TextEncapsulator)
		}
		dst = append(dst, s[i])
	}
	return append(dst, schema.TextEncapsulator)
}

func needsQuotes(s string) bool {
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, specialChars) {
		return true
	}
	return isEdgeSpace(s[0]) || isEdgeSpace(s[len(s)-1])
}

func isEdgeSpace(b byte) bool { return b == ' ' || b == '\t' }
