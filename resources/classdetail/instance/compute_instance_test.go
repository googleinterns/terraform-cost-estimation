package instance

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const epsilon = 1e-10

func TestGetMachineDetails(t *testing.T) {
	machineTypes, err := ReadMachineTypes()
	if err != nil {
		t.Fatal("could not read machine type information")
	}

	tests := []struct {
		name        string
		machineType string
		cores       int
		mem         float64
		err         error
	}{
		{"predefined_0", "n1-standard-1", 1, 3.75, nil},
		{"predefined_1", "e2-highmem-16", 16, 128, nil},
		{"predefined_2", "c2-standard-8", 8, 32, nil},
		{"predefined_3", "n2d-highcpu-128", 128, 128, nil},
		{"predefined_4", "m2-ultramem-208", 208, 5888, nil},
		{"shared_core_0", "e2-micro", 2, 1, nil},
		{"shared_core_1", "g1-small", 1, 1.7, nil},
		{"unknown_0", "n2-standard-3", 0, 0, fmt.Errorf("machine type not supported")},
		{"unkown_1", "n1-highm-16", 0, 0, fmt.Errorf("machine type not supported")},
		{"unknown_2", "e2-nano", 0, 0, fmt.Errorf("machine type not supported")},
		{"custom_0", "custom-2-1024", 2, 1, nil},
		{"custom_1", "e2-custom-4-2048", 4, 2, nil},
		{"custom_2", "n2d-custom-2-1280", 2, 1.25, nil},
		{"extended_memory", "custom-2-1024-ext", 2, 1, nil},
		{"wrong_custom_0", "custom-2", 0, 0, fmt.Errorf("invalid custom machine type format")},
		{"wrong_custom_1", "custom-2-ext", 0, 0, fmt.Errorf("invalid custom machine type format")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, m, e := GetMachineDetails(machineTypes, test.machineType)
			// Test fails if the errors or the return values are different.
			if !reflect.DeepEqual(e, test.err) || c != test.cores || math.Abs(m-test.mem) > epsilon {
				t.Errorf("GetMachineDetails(%s) = %+v, %+v, %+v; want %+v, %+v, %+v",
					test.machineType, c, m, e, test.cores, test.mem, test.err)
			}
		})
	}
}
