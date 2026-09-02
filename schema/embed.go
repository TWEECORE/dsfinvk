// Package schema holds the DSFinV-K 2.4 table definitions, the mandatory-field
// classification, the closed enumerations and the VAT key table.
package schema

import (
	"bytes"
	_ "embed"
)

// indexXML is the index.xml every DSFinV-K export must ship, byte-exact as published.
//
//go:embed index.xml
var indexXML []byte

// dtd is the GDPdU document type definition referenced by index.xml, byte-exact.
//
//go:embed gdpdu-01-09-2004.dtd
var dtd []byte

// IndexXML returns a copy of the published index.xml.
func IndexXML() []byte { return bytes.Clone(indexXML) }

// DTD returns a copy of the GDPdU document type definition index.xml refers to.
func DTD() []byte { return bytes.Clone(dtd) }
