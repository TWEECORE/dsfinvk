package model

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/csvio"
	"github.com/tweecore/dsfinvk/schema"
)

var (
	// ErrKasseID reports a Kassenabschluss without a Z_KASSE_ID.
	ErrKasseID = errors.New("empty Z_KASSE_ID")
	// ErrClosingNr reports a Z_NR below 1.
	ErrClosingNr = errors.New("Z_NR must be 1 or greater")
	// ErrEnumValue reports a value outside the enumeration the specification defines.
	ErrEnumValue = errors.New("value outside the specification enumeration")
	// ErrColumn reports a column name the table does not carry.
	ErrColumn = errors.New("unknown column")
	// ErrAgenturID reports a declared Agentur whose AGENTUR_ID is below 1.
	ErrAgenturID = errors.New("AGENTUR_ID must be 1 or greater")
	// ErrDuplicateBon reports a BON_ID used twice in one Kassenabschluss.
	ErrDuplicateBon = errors.New("duplicate BON_ID in one Kassenabschluss")
	// ErrDuplicatePosition reports a POS_ZEILE used twice in one Bon.
	ErrDuplicatePosition = errors.New("duplicate POS_ZEILE in one Bon")
	// ErrUnknownAgentur reports an AGENTUR_ID the Kassenabschluss does not declare.
	ErrUnknownAgentur = errors.New("AGENTUR_ID not declared in Agenturen")
	// ErrUnknownUSt reports a UST_SCHLUESSEL the Kassenabschluss does not declare.
	ErrUnknownUSt = errors.New("UST_SCHLUESSEL not declared in USt")
	// ErrUnknownTSE reports a TSE_ID the Kassenabschluss does not declare.
	ErrUnknownTSE = errors.New("TSE_ID not declared in TSEs")
)

// defaultTaxonomieVersion is written when a Kassenabschluss names none.
const defaultTaxonomieVersion = schema.SpecVersion

// tseTimeLayout is the TSE log time format, always UTC. Spec 2.4 p.104.
const tseTimeLayout = "2006-01-02T15:04:05.000Z"

// The TSE certificate is split over columns of 1000 characters, of which
// index.xml declares two. Spec 2.4 p.79.
const (
	certChunkLen   = 1000
	certBaseColumn = 2
	certPrefix     = "TSE_ZERTIFIKAT_"
	certDescr      = "Ggf. Rest des base64-codierten Zertifikats der TSE (in base64-Codierung)"
)

// Rows are the data rows of an export, keyed by CSV file name, in schema column order.
type Rows map[string][][]string

// Build renders e to the rows of every table and returns the table set to write.
func Build(e Export) (Rows, []schema.Table, error) {
	tables, err := exportTables(e)
	if err != nil {
		return nil, nil, err
	}

	b := newBuilder(tables)
	for _, c := range e.Abschluesse {
		if err := b.closing(c); err != nil {
			return nil, nil, err
		}
	}
	if b.err != nil {
		return nil, nil, b.err
	}
	return b.rows, tables, nil
}

// builder renders Kassenabschluesse into Rows.
type builder struct {
	tables map[string]schema.Table
	rows   Rows
	err    error

	kasseID       string
	erstellung    string
	nr            string
	basiswaehrung string
	agenturen     map[int64]struct{}
	ustKeys       map[int64]struct{}
	tseIDs        map[int64]struct{}
}

func newBuilder(tables []schema.Table) *builder {
	b := &builder{
		tables: make(map[string]schema.Table, len(tables)),
		rows:   make(Rows, len(tables)),
	}
	for _, t := range tables {
		b.tables[t.File] = t
		b.rows[t.File] = nil
	}
	return b
}

// keep records the first error the builder hit.
func (b *builder) keep(err error) {
	if err != nil && b.err == nil {
		b.err = err
	}
}

// closing renders one Kassenabschluss and its Bons.
func (b *builder) closing(c Kassenabschluss) error {
	if c.KasseID == "" {
		return fmt.Errorf("model: Kassenabschluss %d: %w", c.Nr, ErrKasseID)
	}
	if c.Nr < 1 {
		return fmt.Errorf("model: Kassenabschluss %s: %w: %d", c.KasseID, ErrClosingNr, c.Nr)
	}

	sums, err := aggregateClosing(c)
	if err != nil {
		return err
	}

	b.kasseID = c.KasseID
	b.erstellung = stamp(c.Erstellung)
	b.nr = strconv.FormatInt(c.Nr, 10)
	b.basiswaehrung = c.Kasse.Basiswaehrung
	b.index(c)

	b.closingRow(c, sums.totals)
	b.locationRow(c.Standort)
	b.cashregisterRow(c.Kasse)
	for _, t := range c.Terminals {
		b.terminalRow(t)
	}
	for _, a := range c.Agenturen {
		if a.ID < 1 {
			return fmt.Errorf("model: Kassenabschluss %s %s: %w: %d", c.KasseID, b.nr, ErrAgenturID, a.ID)
		}
		b.agenturRow(a)
	}
	for _, t := range c.TSEs {
		b.tseRow(t)
	}
	for _, u := range c.USt {
		b.vatRow(u)
	}
	for _, k := range sums.gvOrder {
		b.businesscaseRow(k, sums.gv[k])
	}
	for _, k := range sums.payOrder {
		b.paymentRow(k, sums.pay[k])
	}
	for _, currency := range sums.curOrder {
		b.currencyRow(currency, sums.cur[currency])
	}

	seen := make(map[string]struct{}, len(c.Bons))
	for _, bon := range c.Bons {
		if _, dup := seen[bon.ID]; dup {
			return fmt.Errorf("model: Kassenabschluss %s %s: %w: %q", c.KasseID, b.nr, ErrDuplicateBon, bon.ID)
		}
		seen[bon.ID] = struct{}{}
		if err := b.bon(bon); err != nil {
			return err
		}
	}
	return nil
}

// index records the keys the Bons of one Kassenabschluss may reference.
func (b *builder) index(c Kassenabschluss) {
	b.agenturen = make(map[int64]struct{}, len(c.Agenturen))
	for _, a := range c.Agenturen {
		b.agenturen[a.ID] = struct{}{}
	}
	b.ustKeys = make(map[int64]struct{}, len(c.USt))
	for _, u := range c.USt {
		b.ustKeys[u.Schluessel] = struct{}{}
	}
	b.tseIDs = make(map[int64]struct{}, len(c.TSEs))
	for _, t := range c.TSEs {
		b.tseIDs[t.ID] = struct{}{}
	}
}

// bon renders one Vorgang and its child rows.
func (b *builder) bon(bon Bon) error {
	b.bonRow(bon)
	for _, kreis := range bon.Abrechnungskreise {
		b.allocationRow(bon.ID, kreis)
	}
	for _, z := range bon.Zahlungen {
		b.zahlungRow(bon.ID, z)
	}
	for _, u := range bon.USt {
		b.bonUStRow(bon.ID, u)
	}
	for _, ref := range bon.Referenzen {
		b.referenzRow(bon.ID, ref)
	}
	if bon.TSE != nil {
		b.tseTransaktionRow(bon.ID, *bon.TSE)
	}

	seen := make(map[string]struct{}, len(bon.Positionen))
	for _, p := range bon.Positionen {
		if _, dup := seen[p.Zeile]; dup {
			return fmt.Errorf("model: Bon %s: %w: %q", bon.ID, ErrDuplicatePosition, p.Zeile)
		}
		seen[p.Zeile] = struct{}{}
		b.position(bon.ID, p)
	}
	return nil
}

// position renders one line and its child rows.
func (b *builder) position(bonID string, p Position) {
	b.positionRow(bonID, p)
	for _, u := range p.USt {
		b.posUStRow(bonID, p.Zeile, u)
	}
	for _, pf := range p.Preisfindung {
		b.preisfindungRow(bonID, p.Zeile, pf)
	}
	for _, z := range p.Zusatzinfos {
		b.zusatzinfoRow(bonID, p.Zeile, z)
	}
}

// checkUSt reports a UST_SCHLUESSEL the Kassenabschluss does not declare.
func (b *builder) checkUSt(schluessel int64) {
	if _, ok := b.ustKeys[schluessel]; !ok {
		b.keep(fmt.Errorf("model: Kassenabschluss %s %s: %w: %d", b.kasseID, b.nr, ErrUnknownUSt, schluessel))
	}
}

// checkAgentur reports an AGENTUR_ID the Kassenabschluss does not declare; 0 is the own company.
func (b *builder) checkAgentur(id int64) {
	if id == 0 {
		return
	}
	if _, ok := b.agenturen[id]; !ok {
		b.keep(fmt.Errorf("model: Kassenabschluss %s %s: %w: %d", b.kasseID, b.nr, ErrUnknownAgentur, id))
	}
}

// checkTSE reports a TSE_ID the Kassenabschluss does not declare.
func (b *builder) checkTSE(id int64) {
	if _, ok := b.tseIDs[id]; !ok {
		b.keep(fmt.Errorf("model: Kassenabschluss %s %s: %w: %d", b.kasseID, b.nr, ErrUnknownTSE, id))
	}
}

// exportTables returns the table set of e, with tse.csv extended when a
// certificate needs more than the two declared columns. Spec 2.4 p.79.
func exportTables(e Export) ([]schema.Table, error) {
	tables := schema.Tables()

	extra := 0
	for _, c := range e.Abschluesse {
		for _, t := range c.TSEs {
			if n := len(certChunks(t.Zertifikat)) - certBaseColumn; n > extra {
				extra = n
			}
		}
	}
	if extra == 0 {
		return tables, nil
	}

	cols := make([]schema.Column, extra)
	for i := range cols {
		cols[i] = schema.Column{
			Name:        certColumn(certBaseColumn + i + 1),
			Description: certDescr,
			Type:        schema.ColumnAlphaNumeric,
			MaxLength:   certChunkLen,
		}
	}
	for i, t := range tables {
		if t.File != "tse.csv" {
			continue
		}
		ext, err := t.WithExtraColumns(cols...)
		if err != nil {
			return nil, fmt.Errorf("model: %w", err)
		}
		tables[i] = ext
	}
	return tables, nil
}

// certChunks splits a certificate into runs of certChunkLen characters.
func certChunks(cert string) []string {
	runes := []rune(cert)
	if len(runes) == 0 {
		return nil
	}

	out := make([]string, 0, (len(runes)+certChunkLen-1)/certChunkLen)
	for len(runes) > certChunkLen {
		out = append(out, string(runes[:certChunkLen]))
		runes = runes[certChunkLen:]
	}
	return append(out, string(runes))
}

// certColumn is the name of the n-th certificate column, counted from one.
func certColumn(n int) string { return certPrefix + roman(n) }

// romanNumerals are the additive and subtractive pairs of the Roman system.
var romanNumerals = []struct {
	value  int
	symbol string
}{
	{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
	{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
	{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
}

// roman renders a positive integer as a Roman numeral.
func roman(n int) string {
	var out []byte
	for _, r := range romanNumerals {
		for n >= r.value {
			out = append(out, r.symbol...)
			n -= r.value
		}
	}
	return string(out)
}

// stamp renders a timestamp as RFC 3339 with offset; a zero time is empty. Spec 2.4 p.64.
func stamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// rec starts a row of file, prefilled with the key columns of the closing.
func (b *builder) rec(file string) *record {
	t := b.tables[file]
	r := &record{b: b, table: t, fields: make([]string, len(t.Columns))}
	r.text("Z_KASSE_ID", b.kasseID)
	r.text("Z_ERSTELLUNG", b.erstellung)
	r.set("Z_NR", b.nr)
	return r
}

// add appends a finished row to its table.
func (b *builder) add(r *record) {
	b.rows[r.table.File] = append(b.rows[r.table.File], r.fields)
}

// record is one row under construction, addressed by column name.
type record struct {
	b      *builder
	table  schema.Table
	fields []string
}

// set writes a raw value into the named column.
func (r *record) set(column, value string) {
	i, ok := r.table.ColumnIndex(column)
	if !ok {
		r.b.keep(fmt.Errorf("model: %s: %w %q", r.table.File, ErrColumn, column))
		return
	}
	r.fields[i] = value
}

// column returns the named column definition.
func (r *record) column(name string) (schema.Column, bool) {
	c, ok := r.table.Column(name)
	if !ok {
		r.b.keep(fmt.Errorf("model: %s: %w %q", r.table.File, ErrColumn, name))
	}
	return c, ok
}

// text writes a string value and checks its length.
func (r *record) text(column, value string) {
	c, ok := r.column(column)
	if !ok {
		return
	}
	if err := csvio.CheckLength(c, value); err != nil {
		r.b.keep(fmt.Errorf("model: %s: %w", r.table.File, err))
		return
	}
	r.set(column, value)
}

// enum writes a string value that must be one of allowed.
func (r *record) enum(column, value string, allowed []string) {
	for _, a := range allowed {
		if a == value {
			r.text(column, value)
			return
		}
	}
	r.b.keep(fmt.Errorf("model: %s: %s: %w: %q", r.table.File, column, ErrEnumValue, value))
}

// num writes a decimal in the accuracy of its column.
func (r *record) num(column string, v dsfinvk.Decimal) {
	c, ok := r.column(column)
	if !ok {
		return
	}
	s, err := csvio.FormatValue(c, v)
	if err != nil {
		r.b.keep(fmt.Errorf("model: %s: %w", r.table.File, err))
		return
	}
	r.set(column, s)
}

// numOpt writes a decimal, leaving the column empty when the value is zero.
func (r *record) numOpt(column string, v dsfinvk.Decimal) {
	if v.IsZero() {
		return
	}
	r.num(column, v)
}

// id writes an integer key column.
func (r *record) id(column string, v int64) {
	r.set(column, strconv.FormatInt(v, 10))
}

// flag writes a boolean column as 0 or 1. Anhang G legend K228.
func (r *record) flag(column string, v bool) {
	if v {
		r.set(column, schema.Bool01True)
		return
	}
	r.set(column, schema.Bool01False)
}

// stamp writes a timestamp as RFC 3339 with offset; a zero time stays empty.
func (r *record) stamp(column string, t time.Time) {
	r.text(column, stamp(t))
}

// day writes a date; a zero time stays empty. Spec 2.4 p.66.
func (r *record) day(column string, t time.Time) {
	if t.IsZero() {
		return
	}
	r.text(column, t.Format(time.DateOnly))
}

// tseStamp writes a TSE log time as UTC with milliseconds. Spec 2.4 p.104.
func (r *record) tseStamp(column string, t time.Time) {
	if t.IsZero() {
		return
	}
	r.text(column, t.UTC().Format(tseTimeLayout))
}

// WriteExport builds e and writes the whole export, index.xml included, into sink.
func WriteExport(e Export, sink csvio.Sink, s csvio.DataSupplier) error {
	rows, tables, err := Build(e)
	if err != nil {
		return err
	}

	w, err := csvio.NewExportWriter(sink, tables, s)
	if err != nil {
		return err
	}

	for _, t := range tables {
		data := rows[t.File]
		if len(data) == 0 {
			continue
		}
		tw, terr := w.Table(t.File)
		if terr != nil {
			return errors.Join(terr, w.Close())
		}
		for _, row := range data {
			if werr := tw.Write(row); werr != nil {
				return errors.Join(werr, w.Close())
			}
		}
	}
	return w.Close()
}
