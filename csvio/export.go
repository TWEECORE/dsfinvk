package csvio

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/tweecore/dsfinvk/schema"
)

var (
	// ErrDirNotEmpty reports a target directory that already holds files.
	ErrDirNotEmpty = errors.New("target directory is not empty")
	// ErrNoTables reports an export without a single table.
	ErrNoTables = errors.New("no tables to export")
	// ErrDuplicateFile reports two tables that share a CSV file name.
	ErrDuplicateFile = errors.New("duplicate table file name")
	// ErrUnknownFile reports a file name the export does not describe.
	ErrUnknownFile = errors.New("unknown table file name")
	// ErrExportClosed reports use of an ExportWriter after Close.
	ErrExportClosed = errors.New("export is closed")
	// ErrNotADirectory reports an export path that is not a directory.
	ErrNotADirectory = errors.New("not a directory")
)

// defaultMediaName is the <Media><Name> the published index.xml carries.
const defaultMediaName = "CD Nummer 1"

// zipEpoch is the modification time of every zip entry, so an export is reproducible.
var zipEpoch = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

// dirPerm and filePerm are the modes of a directory export.
const (
	dirPerm  = 0o750
	filePerm = 0o600
)

// Sink accepts the files of one export.
type Sink interface {
	Create(name string) (io.WriteCloser, error)
	Close() error
}

// dirSink writes the export files into a directory.
type dirSink struct{ path string }

// NewDirSink returns a Sink writing into path, which it creates.
// An existing directory must be empty unless overwrite is set.
func NewDirSink(path string, overwrite bool) (Sink, error) {
	if err := os.MkdirAll(path, dirPerm); err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	if len(entries) > 0 && !overwrite {
		return nil, fmt.Errorf("csvio: %s: %w", path, ErrDirNotEmpty)
	}
	return &dirSink{path: path}, nil
}

// Create opens one file of the export; name is taken as a plain file name.
func (s *dirSink) Create(name string) (io.WriteCloser, error) {
	path := filepath.Join(s.path, filepath.Base(name))
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerm) //nolint:gosec // the name is reduced to a base name inside the sink's own directory
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	return f, nil
}

func (s *dirSink) Close() error { return nil }

// zipSink collects the export files and writes them as one archive on Close.
type zipSink struct {
	w       io.Writer
	names   []string
	entries map[string]*bytes.Buffer
	closed  bool
	err     error
}

// NewZipSink returns a Sink that writes a deflated zip archive to w when it is closed.
func NewZipSink(w io.Writer) Sink {
	return &zipSink{w: w, entries: make(map[string]*bytes.Buffer)}
}

func (s *zipSink) Create(name string) (io.WriteCloser, error) {
	buf, ok := s.entries[name]
	if !ok {
		buf = &bytes.Buffer{}
		s.entries[name] = buf
		s.names = append(s.names, name)
	}
	buf.Reset()
	return nopCloser{buf}, nil
}

// Close writes the archive; a second call repeats the first result.
func (s *zipSink) Close() error {
	if s.closed {
		return s.err
	}
	s.closed = true

	zw := zip.NewWriter(s.w)
	for _, name := range s.names {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: zipEpoch})
		if err != nil {
			s.err = fmt.Errorf("csvio: %s: %w", name, err)
			return s.err
		}
		if _, err := w.Write(s.entries[name].Bytes()); err != nil {
			s.err = fmt.Errorf("csvio: %s: %w", name, err)
			return s.err
		}
	}
	if err := zw.Close(); err != nil {
		s.err = fmt.Errorf("csvio: %w", err)
	}
	return s.err
}

// nopCloser adds a no-op Close to a writer.
type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

// ExportWriter writes the CSV files, index.xml and the DTD of one export.
type ExportWriter struct {
	sink     Sink
	tables   []schema.Table
	byFile   map[string]int
	supplier DataSupplier
	writers  map[string]*Writer
	closers  map[string]io.WriteCloser
	order    []string
	closed   bool
	err      error
}

// NewExportWriter returns an ExportWriter for tables, writing into sink.
func NewExportWriter(sink Sink, tables []schema.Table, s DataSupplier) (*ExportWriter, error) {
	if len(tables) == 0 {
		return nil, fmt.Errorf("csvio: %w", ErrNoTables)
	}

	byFile := make(map[string]int, len(tables))
	own := make([]schema.Table, len(tables))
	for i, t := range tables {
		if _, dup := byFile[t.File]; dup {
			return nil, fmt.Errorf("csvio: %w %q", ErrDuplicateFile, t.File)
		}
		byFile[t.File] = i
		own[i] = cloneTable(t)
	}

	return &ExportWriter{
		sink:     sink,
		tables:   own,
		byFile:   byFile,
		supplier: s,
		writers:  make(map[string]*Writer, len(own)),
		closers:  make(map[string]io.WriteCloser, len(own)),
	}, nil
}

// Table returns the Writer of one CSV file, creating it with its header on first use.
func (e *ExportWriter) Table(file string) (*Writer, error) {
	if e.closed {
		return nil, fmt.Errorf("csvio: %s: %w", file, ErrExportClosed)
	}
	return e.table(file)
}

func (e *ExportWriter) table(file string) (*Writer, error) {
	if w, ok := e.writers[file]; ok {
		return w, nil
	}

	i, ok := e.byFile[file]
	if !ok {
		return nil, fmt.Errorf("csvio: %w %q", ErrUnknownFile, file)
	}

	wc, err := e.sink.Create(file)
	if err != nil {
		return nil, fmt.Errorf("csvio: %s: %w", file, err)
	}

	w := NewWriter(wc, e.tables[i])
	if err := w.WriteHeader(); err != nil {
		return nil, err
	}
	e.writers[file] = w
	e.closers[file] = wc
	e.order = append(e.order, file)
	return w, nil
}

// Close writes the missing headers, index.xml and the DTD, then closes the sink.
// A second call repeats the first result.
func (e *ExportWriter) Close() error {
	if e.closed {
		return e.err
	}

	for _, t := range e.tables {
		if _, err := e.table(t.File); err != nil {
			e.keep(err)
		}
	}
	e.closed = true

	e.keep(e.writeIndex())
	e.keep(e.writeBytes(dtdName, schema.DTD()))

	for _, file := range e.order {
		w := e.writers[file]
		w.Flush()
		e.keep(w.Error())
		e.keep(e.closers[file].Close())
	}
	e.keep(e.sink.Close())
	return e.err
}

func (e *ExportWriter) writeIndex() error {
	wc, err := e.sink.Create(indexXMLName)
	if err != nil {
		return fmt.Errorf("csvio: %s: %w", indexXMLName, err)
	}
	return errors.Join(WriteIndexXML(wc, e.tables, e.supplier, defaultMediaName), wc.Close())
}

func (e *ExportWriter) writeBytes(name string, b []byte) error {
	wc, err := e.sink.Create(name)
	if err != nil {
		return fmt.Errorf("csvio: %s: %w", name, err)
	}
	if _, err := wc.Write(b); err != nil {
		return errors.Join(fmt.Errorf("csvio: %s: %w", name, err), wc.Close())
	}
	return wc.Close()
}

// keep records the first error of the export.
func (e *ExportWriter) keep(err error) {
	if err != nil && e.err == nil {
		e.err = err
	}
}

// Source reads the files of one export.
type Source interface {
	Open(name string) (io.ReadCloser, error)
	Names() ([]string, error)
}

// dirSource reads an export from a directory.
type dirSource struct{ path string }

// OpenDir returns a Source reading the export in path.
func OpenDir(path string) (Source, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("csvio: %s: %w", path, ErrNotADirectory)
	}
	return &dirSource{path: path}, nil
}

func (s *dirSource) Open(name string) (io.ReadCloser, error) {
	f, err := os.Open(filepath.Join(s.path, filepath.Base(name)))
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	return f, nil
}

// Names returns the file names of the export, sorted.
func (s *dirSource) Names() ([]string, error) {
	entries, err := os.ReadDir(s.path)
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	slices.Sort(names)
	return names, nil
}

// zipSource reads an export from a zip archive.
type zipSource struct {
	r      *zip.Reader
	closer io.Closer
}

// NewZipSource returns a Source reading the export in the archive r of size bytes.
func NewZipSource(r io.ReaderAt, size int64) (Source, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	return &zipSource{r: zr}, nil
}

// OpenZip returns a Source reading the zip archive at path; it is also an io.Closer.
func OpenZip(path string) (Source, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	return &zipSource{r: &zr.Reader, closer: zr}, nil
}

func (s *zipSource) Open(name string) (io.ReadCloser, error) {
	rc, err := s.r.Open(name)
	if err != nil {
		return nil, fmt.Errorf("csvio: %w", err)
	}
	return rc, nil
}

// Names returns the entry names of the archive, sorted.
func (s *zipSource) Names() ([]string, error) {
	names := make([]string, 0, len(s.r.File))
	for _, f := range s.r.File {
		names = append(names, f.Name)
	}
	slices.Sort(names)
	return names, nil
}

// Close releases the archive file OpenZip holds open.
func (s *zipSource) Close() error {
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

// ReadIndex parses the index.xml of an export.
func ReadIndex(src Source) ([]schema.Table, DataSupplier, string, error) {
	rc, err := src.Open(indexXMLName)
	if err != nil {
		return nil, DataSupplier{}, "", err
	}

	tables, supplier, media, err := ReadIndexXML(rc)
	if cerr := rc.Close(); cerr != nil && err == nil {
		err = fmt.Errorf("csvio: %s: %w", indexXMLName, cerr)
	}
	if err != nil {
		return nil, DataSupplier{}, "", err
	}
	return tables, supplier, media, nil
}
