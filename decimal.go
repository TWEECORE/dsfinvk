package dsfinvk

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

// DecimalScale is the number of fraction digits a Decimal stores internally.
const DecimalScale = 5

var (
	// ErrSyntax reports malformed input to ParseComma or ParseDot.
	ErrSyntax = errors.New("invalid decimal syntax")
	// ErrOverflow reports a value outside [MinDecimal, MaxDecimal].
	ErrOverflow = errors.New("decimal overflow")
	// ErrDivideByZero reports a division by a zero divisor.
	ErrDivideByZero = errors.New("decimal division by zero")
)

// Decimal is an immutable fixed-point number with DecimalScale fraction digits.
type Decimal struct {
	units int64
}

// Zero is the Decimal value 0.
var Zero Decimal

// MaxDecimal and MinDecimal are the largest and smallest representable values.
var (
	MaxDecimal = Decimal{units: math.MaxInt64}
	MinDecimal = Decimal{units: math.MinInt64}
)

var (
	pow10  = [DecimalScale + 1]int64{1, 10, 100, 1000, 10000, 100000}
	upow10 = [DecimalScale + 1]uint64{1, 10, 100, 1000, 10000, 100000}
)

const (
	scaleFactor  = 100000
	unitsPerCent = scaleFactor / 100
)

const maxNegMagnitude uint64 = 1 << 63

// Units returns the raw fixed-point representation, the value times 10^DecimalScale.
func (d Decimal) Units() int64 { return d.units }

// ParseComma parses a CSV decimal number; ',' is the only accepted separator.
func ParseComma(s string) (Decimal, error) { return parse(s, ',', '.') }

// ParseDot parses a processData decimal number; '.' is the only accepted separator.
func ParseDot(s string) (Decimal, error) { return parse(s, '.', ',') }

func parse(s string, sep, rejected byte) (Decimal, error) {
	units, err := parseUnits(s, sep, rejected)
	if err != nil {
		return Zero, fmt.Errorf("dsfinvk: parse %q: %w", s, err)
	}
	return Decimal{units: units}, nil
}

func parseUnits(s string, sep, rejected byte) (int64, error) {
	if strings.IndexByte(s, rejected) >= 0 {
		return 0, ErrSyntax
	}

	neg := false
	if s != "" && s[0] == '-' {
		neg = true
		s = s[1:]
	}

	intPart, fracPart, found := strings.Cut(s, string(sep))
	if intPart == "" || !allDigits(intPart) {
		return 0, ErrSyntax
	}
	if found && (fracPart == "" || len(fracPart) > DecimalScale || !allDigits(fracPart)) {
		return 0, ErrSyntax
	}

	var (
		mag uint64
		ok  bool
	)
	for i := 0; i < len(intPart); i++ {
		if mag, ok = mulAddDigit(mag, intPart[i]); !ok {
			return 0, ErrOverflow
		}
	}
	for i := 0; i < DecimalScale; i++ {
		c := byte('0')
		if i < len(fracPart) {
			c = fracPart[i]
		}
		if mag, ok = mulAddDigit(mag, c); !ok {
			return 0, ErrOverflow
		}
	}
	return signedUnits(neg, mag)
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func mulAddDigit(mag uint64, c byte) (uint64, bool) {
	const limit = math.MaxUint64 / 10
	if mag > limit {
		return 0, false
	}
	mag *= 10
	d := uint64(c - '0')
	if mag > math.MaxUint64-d {
		return 0, false
	}
	return mag + d, true
}

func signedUnits(neg bool, mag uint64) (int64, error) {
	if neg {
		if mag <= math.MaxInt64 {
			return -int64(mag), nil
		}
		if mag == maxNegMagnitude {
			return math.MinInt64, nil
		}
		return 0, ErrOverflow
	}
	if mag > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return int64(mag), nil
}

func checkScale(scale int) {
	if scale < 0 || scale > DecimalScale {
		panic(fmt.Sprintf("dsfinvk: scale %d out of range [0,%d]", scale, DecimalScale))
	}
}

// FromInt returns the Decimal for a whole number and panics when it is out of range.
func FromInt(n int64) Decimal {
	return Decimal{units: mustMulPow("FromInt", n, scaleFactor)}
}

// FromCents returns the Decimal for an amount in hundredths and panics when it is out of range.
func FromCents(c int64) Decimal {
	return Decimal{units: mustMulPow("FromCents", c, unitsPerCent)}
}

func mustMulPow(name string, n, factor int64) int64 {
	units, err := mulPow(n, factor)
	if err != nil {
		panic(fmt.Sprintf("dsfinvk: %s(%d): value out of range [%s, %s]", name, n, MinDecimal, MaxDecimal))
	}
	return units
}

func mulPow(n, factor int64) (int64, error) {
	if n > math.MaxInt64/factor || n < math.MinInt64/factor {
		return 0, ErrOverflow
	}
	return n * factor, nil
}

// FromScaled returns the Decimal for n * 10^-scale; scale must be in [0, DecimalScale].
func FromScaled(n int64, scale int) (Decimal, error) {
	checkScale(scale)

	units, err := mulPow(n, pow10[DecimalScale-scale])
	if err != nil {
		return Zero, fmt.Errorf("dsfinvk: FromScaled(%d, %d): %w", n, scale, err)
	}
	return Decimal{units: units}, nil
}

// Cents returns the value in hundredths; ok is false when it is not whole cents.
func (d Decimal) Cents() (cents int64, ok bool) {
	if d.units%unitsPerCent != 0 {
		return 0, false
	}
	return d.units / unitsPerCent, true
}

// Rat returns the exact value as a new big.Rat.
func (d Decimal) Rat() *big.Rat {
	return big.NewRat(d.units, scaleFactor)
}

// FromRat returns the Decimal closest to r, rounding half away from zero.
func FromRat(r *big.Rat) (Decimal, error) {
	scaled := new(big.Int).Mul(r.Num(), big.NewInt(scaleFactor))

	quo, rem := new(big.Int).QuoRem(scaled, r.Denom(), new(big.Int))
	if rem.Sign() != 0 {
		twice := new(big.Int).Abs(rem)
		twice.Lsh(twice, 1)
		if twice.Cmp(r.Denom()) >= 0 {
			if scaled.Sign() < 0 {
				quo.Sub(quo, big.NewInt(1))
			} else {
				quo.Add(quo, big.NewInt(1))
			}
		}
	}

	if !quo.IsInt64() {
		return Zero, fmt.Errorf("dsfinvk: FromRat(%s): %w", r.RatString(), ErrOverflow)
	}
	return Decimal{units: quo.Int64()}, nil
}

// Sign returns -1 for a negative value, 0 for zero and +1 for a positive one.
func (d Decimal) Sign() int {
	switch {
	case d.units < 0:
		return -1
	case d.units > 0:
		return 1
	default:
		return 0
	}
}

// IsZero reports whether the value is zero.
func (d Decimal) IsZero() bool { return d.units == 0 }

// Neg returns -d. MinDecimal saturates to MaxDecimal.
func (d Decimal) Neg() Decimal {
	if d.units == math.MinInt64 {
		return MaxDecimal
	}
	return Decimal{units: -d.units}
}

// Abs returns the absolute value of d. MinDecimal saturates to MaxDecimal.
func (d Decimal) Abs() Decimal {
	if d.units < 0 {
		return d.Neg()
	}
	return d
}

// Cmp compares d and o and returns -1, 0 or +1.
func (d Decimal) Cmp(o Decimal) int {
	switch {
	case d.units < o.units:
		return -1
	case d.units > o.units:
		return 1
	default:
		return 0
	}
}

// Equal reports whether d and o represent the same value.
func (d Decimal) Equal(o Decimal) bool { return d.units == o.units }

// Round returns d rounded to scale fraction digits, half away from zero, and
// panics when the rounded value leaves the representable range.
func (d Decimal) Round(scale int) Decimal {
	checkScale(scale)

	factor := pow10[DecimalScale-scale]
	quo := d.units / factor
	rem := d.units % factor
	if abs64(rem)*2 >= factor {
		if d.units < 0 {
			quo--
		} else {
			quo++
		}
	}

	units, err := mulPow(quo, factor)
	if err != nil {
		panic(fmt.Sprintf("dsfinvk: %s.Round(%d): result out of range [%s, %s]", d, scale, MinDecimal, MaxDecimal))
	}
	return Decimal{units: units}
}

// Truncate returns d with the digits beyond scale fraction digits dropped.
func (d Decimal) Truncate(scale int) Decimal {
	checkScale(scale)

	factor := pow10[DecimalScale-scale]
	return Decimal{units: d.units - d.units%factor}
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func magnitude(units int64) uint64 {
	if units < 0 {
		positive := -(units + 1)
		return uint64(positive) + 1 //nolint:gosec // positive is in [0, math.MaxInt64] by construction
	}
	return uint64(units)
}

// FormatComma renders the value with a decimal comma and exactly scale fraction digits.
func (d Decimal) FormatComma(scale int) string {
	return string(d.appendFormat(nil, ',', scale))
}

// FormatDot renders the value like FormatComma but with a decimal point.
func (d Decimal) FormatDot(scale int) string {
	return string(d.appendFormat(nil, '.', scale))
}

func (d Decimal) appendFormat(dst []byte, sep byte, scale int) []byte {
	checkScale(scale)

	neg := d.units < 0
	mag := magnitude(d.units)

	factor := upow10[DecimalScale-scale]
	quo, rem := mag/factor, mag%factor
	if rem*2 >= factor {
		quo++
	}

	if quo == 0 {
		neg = false
	}
	if neg {
		dst = append(dst, '-')
	}

	unit := upow10[scale]
	dst = strconv.AppendUint(dst, quo/unit, 10)
	if scale == 0 {
		return dst
	}

	dst = append(dst, sep)
	frac := quo % unit
	for p := scale - 1; p > 0; p-- {
		if frac >= upow10[p] {
			break
		}
		dst = append(dst, '0')
	}
	return strconv.AppendUint(dst, frac, 10)
}

// processDataScale and quantityScale are the fraction digit counts of the two
// processData number forms. Spec 2.4 p.114, p.116, p.117.
const (
	processDataScale = 2
	quantityScale    = 3
)

// FormatProcessData renders the value as a TSE processData amount: a decimal
// point and exactly two fraction digits, truncated. Spec 2.4 p.114, p.117.
func (d Decimal) FormatProcessData() string {
	return string(appendTruncated(nil, d.units, processDataScale, false))
}

// FormatQuantity renders the value as a processData <Menge>: a decimal point,
// at most three fraction digits, truncated, without trailing zeros. Spec 2.4 p.116.
func (d Decimal) FormatQuantity() string {
	return string(appendTruncated(nil, d.units, quantityScale, true))
}

func appendTruncated(dst []byte, units int64, scale int, trim bool) []byte {
	mag := magnitude(units) / upow10[DecimalScale-scale]
	if units < 0 && mag != 0 {
		dst = append(dst, '-')
	}

	unit := upow10[scale]
	dst = strconv.AppendUint(dst, mag/unit, 10)

	frac := mag % unit
	if trim {
		for frac != 0 && frac%10 == 0 {
			frac /= 10
			scale--
		}
		if frac == 0 {
			return dst
		}
	}

	dst = append(dst, '.')
	for p := scale - 1; p > 0; p-- {
		if frac >= upow10[p] {
			break
		}
		dst = append(dst, '0')
	}
	return strconv.AppendUint(dst, frac, 10)
}

// String renders the value with a decimal point and no trailing zeros.
func (d Decimal) String() string {
	s := d.FormatDot(DecimalScale)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

// Add returns d + o, or ErrOverflow when the sum leaves the representable range.
func (d Decimal) Add(o Decimal) (Decimal, error) {
	sum := d.units + o.units
	if (d.units^sum)&(o.units^sum) < 0 {
		return Zero, fmt.Errorf("dsfinvk: %s + %s: %w", d, o, ErrOverflow)
	}
	return Decimal{units: sum}, nil
}

// Sub returns d - o, or ErrOverflow when the difference leaves the range.
func (d Decimal) Sub(o Decimal) (Decimal, error) {
	diff := d.units - o.units
	if (d.units^o.units)&(d.units^diff) < 0 {
		return Zero, fmt.Errorf("dsfinvk: %s - %s: %w", d, o, ErrOverflow)
	}
	return Decimal{units: diff}, nil
}

// Mul returns d * o, rounded half away from zero to DecimalScale fraction digits.
func (d Decimal) Mul(o Decimal) (Decimal, error) {
	neg := (d.units < 0) != (o.units < 0)

	hi, lo := bits.Mul64(magnitude(d.units), magnitude(o.units))
	mag, ok := divRoundHalfUp(hi, lo, scaleFactor)
	if ok {
		var units int64
		if units, ok = signedUnitsOK(neg, mag); ok {
			return Decimal{units: units}, nil
		}
	}
	return Zero, fmt.Errorf("dsfinvk: %s * %s: %w", d, o, ErrOverflow)
}

// Div returns d / o, rounded half away from zero to DecimalScale fraction digits.
func (d Decimal) Div(o Decimal) (Decimal, error) {
	if o.units == 0 {
		return Zero, fmt.Errorf("dsfinvk: %s / %s: %w", d, o, ErrDivideByZero)
	}

	neg := (d.units < 0) != (o.units < 0)

	hi, lo := bits.Mul64(magnitude(d.units), scaleFactor)
	mag, ok := divRoundHalfUp(hi, lo, magnitude(o.units))
	if ok {
		var units int64
		if units, ok = signedUnitsOK(neg, mag); ok {
			return Decimal{units: units}, nil
		}
	}
	return Zero, fmt.Errorf("dsfinvk: %s / %s: %w", d, o, ErrOverflow)
}

func divRoundHalfUp(hi, lo, divisor uint64) (uint64, bool) {
	if hi >= divisor {
		return 0, false
	}

	quo, rem := bits.Div64(hi, lo, divisor)
	if rem >= divisor-rem {
		if quo == math.MaxUint64 {
			return 0, false
		}
		quo++
	}
	return quo, true
}

func signedUnitsOK(neg bool, mag uint64) (int64, bool) {
	units, err := signedUnits(neg, mag)
	return units, err == nil
}

// MarshalText implements encoding.TextMarshaler.
func (d Decimal) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Decimal) UnmarshalText(text []byte) error {
	parsed, err := ParseDot(string(text))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalJSON implements json.Marshaler, encoding the value as a JSON string.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, d.String()), nil
}

// UnmarshalJSON implements json.Unmarshaler, accepting a JSON string, number or null.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		return nil
	}
	if len(s) >= 2 && s[0] == '"' {
		unquoted, err := strconv.Unquote(s)
		if err != nil {
			return fmt.Errorf("dsfinvk: parse %s: %w", s, ErrSyntax)
		}
		s = unquoted
	} else if s == "" || (s[0] != '-' && (s[0] < '0' || s[0] > '9')) {
		return fmt.Errorf("dsfinvk: parse %s: %w", data, ErrSyntax)
	}

	parsed, err := ParseDot(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
