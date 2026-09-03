// Package csvio reads and writes the DSFinV-K CSV files, the index.xml that
// describes them and the directory or zip container that holds them.
package csvio

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/tweecore/dsfinvk"
	"github.com/tweecore/dsfinvk/schema"
)

var (
	// ErrColumnType reports a column whose declared type does not allow the operation.
	ErrColumnType = errors.New("wrong column type")
	// ErrAccuracy reports a value with more fraction digits than the column declares.
	ErrAccuracy = errors.New("too many fraction digits")
	// ErrTooLong reports a field longer than the column's MaxLength.
	ErrTooLong = errors.New("field exceeds MaxLength")
)

// amountFractionDigits is the presentation scale of every amount. Spec 2.4 p.31, p.90, p.91.
const amountFractionDigits = 2

// quantityAccuracy marks the quantity columns, which keep all three digits. Spec 2.4 p.96.
const quantityAccuracy = 3

// FormatValue renders v as the CSV field of a numeric column.
func FormatValue(col schema.Column, v dsfinvk.Decimal) (string, error) {
	if col.Type != schema.ColumnNumeric {
		return "", fmt.Errorf("csvio: column %s: %w: not numeric", col.Name, ErrColumnType)
	}
	if col.Accuracy < 0 {
		return "", fmt.Errorf("csvio: column %s: %w: negative accuracy %d", col.Name, ErrAccuracy, col.Accuracy)
	}

	need := fractionDigits(v)
	if need > col.Accuracy {
		return "", fmt.Errorf("csvio: column %s: %w: %s needs %d, accuracy is %d", col.Name, ErrAccuracy, v, need, col.Accuracy)
	}

	scale := minFractionDigits(col.Accuracy)
	if need > scale {
		scale = need
	}
	return v.FormatComma(scale), nil
}

// minFractionDigits is the number of digits a column always shows.
func minFractionDigits(accuracy int) int {
	switch {
	case accuracy <= 0:
		return 0
	case accuracy == quantityAccuracy:
		return quantityAccuracy
	case accuracy < amountFractionDigits:
		return accuracy
	default:
		return amountFractionDigits
	}
}

// fractionDigits is the smallest scale that represents v exactly.
func fractionDigits(v dsfinvk.Decimal) int {
	for scale := 0; scale < dsfinvk.DecimalScale; scale++ {
		if v.Truncate(scale).Equal(v) {
			return scale
		}
	}
	return dsfinvk.DecimalScale
}

// ParseValue parses the CSV field of a numeric column; ok is false for an empty field.
func ParseValue(col schema.Column, s string) (v dsfinvk.Decimal, ok bool, err error) {
	if col.Type != schema.ColumnNumeric {
		return dsfinvk.Zero, false, fmt.Errorf("csvio: column %s: %w: not numeric", col.Name, ErrColumnType)
	}
	if s == "" {
		return dsfinvk.Zero, false, nil
	}

	v, err = dsfinvk.ParseComma(s)
	if err != nil {
		return dsfinvk.Zero, false, fmt.Errorf("csvio: column %s: %w", col.Name, err)
	}
	if digits := writtenFractionDigits(s); digits > col.Accuracy {
		return dsfinvk.Zero, false, fmt.Errorf("csvio: column %s: %w: %q has %d, accuracy is %d", col.Name, ErrAccuracy, s, digits, col.Accuracy)
	}
	return v, true, nil
}

// writtenFractionDigits counts the characters after the decimal comma of s.
func writtenFractionDigits(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == schema.DecimalSymbol {
			return len(s) - 1 - i
		}
	}
	return 0
}

// CheckLength reports whether s fits the column's MaxLength, counted in runes.
// Columns without a declared MaxLength, which is every numeric one, are unlimited.
func CheckLength(col schema.Column, s string) error {
	if col.MaxLength <= 0 {
		return nil
	}
	if n := utf8.RuneCountInString(s); n > col.MaxLength {
		return fmt.Errorf("csvio: column %s: %w: %d runes, MaxLength is %d", col.Name, ErrTooLong, n, col.MaxLength)
	}
	return nil
}
