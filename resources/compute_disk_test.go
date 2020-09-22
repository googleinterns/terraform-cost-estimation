package resources

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewComputeDisk(t *testing.T) {
	tests := []struct {
		name     string
		diskName string
		id       string
		diskType string
		zone     string
		image    string
		snapshot string
		size     int64
		disk     *ComputeDisk
		err      error
	}{
		{"wrong_zone", "", "", "", "us", "", "", 10,
			nil, fmt.Errorf("invalid zone format")},

		{"size_out_of_bounds", "", "", "pd-standard", "us-central1-a", "", "", 9,
			nil, fmt.Errorf("size is not in the valid range")},

		{"image_size_out_of_bounds", "", "", "pd-standard", "us-central1-a", "fedora-coreos-testing", "", 0,
			nil, fmt.Errorf("size is not in the valid range")},

		{"size_smaller_than_image", "", "", "pd-standard", "us-central1-a", "centos-7", "", 10,
			nil, fmt.Errorf("size should at least be the size of the specified image")},

		{"default_size", "", "", "pd-standard", "us-central1-a", "", "", 0,
			&ComputeDisk{Type: "pd-standard", Zone: "us-central1-a", Region: "us-central1", SizeGiB: 500}, nil},

		{"just_size", "", "", "pd-standard", "us-central1-a", "", "", 100,
			&ComputeDisk{Type: "pd-standard", Zone: "us-central1-a", Region: "us-central1", SizeGiB: 100}, nil},

		{"just_image", "", "", "pd-standard", "us-central1-a", "centos-7", "", 0,
			&ComputeDisk{Type: "pd-standard", Zone: "us-central1-a", Region: "us-central1", SizeGiB: 20}, nil},

		{"size_and_image", "", "", "pd-standard", "us-central1-a", "centos-7", "", 100,
			&ComputeDisk{Type: "pd-standard", Zone: "us-central1-a", Region: "us-central1", SizeGiB: 100}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := NewComputeDisk(test.diskName, test.id, test.diskType, test.zone, test.image, test.snapshot, test.size)
			// Test fails if errors have different values, messages or the return values are different.
			f1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
			f2 := err != nil && test.err != nil && err.Error() != test.err.Error()
			if f1 || f2 || !reflect.DeepEqual(d, test.disk) {
				t.Errorf("NewComputeDisk(%s, %s, %s, %s, %s, %s, %d) = %+v, %+v; want %+v, %+v",
					test.diskName, test.id, test.diskType, test.zone, test.image, test.snapshot, test.size, d, err, test.disk, test.err)
			}
		})
	}
}
