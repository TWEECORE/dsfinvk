package dsfinvk_test

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/tweecore/dsfinvk"
)

func TestDecimalScaleConstant(t *testing.T) {
	t.Parallel()

	if dsfinvk.DecimalScale != 5 {
		t.Fatalf("DecimalScale = %d, want 5", dsfinvk.DecimalScale)
	}
	if got := dsfinvk.Zero.Units(); got != 0 {
		t.Fatalf("Zero.Units() = %d, want 0", got)
	}
	if got := dsfinvk.MaxDecimal.Units(); got != 9223372036854775807 {
		t.Fatalf("MaxDecimal.Units() = %d, want 9223372036854775807", got)
	}
	if got := dsfinvk.MinDecimal.Units(); got != -9223372036854775808 {
		t.Fatalf("MinDecimal.Units() = %d, want -9223372036854775808", got)
	}
}

func mustComma(s string) dsfinvk.Decimal {
	d, err := dsfinvk.ParseComma(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestParseCommaValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in    string
		units int64
	}{
		{"0", 0},
		{"0,00", 0},
		{"-0", 0},
		{"-0,0", 0},
		{"1", 100000},
		{"1234,56", 123456000},
		{"-0,5", -50000},
		{"12,34567", 1234567},
		{"-12,34567", -1234567},
		{"0,00001", 1},
		{"000123,4", 12340000},
		{"92233720368547,75807", 9223372036854775807},
		{"-92233720368547,75808", -9223372036854775808},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got, err := dsfinvk.ParseComma(tt.in)
			if err != nil {
				t.Fatalf("ParseComma(%q) returned error: %v", tt.in, err)
			}
			if got.Units() != tt.units {
				t.Fatalf("ParseComma(%q).Units() = %d, want %d", tt.in, got.Units(), tt.units)
			}
		})
	}
}

func TestParseDotValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in    string
		units int64
	}{
		{"0", 0},
		{"0.00", 0},
		{"-0", 0},
		{"1234.56", 123456000},
		{"-0.5", -50000},
		{"0.00001", 1},
		{"000123.4", 12340000},
		{"92233720368547.75807", 9223372036854775807},
		{"-92233720368547.75808", -9223372036854775808},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got, err := dsfinvk.ParseDot(tt.in)
			if err != nil {
				t.Fatalf("ParseDot(%q) returned error: %v", tt.in, err)
			}
			if got.Units() != tt.units {
				t.Fatalf("ParseDot(%q).Units() = %d, want %d", tt.in, got.Units(), tt.units)
			}
		})
	}
}

func TestParseCommaInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want error
	}{
		{"empty", "", dsfinvk.ErrSyntax},
		{"leading space", " 1", dsfinvk.ErrSyntax},
		{"trailing space", "1 ", dsfinvk.ErrSyntax},
		{"explicit plus", "+1", dsfinvk.ErrSyntax},
		{"six fraction digits", "1,234567", dsfinvk.ErrSyntax},
		{"grouping separator", "1.234,56", dsfinvk.ErrSyntax},
		{"decimal point", "1.234", dsfinvk.ErrSyntax},
		{"trailing point", "1.", dsfinvk.ErrSyntax},
		{"exponent", "1e3", dsfinvk.ErrSyntax},
		{"only sign", "-", dsfinvk.ErrSyntax},
		{"only separator", ",", dsfinvk.ErrSyntax},
		{"no integer digits", ",5", dsfinvk.ErrSyntax},
		{"no fraction digits", "1,", dsfinvk.ErrSyntax},
		{"two separators", "1,2,3", dsfinvk.ErrSyntax},
		{"letters", "abc", dsfinvk.ErrSyntax},
		{"trailing letter", "1,5x", dsfinvk.ErrSyntax},
		{"nbsp", "1 234,56", dsfinvk.ErrSyntax},
		{"overflow positive", "92233720368547,75808", dsfinvk.ErrOverflow},
		{"overflow negative", "-92233720368547,75809", dsfinvk.ErrOverflow},
		{"overflow huge", "99999999999999999999", dsfinvk.ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := dsfinvk.ParseComma(tt.in)
			if err == nil {
				t.Fatalf("ParseComma(%q) = %v, want error", tt.in, got)
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("ParseComma(%q) error = %v, want errors.Is(..., %v)", tt.in, err, tt.want)
			}
		})
	}
}

func TestParseDotInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want error
	}{
		{"empty", "", dsfinvk.ErrSyntax},
		{"decimal comma", "1,5", dsfinvk.ErrSyntax},
		{"grouping separator", "1,234.56", dsfinvk.ErrSyntax},
		{"trailing comma", "1,", dsfinvk.ErrSyntax},
		{"explicit plus", "+1", dsfinvk.ErrSyntax},
		{"six fraction digits", "1.234567", dsfinvk.ErrSyntax},
		{"two separators", "1.2.3", dsfinvk.ErrSyntax},
		{"only sign", "-", dsfinvk.ErrSyntax},
		{"no integer digits", ".5", dsfinvk.ErrSyntax},
		{"no fraction digits", "1.", dsfinvk.ErrSyntax},
		{"letters", "abc", dsfinvk.ErrSyntax},
		{"overflow positive", "92233720368547.75808", dsfinvk.ErrOverflow},
		{"overflow huge", "99999999999999999999", dsfinvk.ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := dsfinvk.ParseDot(tt.in)
			if err == nil {
				t.Fatalf("ParseDot(%q) = %v, want error", tt.in, got)
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("ParseDot(%q) error = %v, want errors.Is(..., %v)", tt.in, err, tt.want)
			}
		})
	}
}

func TestFromInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    int64
		units int64
	}{
		{"zero", 0, 0},
		{"one", 1, 100000},
		{"negative", -42, -4200000},
		{"largest exact", 92233720368547, 9223372036854700000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := dsfinvk.FromInt(tt.in).Units(); got != tt.units {
				t.Fatalf("FromInt(%d).Units() = %d, want %d", tt.in, got, tt.units)
			}
		})
	}
}

func TestFromCents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    int64
		units int64
	}{
		{"zero", 0, 0},
		{"one cent", 1, 1000},
		{"euro", 100, 100000},
		{"negative", -1234, -1234000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := dsfinvk.FromCents(tt.in).Units(); got != tt.units {
				t.Fatalf("FromCents(%d).Units() = %d, want %d", tt.in, got, tt.units)
			}
		})
	}
}

func TestFromScaled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		n       int64
		scale   int
		units   int64
		wantErr error
	}{
		{name: "scale 0", n: 7, scale: 0, units: 700000},
		{name: "scale 2", n: 1234, scale: 2, units: 1234000},
		{name: "scale 3", n: -1500, scale: 3, units: -150000},
		{name: "scale 5 identity", n: 9223372036854775807, scale: 5, units: 9223372036854775807},
		{name: "overflow", n: 9223372036854775807, scale: 4, wantErr: dsfinvk.ErrOverflow},
		{name: "overflow negative", n: -9223372036854775808, scale: 0, wantErr: dsfinvk.ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := dsfinvk.FromScaled(tt.n, tt.scale)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("FromScaled(%d, %d) error = %v, want %v", tt.n, tt.scale, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FromScaled(%d, %d) returned error: %v", tt.n, tt.scale, err)
			}
			if got.Units() != tt.units {
				t.Fatalf("FromScaled(%d, %d).Units() = %d, want %d", tt.n, tt.scale, got.Units(), tt.units)
			}
		})
	}

	t.Run("panics on invalid scale", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if recover() == nil {
				t.Fatal("FromScaled did not panic on out-of-range scale")
			}
		}()
		_, _ = dsfinvk.FromScaled(1, 6)
	})
}

func TestCents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in    string
		cents int64
		ok    bool
	}{
		{"0", 0, true},
		{"12,34", 1234, true},
		{"-12,34", -1234, true},
		{"12,345", 0, false},
		{"0,00001", 0, false},
		{"-0,001", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			cents, ok := mustComma(tt.in).Cents()
			if ok != tt.ok || cents != tt.cents {
				t.Fatalf("%q.Cents() = (%d, %t), want (%d, %t)", tt.in, cents, ok, tt.cents, tt.ok)
			}
		})
	}
}

func TestRat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  string
		num int64
		den int64
	}{
		{"0", 0, 1},
		{"1,5", 3, 2},
		{"-0,25", -1, 4},
		{"12,34567", 1234567, 100000},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			got := mustComma(tt.in).Rat()
			want := big.NewRat(tt.num, tt.den)
			if got.Cmp(want) != 0 {
				t.Fatalf("%q.Rat() = %s, want %s", tt.in, got.RatString(), want.RatString())
			}
		})
	}
}

func TestFromRat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		num     int64
		den     int64
		units   int64
		wantErr error
	}{
		{name: "exact", num: 3, den: 2, units: 150000},
		{name: "one third rounds down", num: 1, den: 3, units: 33333},
		{name: "two thirds rounds up", num: 2, den: 3, units: 66667},
		{name: "half away from zero up", num: 1, den: 200000, units: 1},
		{name: "half away from zero down", num: -1, den: 200000, units: -1},
		{name: "half away from zero at 1.5 units", num: 3, den: 200000, units: 2},
		{name: "below half stays", num: 4, den: 1000000, units: 0},
		{name: "max", num: 9223372036854775807, den: 100000, units: 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := dsfinvk.FromRat(big.NewRat(tt.num, tt.den))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("FromRat(%d/%d) error = %v, want %v", tt.num, tt.den, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FromRat(%d/%d) returned error: %v", tt.num, tt.den, err)
			}
			if got.Units() != tt.units {
				t.Fatalf("FromRat(%d/%d).Units() = %d, want %d", tt.num, tt.den, got.Units(), tt.units)
			}
		})
	}

	t.Run("overflow", func(t *testing.T) {
		t.Parallel()

		huge := new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 100))
		if _, err := dsfinvk.FromRat(huge); !errors.Is(err, dsfinvk.ErrOverflow) {
			t.Fatalf("FromRat(2^100) error = %v, want ErrOverflow", err)
		}
	})

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		for _, s := range []string{"0", "-0,5", "12,34567", "92233720368547,75807", "-92233720368547,75808"} {
			d := mustComma(s)
			back, err := dsfinvk.FromRat(d.Rat())
			if err != nil {
				t.Fatalf("FromRat(%q.Rat()) returned error: %v", s, err)
			}
			if back != d {
				t.Fatalf("FromRat(%q.Rat()).Units() = %d, want %d", s, back.Units(), d.Units())
			}
		}
	})
}

func TestSignAndZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in     string
		sign   int
		isZero bool
	}{
		{"0", 0, true},
		{"-0", 0, true},
		{"0,00001", 1, false},
		{"-0,00001", -1, false},
		{"1234,56", 1, false},
		{"-1234,56", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			d := mustComma(tt.in)
			if got := d.Sign(); got != tt.sign {
				t.Errorf("%q.Sign() = %d, want %d", tt.in, got, tt.sign)
			}
			if got := d.IsZero(); got != tt.isZero {
				t.Errorf("%q.IsZero() = %t, want %t", tt.in, got, tt.isZero)
			}
		})
	}
}

func TestNegAbs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  string
		neg string
		abs string
	}{
		{"0", "0", "0"},
		{"1,5", "-1,5", "1,5"},
		{"-1,5", "1,5", "1,5"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			d := mustComma(tt.in)
			if got, want := d.Neg(), mustComma(tt.neg); got != want {
				t.Errorf("%q.Neg().Units() = %d, want %d", tt.in, got.Units(), want.Units())
			}
			if got, want := d.Abs(), mustComma(tt.abs); got != want {
				t.Errorf("%q.Abs().Units() = %d, want %d", tt.in, got.Units(), want.Units())
			}
		})
	}

	t.Run("MinDecimal saturates", func(t *testing.T) {
		t.Parallel()

		if got := dsfinvk.MinDecimal.Neg(); got != dsfinvk.MaxDecimal {
			t.Errorf("MinDecimal.Neg().Units() = %d, want MaxDecimal", got.Units())
		}
		if got := dsfinvk.MinDecimal.Abs(); got != dsfinvk.MaxDecimal {
			t.Errorf("MinDecimal.Abs().Units() = %d, want MaxDecimal", got.Units())
		}
	})
}

func TestCmpEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b string
		cmp  int
	}{
		{"0", "0", 0},
		{"0", "-0", 0},
		{"1,5", "1,50000", 0},
		{"1,5", "1,50001", -1},
		{"1,50001", "1,5", 1},
		{"-1", "1", -1},
		{"-92233720368547,75808", "92233720368547,75807", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			t.Parallel()

			a := mustComma(tt.a)
			b := mustComma(tt.b)
			if got := a.Cmp(b); got != tt.cmp {
				t.Errorf("%q.Cmp(%q) = %d, want %d", tt.a, tt.b, got, tt.cmp)
			}
			if got, want := a.Equal(b), tt.cmp == 0; got != want {
				t.Errorf("%q.Equal(%q) = %t, want %t", tt.a, tt.b, got, want)
			}
		})
	}
}

func TestRound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in    string
		scale int
		want  string
	}{
		{"0,125", 2, "0,13"},
		{"-0,125", 2, "-0,13"},
		{"0,124", 2, "0,12"},
		{"0,004", 2, "0"},
		{"-0,004", 2, "0"},
		{"2,5", 0, "3"},
		{"-2,5", 0, "-3"},
		{"2,4", 0, "2"},
		{"1,23456", 3, "1,235"},
		{"-1,23456", 3, "-1,235"},
		{"1,2345", 4, "1,2345"},
		{"12,34567", 5, "12,34567"},
		{"1,0005", 3, "1,001"},
	}

	for _, tt := range tests {
		t.Run(tt.in+"_at_"+string(rune('0'+tt.scale)), func(t *testing.T) {
			t.Parallel()

			got := mustComma(tt.in).Round(tt.scale)
			want := mustComma(tt.want)
			if got != want {
				t.Fatalf("%q.Round(%d).Units() = %d, want %d", tt.in, tt.scale, got.Units(), want.Units())
			}
		})
	}

	t.Run("panics when rounding leaves the range", func(t *testing.T) {
		t.Parallel()

		for name, d := range map[string]dsfinvk.Decimal{"max": dsfinvk.MaxDecimal, "min": dsfinvk.MinDecimal} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				defer func() {
					r := recover()
					if r == nil {
						t.Fatal("Round did not panic on an out-of-range result")
					}
					if msg, ok := r.(string); !ok || !strings.Contains(msg, d.String()) {
						t.Fatalf("panic value %v does not name %s", r, d)
					}
				}()
				_ = d.Round(0)
			})
		}
	})

	t.Run("panics on invalid scale", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if recover() == nil {
				t.Fatal("Round did not panic on out-of-range scale")
			}
		}()
		_ = dsfinvk.Zero.Round(-1)
	})
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in    string
		scale int
		want  string
	}{
		{"0,125", 2, "0,12"},
		{"-0,125", 2, "-0,12"},
		{"2,9", 0, "2"},
		{"-2,9", 0, "-2"},
		{"12,34567", 5, "12,34567"},
		{"-0,004", 2, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.in+"_at_"+string(rune('0'+tt.scale)), func(t *testing.T) {
			t.Parallel()

			got := mustComma(tt.in).Truncate(tt.scale)
			want := mustComma(tt.want)
			if got != want {
				t.Fatalf("%q.Truncate(%d).Units() = %d, want %d", tt.in, tt.scale, got.Units(), want.Units())
			}
		})
	}
}

func TestFromIntPanicsOnOverflow(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		call    func()
		offends string
	}{
		"FromInt positive":   {func() { _ = dsfinvk.FromInt(92233720368548) }, "92233720368548"},
		"FromInt negative":   {func() { _ = dsfinvk.FromInt(-92233720368548) }, "-92233720368548"},
		"FromCents positive": {func() { _ = dsfinvk.FromCents(9223372036854775807) }, "9223372036854775807"},
		"FromCents negative": {func() { _ = dsfinvk.FromCents(-9223372036854775808) }, "-9223372036854775808"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("no panic on an out-of-range value")
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("panic value %v is not a string", r)
				}
				if !strings.Contains(msg, "("+tt.offends+")") {
					t.Fatalf("panic message %q does not name the offending value %s", msg, tt.offends)
				}
			}()
			tt.call()
		})
	}
}

func TestFormatComma(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		in    string
		scale int
		want  string
	}{
		{name: "zero scale 0", in: "0", scale: 0, want: "0"},
		{name: "zero scale 2", in: "0", scale: 2, want: "0,00"},
		{name: "zero scale 3", in: "0", scale: 3, want: "0,000"},
		{name: "zero scale 5", in: "0", scale: 5, want: "0,00000"},
		{name: "amount", in: "1234,56", scale: 2, want: "1234,56"},
		{name: "negative amount", in: "-1234,56", scale: 2, want: "-1234,56"},
		{name: "pads fraction", in: "1,5", scale: 3, want: "1,500"},
		{name: "quantity scale 3", in: "2,5", scale: 3, want: "2,500"},
		{name: "half up", in: "0,125", scale: 2, want: "0,13"},
		{name: "half away from zero", in: "-0,125", scale: 2, want: "-0,13"},
		{name: "no negative zero", in: "-0,004", scale: 2, want: "0,00"},
		{name: "no negative zero scale 0", in: "-0,4", scale: 0, want: "0"},
		{name: "small positive rounds to zero", in: "0,004", scale: 2, want: "0,00"},
		{name: "half at scale 0", in: "2,5", scale: 0, want: "3"},
		{name: "negative half at scale 0", in: "-2,5", scale: 0, want: "-3"},
		{name: "full precision", in: "12,34567", scale: 5, want: "12,34567"},
		{name: "leading zeros trimmed", in: "007,5", scale: 2, want: "7,50"},
		{name: "max", in: "92233720368547,75807", scale: 5, want: "92233720368547,75807"},
		{name: "max rounded to units", in: "92233720368547,75807", scale: 0, want: "92233720368548"},
		{name: "min", in: "-92233720368547,75808", scale: 5, want: "-92233720368547,75808"},
		{name: "min rounded to units", in: "-92233720368547,75808", scale: 0, want: "-92233720368548"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := mustComma(tt.in)
			if got := d.FormatComma(tt.scale); got != tt.want {
				t.Fatalf("%q.FormatComma(%d) = %q, want %q", tt.in, tt.scale, got, tt.want)
			}
			wantDot := strings.ReplaceAll(tt.want, ",", ".")
			if got := d.FormatDot(tt.scale); got != wantDot {
				t.Fatalf("%q.FormatDot(%d) = %q, want %q", tt.in, tt.scale, got, wantDot)
			}
		})
	}

	t.Run("panics on invalid scale", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if recover() == nil {
				t.Fatal("FormatComma did not panic on out-of-range scale")
			}
		}()
		_ = dsfinvk.Zero.FormatComma(6)
	})
}

// Spec 2.4 p.114, p.117: processData amounts are truncated, never rounded.
func TestFormatProcessData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"0", "0.00"},
		{"-0", "0.00"},
		{"1,999", "1.99"},
		{"0,009", "0.00"},
		{"1,005", "1.00"},
		{"-1,999", "-1.99"},
		{"-0,004", "0.00"},
		{"-5", "-5.00"},
		{"12,5", "12.50"},
		{"0,5", "0.50"},
		{"-0,5", "-0.50"},
		{"1000", "1000.00"},
		{"-0,001", "0.00"},
		{"007,5", "7.50"},
		{"92233720368547,75807", "92233720368547.75"},
		{"-92233720368547,75808", "-92233720368547.75"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := mustComma(tt.in).FormatProcessData(); got != tt.want {
				t.Fatalf("%q.FormatProcessData() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// Spec 2.4 p.116: <Menge> carries the fewest digits possible, truncated to three.
func TestFormatQuantity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"1,0000", "1"},
		{"0,5", "0.5"},
		{"2,1256", "2.125"},
		{"0", "0"},
		{"-0", "0"},
		{"0,451", "0.451"},
		{"0,045", "0.045"},
		{"0,00009", "0"},
		{"-0,00009", "0"},
		{"-2,5", "-2.5"},
		{"-0,4509", "-0.45"},
		{"1000", "1000"},
		{"12,34567", "12.345"},
		{"92233720368547,75807", "92233720368547.758"},
		{"-92233720368547,75808", "-92233720368547.758"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := mustComma(tt.in).FormatQuantity(); got != tt.want {
				t.Fatalf("%q.FormatQuantity() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"0", "0"},
		{"-0", "0"},
		{"1", "1"},
		{"1,5", "1.5"},
		{"-1,5", "-1.5"},
		{"1,50000", "1.5"},
		{"0,00001", "0.00001"},
		{"-0,00001", "-0.00001"},
		{"12,34567", "12.34567"},
		{"92233720368547,75807", "92233720368547.75807"},
		{"-92233720368547,75808", "-92233720368547.75808"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := mustComma(tt.in).String(); got != tt.want {
				t.Fatalf("%q.String() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestAddSub(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		a, b   string
		add    string
		sub    string
		addErr error
		subErr error
	}{
		{name: "simple", a: "1,5", b: "2,25", add: "3,75", sub: "-0,75"},
		{name: "zero", a: "0", b: "0", add: "0", sub: "0"},
		{name: "negatives", a: "-1,5", b: "-2,5", add: "-4", sub: "1"},
		{name: "cancel", a: "1234,56", b: "-1234,56", add: "0", sub: "2469,12"},
		{name: "smallest unit", a: "0,00001", b: "0,00001", add: "0,00002", sub: "0"},
		{name: "max plus zero", a: "92233720368547,75807", b: "0", add: "92233720368547,75807", sub: "92233720368547,75807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := mustComma(tt.a)
			b := mustComma(tt.b)

			sum, err := a.Add(b)
			if err != nil {
				t.Fatalf("%q.Add(%q) returned error: %v", tt.a, tt.b, err)
			}
			if want := mustComma(tt.add); sum != want {
				t.Errorf("%q.Add(%q) = %s, want %s", tt.a, tt.b, sum, want)
			}

			diff, err := a.Sub(b)
			if err != nil {
				t.Fatalf("%q.Sub(%q) returned error: %v", tt.a, tt.b, err)
			}
			if want := mustComma(tt.sub); diff != want {
				t.Errorf("%q.Sub(%q) = %s, want %s", tt.a, tt.b, diff, want)
			}
		})
	}
}

func TestAddSubOverflow(t *testing.T) {
	t.Parallel()

	one := dsfinvk.FromInt(1)

	if _, err := dsfinvk.MaxDecimal.Add(one); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("MaxDecimal.Add(1) error = %v, want ErrOverflow", err)
	}
	if _, err := dsfinvk.MaxDecimal.Add(mustComma("0,00001")); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("MaxDecimal.Add(0,00001) error = %v, want ErrOverflow", err)
	}
	if _, err := dsfinvk.MinDecimal.Sub(mustComma("0,00001")); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("MinDecimal.Sub(0,00001) error = %v, want ErrOverflow", err)
	}
	if _, err := dsfinvk.MinDecimal.Add(dsfinvk.MinDecimal); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("MinDecimal.Add(MinDecimal) error = %v, want ErrOverflow", err)
	}
	if _, err := dsfinvk.Zero.Sub(dsfinvk.MinDecimal); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("Zero.Sub(MinDecimal) error = %v, want ErrOverflow", err)
	}
}

func TestMul(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want string
	}{
		{name: "whole", a: "1,5", b: "2", want: "3"},
		{name: "fraction", a: "0,1", b: "0,1", want: "0,01"},
		{name: "by zero", a: "1234,56", b: "0", want: "0"},
		{name: "negative", a: "-1,5", b: "2", want: "-3"},
		{name: "both negative", a: "-1,5", b: "-2", want: "3"},
		{name: "vat 19 percent", a: "100", b: "0,19", want: "19"},
		{name: "intermediate exceeds int64", a: "1000000", b: "1000000", want: "1000000000000"},
		{name: "smallest unit scaled", a: "100000000", b: "0,00001", want: "1000"},
		{name: "half away from zero up", a: "0,00001", b: "0,5", want: "0,00001"},
		{name: "half away from zero down", a: "-0,00001", b: "0,5", want: "-0,00001"},
		{name: "below half", a: "0,00001", b: "0,4", want: "0"},
		{name: "rounds to five digits", a: "0,33333", b: "0,33333", want: "0,11111"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := mustComma(tt.a).Mul(mustComma(tt.b))
			if err != nil {
				t.Fatalf("%q.Mul(%q) returned error: %v", tt.a, tt.b, err)
			}
			if want := mustComma(tt.want); got != want {
				t.Fatalf("%q.Mul(%q) = %s, want %s", tt.a, tt.b, got, want)
			}
		})
	}
}

func TestMulOverflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
	}{
		{name: "max times two", a: "92233720368547,75807", b: "2"},
		{name: "min times minus one", a: "-92233720368547,75808", b: "-1"},
		{name: "far out of range", a: "10000000", b: "10000000"},
		{name: "product needs more than 64 bits", a: "20000000", b: "20000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := mustComma(tt.a).Mul(mustComma(tt.b)); !errors.Is(err, dsfinvk.ErrOverflow) {
				t.Fatalf("%q.Mul(%q) error = %v, want ErrOverflow", tt.a, tt.b, err)
			}
		})
	}
}

func TestDiv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want string
	}{
		{name: "exact", a: "10", b: "4", want: "2,5"},
		{name: "exact fraction", a: "12,5", b: "2,5", want: "5"},
		{name: "one third", a: "1", b: "3", want: "0,33333"},
		{name: "two thirds", a: "2", b: "3", want: "0,66667"},
		{name: "negative dividend", a: "-2", b: "3", want: "-0,66667"},
		{name: "negative divisor", a: "2", b: "-3", want: "-0,66667"},
		{name: "both negative", a: "-2", b: "-3", want: "0,66667"},
		{name: "zero dividend", a: "0", b: "5", want: "0"},
		{name: "half away from zero", a: "0,00001", b: "2", want: "0,00001"},
		{name: "by one", a: "-1234,56", b: "1", want: "-1234,56"},
		{name: "large quotient", a: "92233720368547,75807", b: "1", want: "92233720368547,75807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := mustComma(tt.a).Div(mustComma(tt.b))
			if err != nil {
				t.Fatalf("%q.Div(%q) returned error: %v", tt.a, tt.b, err)
			}
			if want := mustComma(tt.want); got != want {
				t.Fatalf("%q.Div(%q) = %s, want %s", tt.a, tt.b, got, want)
			}
		})
	}
}

func TestDivErrors(t *testing.T) {
	t.Parallel()

	if _, err := dsfinvk.FromInt(1).Div(dsfinvk.Zero); !errors.Is(err, dsfinvk.ErrDivideByZero) {
		t.Errorf("1.Div(0) error = %v, want ErrDivideByZero", err)
	}
	if _, err := dsfinvk.Zero.Div(dsfinvk.Zero); !errors.Is(err, dsfinvk.ErrDivideByZero) {
		t.Errorf("0.Div(0) error = %v, want ErrDivideByZero", err)
	}
	if _, err := dsfinvk.MaxDecimal.Div(mustComma("0,5")); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("MaxDecimal.Div(0,5) error = %v, want ErrOverflow", err)
	}
	if _, err := dsfinvk.MinDecimal.Div(dsfinvk.FromInt(-1)); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("MinDecimal.Div(-1) error = %v, want ErrOverflow", err)
	}
	if _, err := dsfinvk.MaxDecimal.Div(mustComma("0,00001")); !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Errorf("MaxDecimal.Div(0,00001) error = %v, want ErrOverflow", err)
	}
	if _, err := dsfinvk.FromInt(1).Div(mustComma("0,00001")); err != nil {
		t.Errorf("1.Div(0,00001) returned error: %v", err)
	}
}

func TestMarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"0", "0"},
		{"1,5", "1.5"},
		{"-1234,56", "-1234.56"},
		{"0,00001", "0.00001"},
		{"-92233720368547,75808", "-92233720368547.75808"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			d := mustComma(tt.in)
			text, err := d.MarshalText()
			if err != nil {
				t.Fatalf("%q.MarshalText() returned error: %v", tt.in, err)
			}
			if string(text) != tt.want {
				t.Fatalf("%q.MarshalText() = %q, want %q", tt.in, text, tt.want)
			}

			var back dsfinvk.Decimal
			if err := back.UnmarshalText(text); err != nil {
				t.Fatalf("UnmarshalText(%q) returned error: %v", text, err)
			}
			if back != d {
				t.Fatalf("UnmarshalText(%q) = %s, want %s", text, back, d)
			}
		})
	}
}

func TestUnmarshalTextInvalid(t *testing.T) {
	t.Parallel()

	var d dsfinvk.Decimal
	if err := d.UnmarshalText([]byte("1.234567")); !errors.Is(err, dsfinvk.ErrSyntax) {
		t.Fatalf("UnmarshalText(%q) error = %v, want ErrSyntax", "1.234567", err)
	}
	if err := d.UnmarshalText([]byte("1,5")); !errors.Is(err, dsfinvk.ErrSyntax) {
		t.Fatalf("UnmarshalText(%q) error = %v, want ErrSyntax", "1,5", err)
	}
	if err := d.UnmarshalText([]byte("1.5")); err != nil {
		t.Fatalf("UnmarshalText(%q) returned error: %v", "1.5", err)
	}
	if want := mustComma("1,5"); d != want {
		t.Fatalf("UnmarshalText(%q) = %s, want %s", "1.5", d, want)
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	type wrapper struct {
		Amount dsfinvk.Decimal `json:"amount"`
	}

	t.Run("marshals as string", func(t *testing.T) {
		t.Parallel()

		got, err := json.Marshal(wrapper{Amount: mustComma("-1234,56")})
		if err != nil {
			t.Fatalf("json.Marshal returned error: %v", err)
		}
		if want := `{"amount":"-1234.56"}`; string(got) != want {
			t.Fatalf("json.Marshal = %s, want %s", got, want)
		}
	})

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		for _, s := range []string{"0", "1,5", "-0,00001", "92233720368547,75807"} {
			in := wrapper{Amount: mustComma(s)}
			data, err := json.Marshal(in)
			if err != nil {
				t.Fatalf("json.Marshal(%q) returned error: %v", s, err)
			}
			var out wrapper
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("json.Unmarshal(%s) returned error: %v", data, err)
			}
			if out != in {
				t.Fatalf("round trip of %q = %s, want %s", s, out.Amount, in.Amount)
			}
		}
	})

	t.Run("accepts number and null", func(t *testing.T) {
		t.Parallel()

		var out wrapper
		if err := json.Unmarshal([]byte(`{"amount":12.5}`), &out); err != nil {
			t.Fatalf("json.Unmarshal of a number returned error: %v", err)
		}
		if want := mustComma("12,5"); out.Amount != want {
			t.Fatalf("json.Unmarshal of a number = %s, want %s", out.Amount, want)
		}

		out.Amount = mustComma("7")
		if err := json.Unmarshal([]byte(`{"amount":null}`), &out); err != nil {
			t.Fatalf("json.Unmarshal of null returned error: %v", err)
		}
		if want := mustComma("7"); out.Amount != want {
			t.Fatalf("json.Unmarshal of null = %s, want %s (unchanged)", out.Amount, want)
		}
	})

	t.Run("rejects invalid", func(t *testing.T) {
		t.Parallel()

		var out wrapper
		err := json.Unmarshal([]byte(`{"amount":"1.234567"}`), &out)
		if !errors.Is(err, dsfinvk.ErrSyntax) {
			t.Fatalf("json.Unmarshal error = %v, want ErrSyntax", err)
		}
		if err := json.Unmarshal([]byte(`{"amount":true}`), &out); !errors.Is(err, dsfinvk.ErrSyntax) {
			t.Fatalf("json.Unmarshal of a bool error = %v, want ErrSyntax", err)
		}
	})
}

// FuzzParseFormat checks that every value ParseComma accepts survives a
// round trip through each output form.
func FuzzParseFormat(f *testing.F) {
	seeds := []string{
		"0", "0,00", "-0", "1234,56", "-0,5", "12,34567",
		"92233720368547,75807", "-92233720368547,75808", "000123,4",
		"1,234567", "+1", "", " 1", "1.234,56", "1e3", ",", "-", "abc",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		d, err := dsfinvk.ParseComma(s)
		if err != nil {
			t.Skip()
		}

		back, err := dsfinvk.ParseComma(d.FormatComma(dsfinvk.DecimalScale))
		if err != nil {
			t.Fatalf("ParseComma(FormatComma(%q)) returned error: %v", s, err)
		}
		if back != d {
			t.Fatalf("round trip of %q via FormatComma = %s, want %s", s, back, d)
		}

		for _, form := range []struct {
			name string
			out  string
		}{
			{"FormatDot", d.FormatDot(dsfinvk.DecimalScale)},
			{"String", d.String()},
			{"FormatProcessData", d.FormatProcessData()},
			{"FormatQuantity", d.FormatQuantity()},
		} {
			if _, err := dsfinvk.ParseDot(form.out); err != nil {
				t.Fatalf("ParseDot(%s(%q) = %q) returned error: %v", form.name, s, form.out, err)
			}
		}

		for scale := 0; scale <= dsfinvk.DecimalScale; scale++ {
			if out := d.FormatComma(scale); isNegativeZero(out) {
				t.Fatalf("%q.FormatComma(%d) = %q, want no sign on zero", s, scale, out)
			}
			if out := d.FormatDot(scale); isNegativeZero(out) {
				t.Fatalf("%q.FormatDot(%d) = %q, want no sign on zero", s, scale, out)
			}
		}
		if out := d.FormatProcessData(); isNegativeZero(out) {
			t.Fatalf("%q.FormatProcessData() = %q, want no sign on zero", s, out)
		}
		if out := d.FormatQuantity(); isNegativeZero(out) {
			t.Fatalf("%q.FormatQuantity() = %q, want no sign on zero", s, out)
		}
	})
}

func isNegativeZero(s string) bool {
	if !strings.HasPrefix(s, "-") {
		return false
	}
	return strings.Trim(s, "-0,.") == ""
}

func TestParseOverflowInFraction(t *testing.T) {
	t.Parallel()

	tests := []string{
		"184467440737095516,1",  // overflows while scaling the fraction
		"18446744073709551617",  // overflows on the final digit, not the shift
		"-18446744073709551617", // same, negative
	}

	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			t.Parallel()

			if _, err := dsfinvk.ParseComma(in); !errors.Is(err, dsfinvk.ErrOverflow) {
				t.Fatalf("ParseComma(%q) error = %v, want ErrOverflow", in, err)
			}
		})
	}
}

// TestMulQuotientFillsUint64 pins the guard for a 128-bit rounded quotient that
// no longer fits into 64 bits.
func TestMulQuotientFillsUint64(t *testing.T) {
	t.Parallel()

	a := mustComma("1452951435,58111")
	b := mustComma("126960,5")

	got, err := a.Mul(b)
	if !errors.Is(err, dsfinvk.ErrOverflow) {
		t.Fatalf("%s.Mul(%s) = (%s, %v), want ErrOverflow", a, b, got, err)
	}
}

func TestUnmarshalJSONDirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"bad escape", `"\q"`},
		{"lone quote", `"`},
		{"unterminated string", `"1.5`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d dsfinvk.Decimal
			if err := d.UnmarshalJSON([]byte(tt.in)); !errors.Is(err, dsfinvk.ErrSyntax) {
				t.Fatalf("UnmarshalJSON(%q) error = %v, want ErrSyntax", tt.in, err)
			}
			if !d.IsZero() {
				t.Fatalf("UnmarshalJSON(%q) modified the receiver to %s", tt.in, d)
			}
		})
	}
}
