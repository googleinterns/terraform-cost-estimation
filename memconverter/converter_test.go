package memconverter

import (
	"fmt"
	"math"
	"testing"
)

const (
	epsilon = 1e-10

	u1 = float64(1000)
	u2 = float64(1024)

	b   = float64(1)
	kb  = b * u1
	mb  = kb * u1
	gb  = mb * u1
	tb  = gb * u1
	pb  = tb * u1
	kib = u2
	mib = kib * u2
	gib = mib * u2
	tib = gib * u2
	pib = tib * u2
)

func TestConvert(t *testing.T) {
	var tests = []struct {
		from string
		num  float64
		to   string
		rez  float64
		err  error
	}{
		{"gb", 1, "gib", 1 * gb / gib, nil},
		{"tb", 1, "tib", 1 * tb / tib, nil},
		{"tb", 1, "gib", 1 * tb / gib, nil},
		{"tib", 1, "tb", 1 * tib / tb, nil},
		{"tib", 1, "gb", 1 * tib / gb, nil},
		{"mb", 10000, "gb", 10000 * mb / gb, nil},
		{"tb", 0.5, "mb", 0.5 * tb / mb, nil},
		{"b", 100, "kb", 100 * b / kb, nil},
		{"mib", 123, "kib", 123 * mib / kib, nil},
		{"tib", 567, "gib", 567 * tib / gib, nil},
		{"pb", 3, "mb", 3 * pb / mb, nil},
		{"tib", 178, "mib", 178 * tib / mib, nil},
		{"gb", 1000, "pib", 1000 * gb / pib, nil},
		{"Mb", 178, "gb", 0, fmt.Errorf("unknown initial unit Mb")},
		{"pb", 567, "PIB", 0, fmt.Errorf("unknown final unit PIB")},
		{"gigabyte", 1, "gibibyte", 1 * gb / gib, nil},
		{"terabyte", 1, "tebibyte", 1 * tb / tib, nil},
		{"terabyte", 0.11, "gibibyte", 0.11 * tb / gib, nil},
		{"tebibyte", 1, "terabyte", 1 * tib / tb, nil},
		{"tebibyte", 1, "gigabyte", 1 * tib / gb, nil},
		{"megabyte", 10000, "gigabyte", 10000 * mb / gb, nil},
		{"terabyte", 0.5, "megabyte", 0.5 * tb / mb, nil},
		{"byte", 100, "kilobyte", 100 * b / kb, nil},
		{"mebibyte", 123, "kibibyte", 123 * mib / kib, nil},
		{"tebibyte", 567, "gibibyte", 567 * tib / gib, nil},
		{"petabyte", 3, "megabyte", 3 * pb / mb, nil},
		{"tebibyte", 178, "mebibyte", 178 * tib / mib, nil},
		{"gigabyte", 1000, "pebibyte", 1000 * gb / pib, nil},
	}

	for _, test := range tests {
		rez, err := Convert(test.from, test.num, test.to)
		pass1 := err == nil && test.err == nil && math.Abs(rez-test.rez) < epsilon
		pass2 := err != nil && test.err != nil && err.Error() == test.err.Error()
		if !(pass1 || pass2) {
			t.Errorf("Convert(%s, %f, %s) = %f, %s; want %f, %s", test.from, test.num, test.to, rez, err.Error(), test.rez, test.err.Error())
		}
	}
}
