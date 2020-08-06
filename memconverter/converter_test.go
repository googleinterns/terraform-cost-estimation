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
		{"GB", 1, "GiB", 0.9313225746154785, nil},
		{"TB", 1, "TiB", 0.9094947017729282, nil},
		{"TB", 1, "GiB", 931.3225746154785, nil},
		{"TiB", 1, "TB", 1.099511627776, nil},
		{"TiB", 1, "GB", 1099.511627776, nil},
		{"MB", 10000, "GB", 10, nil},
		{"TB", 0.5, "MB", 500000, nil},
		{"B", 100, "KB", 0.1, nil},
		{"MiB", 123, "KiB", 125952, nil},
		{"TiB", 0.00011730194091796875, "GiB", 0.1201171875, nil},
		{"PB", 3, "MB", 3000000000, nil},
		{"TiB", 178, "MiB", 186646528, nil},
		{"GB", 1000, "PiB", 0.0008881784197001252, nil},
		{"Mb", 178, "GB", 0, fmt.Errorf("unknown initial unit Mb")},
		{"PB", 567, "Pib", 0, fmt.Errorf("unknown final unit Pib")},
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
