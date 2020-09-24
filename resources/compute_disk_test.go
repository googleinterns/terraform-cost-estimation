package resources

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestNewComputeDisk(t *testing.T) {
	tests := []struct {
		name     string
		diskName string
		id       string
		diskType string
		zones    []string
		image    string
		snapshot string
		size     int64
		disk     *ComputeDisk
		err      error
	}{
		{"wrong_zone", "", "", "", []string{"us"}, "", "", 10,
			nil, fmt.Errorf("invalid zone format")},

		{"size_out_of_bounds", "", "", "pd-standard", []string{"us-central1-a"}, "", "", 9,
			nil, fmt.Errorf("size is not in the valid range")},

		{"image_size_out_of_bounds", "", "", "pd-standard", []string{"us-central1-a"}, "fedora-coreos-testing", "", 0,
			nil, fmt.Errorf("size is not in the valid range")},

		{"size_smaller_than_image", "", "", "pd-standard", []string{"us-central1-a"}, "centos-7", "", 10,
			nil, fmt.Errorf("size should at least be the size of the specified image")},

		{"default_size", "", "", "pd-standard", []string{"us-central1-a"}, "", "", 0,
			&ComputeDisk{Type: "pd-standard", Description: Description{Contains: []string{"Storage PD Capacity"}, Omits: []string{"Regional"}},
				Zones: []string{"us-central1-a"}, Region: "us-central1", SizeGiB: 500}, nil},

		{"just_size", "", "", "pd-standard", []string{"us-central1-a"}, "", "", 100,
			&ComputeDisk{Type: "pd-standard", Description: Description{Contains: []string{"Storage PD Capacity"}, Omits: []string{"Regional"}},
				Zones: []string{"us-central1-a"}, Region: "us-central1", SizeGiB: 100}, nil},

		{"just_image", "", "", "pd-standard", []string{"us-central1-a"}, "centos-7", "", 0,
			&ComputeDisk{Type: "pd-standard", Description: Description{Contains: []string{"Storage PD Capacity"}, Omits: []string{"Regional"}},
				Zones: []string{"us-central1-a"}, Region: "us-central1", SizeGiB: 20}, nil},

		{"size_and_image", "", "", "pd-standard", []string{"us-central1-a"}, "centos-7", "", 100,
			&ComputeDisk{Type: "pd-standard", Description: Description{Contains: []string{"Storage PD Capacity"}, Omits: []string{"Regional"}},
				Zones: []string{"us-central1-a"}, Region: "us-central1", SizeGiB: 100}, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := NewComputeDisk(test.diskName, test.id, test.diskType, test.zones, test.image, test.snapshot, test.size)
			// Test fails if errors have different values, messages or the return values are different.
			f1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
			f2 := err != nil && test.err != nil && err.Error() != test.err.Error()
			if f1 || f2 || !reflect.DeepEqual(d, test.disk) {
				t.Errorf("NewComputeDisk(%s, %s, %s, %s, %s, %s, %d) = %+v, %+v; want %+v, %+v",
					test.diskName, test.id, test.diskType, test.zones, test.image, test.snapshot, test.size, d, err, test.disk, test.err)
			}
		})
	}
}

func TestDiskTotalPrice(t *testing.T) {
	monthlyToHourly := 1.0 / float64(30*24)

	tests := []struct {
		name string
		d    *ComputeDisk
		tot  float64
		err  error
	}{
		{"wrong_unit", &ComputeDisk{SizeGiB: 10, UnitPricing: PricingInfo{HourlyUnitPrice: 1000, UsageUnit: "gibibite"}},
			0, fmt.Errorf("unknown final unit gibibite")},

		{"test_0", &ComputeDisk{SizeGiB: 200, UnitPricing: PricingInfo{HourlyUnitPrice: 1234, UsageUnit: "gibibyte"}},
			200 * 1234 / nano * monthlyToHourly, nil},

		{"test_1", &ComputeDisk{SizeGiB: 50, UnitPricing: PricingInfo{HourlyUnitPrice: 5678, UsageUnit: "gibibyte"}},
			50 * 5678 / nano * monthlyToHourly, nil},
		{"test_2", &ComputeDisk{SizeGiB: 400, UnitPricing: PricingInfo{HourlyUnitPrice: 23456, UsageUnit: "gibibyte"}},
			400 * 23456 / nano * monthlyToHourly, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tot, err := test.d.totalPrice()
			// Test fails if errors have different values, messages or the return values are different.
			f1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
			f2 := err != nil && test.err != nil && err.Error() != test.err.Error()
			if f1 || f2 || math.Abs(tot-test.tot) > epsilon {
				t.Errorf("disk.totalPrice() = %f, %+v; want %f, %+v", tot, err, test.tot, test.err)
			}
		})
	}
}

func TestDiskStateGeneralChanges(t *testing.T) {
	d1 := &ComputeDisk{Name: "test1", Type: "pd-standard", Zones: []string{"us-central1-a", "us-central1-b"}, Image: "centos-7"}
	d2 := &ComputeDisk{Name: "test2", Type: "pd-ssd", Zones: []string{"us-central1-c", "us-central1-b"}, Image: "centos-7", ID: "1234567"}
	d3 := &ComputeDisk{Name: "test2", Type: "pd-standard", Zones: []string{"us-central1-c", "us-central1-b"}, Image: "centos-7", ID: "1234567"}

	tests := []struct {
		name       string
		state      *ComputeDiskState
		nameChange string
		id         string
		action     string
		diskType   string
		zones      string
		image      string
		snapshot   string
	}{
		{"create", &ComputeDiskState{Action: "create", Before: nil, After: d1},
			"test1", "unknown", "create", "pd-standard", "us-central1-a, us-central1-b", "centos-7", ""},

		{"destroy", &ComputeDiskState{Action: "destroy", Before: d2, After: nil},
			"test2", "1234567", "destroy", "pd-ssd", "us-central1-b, us-central1-c", "centos-7", ""},

		{"update_all_changes", &ComputeDiskState{Action: "update", Before: d2, After: d1},
			"test2 -> test1", "1234567", "update", "pd-ssd -> pd-standard", "us-central1-b, us-central1-c -> us-central1-a, us-central1-b", "centos-7", ""},

		{"update_type", &ComputeDiskState{Action: "update", Before: d2, After: d3},
			"test2", "1234567", "update", "pd-ssd -> pd-standard", "us-central1-b, us-central1-c", "centos-7", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, id, action, diskType, zones, image, snapshot := test.state.generalChanges()
			// Test fails if the return values are different.
			f1 := name != test.nameChange || id != test.id || action != test.action || diskType != test.diskType
			f2 := zones != test.zones || image != test.image || snapshot != test.snapshot
			if f1 || f2 {
				t.Errorf("state.generalChanges() = %s, %s, %s, %s, %s, %s, %s ; want %s, %s, %s, %s, %s, %s, %s",
					name, id, action, diskType, zones, image, snapshot,
					test.nameChange, test.id, test.action, test.diskType, test.zones, test.image, test.snapshot)
			}
		})
	}
}

func TestDiskStateCostChanges(t *testing.T) {
	d1 := &ComputeDisk{UnitPricing: PricingInfo{HourlyUnitPrice: 0.1, UsageUnit: "gibibite month"}}
	d2 := &ComputeDisk{SizeGiB: 150, UnitPricing: PricingInfo{HourlyUnitPrice: 0.1, UsageUnit: "gibibyte"}}
	d3 := &ComputeDisk{SizeGiB: 500, UnitPricing: PricingInfo{HourlyUnitPrice: 0.3, UsageUnit: "gibibyte"}}

	tests := []struct {
		name         string
		state        *ComputeDiskState
		costPerUnit1 float64
		costPerUnit2 float64
		units1       int64
		units2       int64
		delta        float64
		err          error
	}{
		{"invalid_memory_unit", &ComputeDiskState{Before: d1, After: d1}, 0, 0, 0, 0, 0, fmt.Errorf("invalid final unit gibibite")},
		{"create", &ComputeDiskState{Before: nil, After: d2}, 0, 0.1, 0, 150, 0.1 * 150, nil},
		{"destroy", &ComputeDiskState{Before: d2, After: nil}, 0.1, 0, 150, 0, -0.1 * 150, nil},
		{"update", &ComputeDiskState{Before: d2, After: d3}, 0.1, 0.3, 150, 500, 0.3*500 - 0.1*150, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			costPerUnit1, costPerUnit2, units1, units2, delta, err := test.state.costChanges()
			// Test fails if error or return values are different.
			f1 := costPerUnit1 != test.costPerUnit1 || costPerUnit2 != test.costPerUnit2
			f2 := units1 != test.units1 || units2 != test.units2 || delta != test.delta
			if f1 || f2 {
				t.Errorf("state.costChanges() = %f, %f, %d, %d, %f, %+v ; want %f, %f, %d, %d, %f, %+v",
					costPerUnit1, costPerUnit2, units1, units2, delta, err,
					test.costPerUnit1, test.costPerUnit2, test.units1, test.units2, test.delta, test.err)
			}
		})
	}
}
