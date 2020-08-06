package memconverter

import "testing"

func TestConvert(t *testing.T) {
	var tests = []struct {
		from    string
		num     float64
		to      string
		rez     float64
		success bool
	}{
		{"GB", 1, "GiB", 0.9313225746154785, true},
		{"TB", 1, "TiB", 0.9094947017729282, true},
		{"TB", 1, "GiB", 931.3225746154785, true},
		{"TiB", 1, "TB", 1.099511627776, true},
		{"TiB", 1, "GB", 1099.511627776, true},
		{"MB", 10000, "GB", 10, true},
		{"TB", 0.5, "MB", 500000, true},
		{"B", 100, "KB", 0.1, true},
		{"MiB", 123, "KiB", 125952, true},
		{"TiB", 0.00011730194091796875, "GiB", 0.1201171875, true},
		{"PB", 3, "MB", 3000000000, true},
		{"TiB", 178, "MiB", 186646528, true},
		{"GB", 1000, "PiB", 0.0008881784197001252, true},
		{"Mb", 178, "GB", 0, false},
		{"PB", 567, "Pib", 0, false},
	}

	for _, test := range tests {
		rez, success := Convert(test.from, test.num, test.to)
		if rez != test.rez || success != test.success {
			t.Errorf("Convert(%s, %f, %s) = %f, %t; want %f, %t", test.from, test.num, test.to, rez, success, test.rez, test.success)
		}
	}
}
