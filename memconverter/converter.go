// Package memconverter holds the conversion values to base unit (byte) for all
// known units and makes the conversion to any unit.
package memconverter

import "fmt"

var unitFactors map[string]float64 = map[string]float64{
	"B":   1,
	"KB":  1000,
	"MB":  1000000,
	"GB":  1000000000,
	"TB":  1000000000000,
	"PB":  1000000000000000,
	"KiB": 1024,
	"MiB": 1048576,
	"GiB": 1073741824,
	"TiB": 1099511627776,
	"PiB": 1125899906842624,
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
