package memconverter

import (
	"fmt"
	"testing"
)

func TestConvert(t *testing.T) {
	var tests = []struct {
		from string
		num  float64
		to   string
		rez  float64
		err  error
	}{
		{"gb", 1, "gib", 0.9313225746154785, nil},
		{"tb", 1, "tib", 0.9094947017729282, nil},
		{"tb", 1, "gib", 931.3225746154785, nil},
		{"tib", 1, "tb", 1.099511627776, nil},
		{"tib", 1, "gb", 1099.511627776, nil},
		{"mb", 10000, "gb", 10, nil},
		{"tb", 0.5, "mb", 500000, nil},
		{"b", 100, "kb", 0.1, nil},
		{"mib", 123, "kib", 125952, nil},
		{"tib", 0.00011730194091796875, "gib", 0.1201171875, nil},
		{"pb", 3, "mb", 3000000000, nil},
		{"tib", 178, "mib", 186646528, nil},
		{"gb", 1000, "pib", 0.0008881784197001252, nil},
		{"Mb", 178, "gb", 0, fmt.Errorf("unknown initial unit Mb")},
		{"pb", 567, "PIB", 0, fmt.Errorf("unknown final unit PIB")},
		{"gigabyte", 1, "gibibyte", 0.9313225746154785, nil},
		{"terabyte", 1, "tebibyte", 0.9094947017729282, nil},
		{"terabyte", 1, "gibibyte", 931.3225746154785, nil},
		{"tebibyte", 1, "terabyte", 1.099511627776, nil},
		{"tebibyte", 1, "gigabyte", 1099.511627776, nil},
		{"megabyte", 10000, "gigabyte", 10, nil},
		{"terabyte", 0.5, "megabyte", 500000, nil},
		{"byte", 100, "kilobyte", 0.1, nil},
		{"mebibyte", 123, "kibibyte", 125952, nil},
		{"tebibyte", 0.00011730194091796875, "gibibyte", 0.1201171875, nil},
		{"petabyte", 3, "megabyte", 3000000000, nil},
		{"tebibyte", 178, "mebibyte", 186646528, nil},
		{"gigabyte", 1000, "pebibyte", 0.0008881784197001252, nil},
	}

	for _, test := range tests {
		rez, err := Convert(test.from, test.num, test.to)
		pass1 := err == nil && test.err == nil && rez == test.rez
		pass2 := err != nil && test.err != nil && err.Error() == test.err.Error()
		if !(pass1 || pass2) {
			t.Errorf("Convert(%s, %f, %s) = %f, %s; want %f, %s", test.from, test.num, test.to, rez, err.Error(), test.rez, test.err.Error())
		}
	}
}
