package disk

import (
	"fmt"
	"testing"
)

func TestDetails(t *testing.T) {
	tests := []struct {
		name     string
		diskType string
		zone     string
		region   string
		def      int64
		min      int64
		max      int64
		err      error
	}{
		{"invalid_type", "pd", "", "", 0, 0, 0, fmt.Errorf("invalid disk type 'pd'")},
		{"invalid_location", "pd-standard", "", "", 0, 0, 0, fmt.Errorf("no disk type running in ''")},
		{"region_0", "pd-standard", "", "us-central1", 500, 200, 65536, nil},
		{"region_1", "pd-ssd", "", "us-central1", 100, 10, 65536, nil},
		{"region_2", "pd-balanced", "", "us-central1", 100, 10, 65536, nil},
		{"zone_0", "local-ssd", "us-central1-a", "", 375, 375, 375, nil},
		{"zone_1", "pd-balanced", "us-central1-b", "", 100, 10, 65536, nil},
		{"zone_2", "pd-standard", "us-west1-b", "", 500, 10, 65536, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			def, min, max, err := Details(test.diskType, test.zone, test.region)
			// Test fails if the errors have different values, messages or the return values differ.
			f1 := (test.err == nil && err != nil) || (test.err != nil && err == nil)
			f2 := test.err != nil && err != nil && test.err.Error() != err.Error()
			f3 := def != test.def || min != test.min || max != test.max
			if f1 || f2 || f3 {
				t.Errorf("Details(%s, %s, %s) = %d, %d, %d, %+v ; want %d, %d, %d, %+v",
					test.diskType, test.zone, test.region, def, min, max, err, test.def, test.min, test.max, test.err)
			}
		})
	}
}
