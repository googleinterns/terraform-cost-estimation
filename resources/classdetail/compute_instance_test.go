package classdetail

import (
	"fmt"
	"math"
	"testing"
)

const epsilon = 1e-10

func TestGetMachineDetails(t *testing.T) {
	tests := []struct {
		machineType string
		cores       int
		mem         float64
		err         error
	}{
		{"n1-standard-1", 1, 3.75, nil},
		{"e2-highmem-16", 16, 128, nil},
		{"c2-standard-8", 8, 32, nil},
		{"n2d-highcpu-128", 128, 128, nil},
		{"e2-micro", 2, 1, nil},
		{"g1-small", 1, 1.7, nil},
		{"m2-ultramem-208", 208, 5888, nil},
		{"n2-standard-3", 0, 0, fmt.Errorf("machine type not supported")},
		{"n1-highm-16", 0, 0, fmt.Errorf("machine type not supported")},
		{"e2-nano", 0, 0, fmt.Errorf("machine type not supported")},
	}

	for _, test := range tests {
		c, m, e := GetMachineDetails(test.machineType)

		fail1 := (e == nil && test.err != nil) || (e != nil && test.err == nil)
		fail2 := e != nil && test.err != nil && e.Error() != test.err.Error()
		fail3 := c != test.cores || math.Abs(m-test.mem) > epsilon
		if fail1 || fail2 || fail3 {
			t.Errorf("GetMachineDetails(%s) = %+v, %+v, %+v; want %+v, %+v, %+v",
				test.machineType, c, m, e, test.cores, test.mem, test.err)
		}
	}
}
