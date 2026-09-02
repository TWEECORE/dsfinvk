package schema_test

import (
	"bytes"
	"testing"

	"github.com/tweecore/dsfinvk/schema"
)

func TestIndexXMLEmbedded(t *testing.T) {
	if len(schema.IndexXML()) == 0 {
		t.Fatal("IndexXML is empty")
	}
	// The published file is UTF-8 with a BOM and CRLF; the copy must be byte-exact.
	if !bytes.HasPrefix(schema.IndexXML(), []byte{0xEF, 0xBB, 0xBF}) {
		t.Error("IndexXML lost its UTF-8 BOM")
	}
	if !bytes.Contains(schema.IndexXML(), []byte("\r\n")) {
		t.Error("IndexXML lost its CRLF line endings")
	}
	if !bytes.Contains(schema.IndexXML(), []byte("<URL>cashpointclosing.csv</URL>")) {
		t.Error("IndexXML does not contain the first table URL")
	}
	if got, want := bytes.Count(schema.IndexXML(), []byte("<VariableColumn>")), 219; got != want {
		t.Errorf("VariableColumn count = %d, want %d", got, want)
	}
}

func TestEmbeddedBytesAreCopies(t *testing.T) {
	a := schema.IndexXML()
	a[0] = 0
	if b := schema.IndexXML(); b[0] != 0xEF {
		t.Error("IndexXML() returns the embedded slice itself")
	}
	c := schema.DTD()
	c[0] = 0
	if d := schema.DTD(); d[0] == 0 {
		t.Error("DTD() returns the embedded slice itself")
	}
}

func TestDTDEmbedded(t *testing.T) {
	if len(schema.DTD()) == 0 {
		t.Fatal("DTD is empty")
	}
	if !bytes.Contains(schema.DTD(), []byte("<!ELEMENT VariableColumn")) {
		t.Error("DTD does not declare VariableColumn")
	}
}
