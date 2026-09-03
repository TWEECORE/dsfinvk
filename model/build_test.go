package model_test

import (
	"errors"
	"testing"
	"time"

	"github.com/tweecore/dsfinvk/model"
	"github.com/tweecore/dsfinvk/schema"
)

var testErstellung = time.Date(2019, time.January, 21, 18, 30, 55, 0, time.FixedZone("CET", 3600))

// minimalClosing is a Kassenabschluss that Build accepts.
func minimalClosing() model.Kassenabschluss {
	return model.Kassenabschluss{
		KasseID:    "K1",
		Erstellung: testErstellung,
		Nr:         1,
	}
}

func TestBuildRejectsEmptyKasseID(t *testing.T) {
	t.Parallel()

	c := minimalClosing()
	c.KasseID = ""

	if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrKasseID) {
		t.Fatalf("Build error = %v, want ErrKasseID", err)
	}
}

func TestBuildRejectsClosingNumber(t *testing.T) {
	t.Parallel()

	for _, nr := range []int64{0, -1} {
		c := minimalClosing()
		c.Nr = nr

		if _, _, err := model.Build(model.Export{Abschluesse: []model.Kassenabschluss{c}}); !errors.Is(err, model.ErrClosingNr) {
			t.Fatalf("Build(Nr=%d) error = %v, want ErrClosingNr", nr, err)
		}
	}
}

func TestBuildReturnsEveryTable(t *testing.T) {
	t.Parallel()

	rows, tables, err := model.Build(model.Export{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(tables) != len(schema.Tables()) {
		t.Fatalf("len(tables) = %d, want %d", len(tables), len(schema.Tables()))
	}
	for _, tbl := range schema.Tables() {
		if _, ok := rows[tbl.File]; !ok {
			t.Errorf("Rows has no entry for %s", tbl.File)
		}
	}
}
