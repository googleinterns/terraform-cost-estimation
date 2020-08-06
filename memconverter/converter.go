// Package memconverter holds the conversion values to base unit (byte) for all
// known units and makes the conversion to any unit.
package memconverter

var kb float64 = 1000
var mb float64 = 1000000
var gb float64 = 1000000000
var tb float64 = 1000000000000
var pb float64 = 1000000000000000

var kib float64 = 1024
var mib float64 = 1048576
var gib float64 = 1073741824
var tib float64 = 1099511627776
var pib float64 = 1125899906842624

func fromGB(num float64, unit string) (rez float64, success bool) {
	success = true
	switch unit {
	case "B":
		rez = num * gb
	case "KB":
		rez = num * gb / kb
	case "MB":
		rez = num * gb / mb
	case "GB":
		rez = num
	case "TB":
		rez = num * gb / tb
	case "PB":
		rez = num * gb / pb
	case "KiB":
		rez = num * gb / kib
	case "MiB":
		rez = num * gb / mib
	case "GiB":
		rez = num * gb / gib
	case "TiB":
		rez = num * gb / tib
	case "PiB":
		rez = num * gb / pib
	default:
		success = false
	}

	return rez, success
}

func toGB(num float64, unit string) (rez float64, success bool) {
	success = true
	switch unit {
	case "B":
		rez = num / gb
	case "KB":
		rez = num * kb / gb
	case "MB":
		rez = num * mb / gb
	case "GB":
		rez = num
	case "TB":
		rez = num * tb / gb
	case "PB":
		rez = num * pb / gb
	case "KiB":
		rez = num * kib / gb
	case "MiB":
		rez = num * mib / gb
	case "GiB":
		rez = num * gib / gb
	case "TiB":
		rez = num * tib / gb
	case "PiB":
		rez = num * pib / gb
	default:
		success = false
	}

	return rez, success
}

// Convert takes an initial unit, its number and the desired unit and returns
// the final number of that unit and the state of conversion (successful or not).
func Convert(from string, num float64, to string) (float64, bool) {
	gbnum, s1 := toGB(num, from)
	rez, s2 := fromGB(gbnum, to)
	return rez, (s1 && s2)
}
