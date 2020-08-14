// Package memconverter holds the conversion values to base unit (byte) for all
// known units and makes the conversion to any unit.
package memconverter

import "fmt"

var unitFactors map[string]float64 = map[string]float64{
	"b":        1,
	"kb":       1000,
	"mb":       1000000,
	"gb":       1000000000,
	"tb":       1000000000000,
	"pb":       1000000000000000,
	"kib":      1024,
	"mib":      1048576,
	"gib":      1073741824,
	"tib":      1099511627776,
	"pib":      1125899906842624,
	"byte":     1,
	"kilobyte": 1000,
	"megabyte": 1000000,
	"gigabyte": 1000000000,
	"terabyte": 1000000000000,
	"petabyte": 1000000000000000,
	"kibibyte": 1024,
	"mebibyte": 1048576,
	"gibibyte": 1073741824,
	"tebibyte": 1099511627776,
	"pebibyte": 1125899906842624,
}

// Convert takes an initial unit, its number and the desired unit and returns
// the final number of that unit and the state of conversion (successful or not).
func Convert(from string, num float64, to string) (rez float64, err error) {
	err = nil
	if fact1, found1 := unitFactors[from]; found1 {
		b := num * fact1

		if fact2, found2 := unitFactors[to]; found2 {
			rez = b / fact2
		} else {
			err = fmt.Errorf("unknown final unit %s", to)
		}
	} else {
		err = fmt.Errorf("unknown initial unit %s", from)
	}

	return rez, err
}
