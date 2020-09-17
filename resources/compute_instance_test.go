package resources

import (
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

func readSKU(path string) (*billingpb.Sku, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sku billingpb.Sku
	if err = jsonpb.UnmarshalString(string(data), &sku); err != nil {
		return nil, err
	}

	return &sku, nil
}

func readSKUs() ([]*billingpb.Sku, error) {
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/testdata/sku_%d.json"

	var skus []*billingpb.Sku
	for i := 0; i <= 3; i++ {
		sku, err := readSKU(fmt.Sprintf(inputPath, i))
		if err != nil {
			return nil, err
		}
		skus = append(skus, sku)
	}
	return skus, nil
}

func mapToDescription(skus []*billingpb.Sku) (mapped []string) {
	for _, sku := range skus {
		mapped = append(mapped, sku.Description)
	}
	return
}

func mapToPricingInfo(skus []*billingpb.Sku) (mapped []PricingInfo) {
	for _, sku := range skus {
		p := PricingInfo{}
		p.fillInfo(sku)
		mapped = append(mapped, p)
	}
	return
}

func TestCompletePricingInfo(t *testing.T) {
	skus, err := readSKUs()
	if err != nil {
		t.Fatal("Failed to read SKU JSON files")
	}

	c1 := CoreInfo{ResourceGroup: "CPU"}
	c2 := CoreInfo{ResourceGroup: "N1Standard"}
	c3 := c1
	m1 := MemoryInfo{ResourceGroup: "RAM"}
	m2 := MemoryInfo{ResourceGroup: "N1Standard"}
	m3 := m1

	p := mapToPricingInfo(skus)

	tests := []struct {
		name    string
		skuObj  skuObject
		skus    []*billingpb.Sku
		pricing PricingInfo
		err     error
	}{
		{"n1standard_core", &c2, []*billingpb.Sku{skus[0], skus[1], skus[2], skus[3]}, p[0], nil},
		{"n1standard_ram", &m2, []*billingpb.Sku{skus[0], skus[1], skus[2], skus[3]}, p[3], nil},
		{"cpu_resoure_group", &c1, []*billingpb.Sku{skus[0], skus[1], skus[2], skus[3]}, p[1], nil},
		{"ram_resource_group", &m1, []*billingpb.Sku{skus[0], skus[1], skus[2], skus[3]}, p[2], nil},
		{"no_core", &c3, []*billingpb.Sku{skus[2], skus[3]}, PricingInfo{}, fmt.Errorf("could not find core pricing information")},
		{"no_ram", &m3, []*billingpb.Sku{skus[0], skus[1]}, PricingInfo{}, fmt.Errorf("could not find memory pricing information")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.skuObj.completePricingInfo(test.skus)
			// Test fails if the error value is different or with different messages or if the pricing information differs from the expected one.
			f1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
			f2 := err != nil && test.err != nil && err.Error() != test.err.Error()
			f3 := test.pricing != test.skuObj.getPricingInfo()
			if f1 || f2 || f3 {
				t.Errorf("{%+v}.completePricingInfo(%+v) -> %+v, %+v; want %+v, %+v",
					test.skuObj, mapToDescription(test.skus), test.skuObj.getPricingInfo(), err, test.pricing, test.err)
			}
		})
	}
}

func TestCoreGetTotalPrice(t *testing.T) {
	c1 := CoreInfo{Number: 2, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000}}
	c2 := CoreInfo{Number: 4, Fractional: 0.125, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000}}
	c3 := CoreInfo{Number: 32, Fractional: 0.5, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733}}
	c4 := CoreInfo{Number: 16, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000}}

	tests := []struct {
		name  string
		core  CoreInfo
		price float64
	}{
		{"no_fractional_0", c1, float64(6980000) * 2 / nano},
		{"no_fractiona_1", c4, float64(2701000) * 16 / nano},
		{"fractional_0", c2, float64(44856000) * 4 / nano * 0.125},
		{"fractional_1", c3, float64(1121733) * 32 / nano * 0.5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.core.getTotalPrice(); math.Abs(actual-test.price) > epsilon {
				t.Errorf("{%+v}.getTotalPrice() = %f ; want %f", test.core, actual, test.price)
			}
		})
	}
}

func TestMemGetTotalPrice(t *testing.T) {
	m1 := MemoryInfo{AmountGiB: 100, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000, UsageUnit: "gigabyte hour"}}
	m2 := MemoryInfo{AmountGiB: 50, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000, UsageUnit: "pebibyte hour"}}
	m3 := MemoryInfo{AmountGiB: 320, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733, UsageUnit: "tebibyte hour"}}
	m4 := MemoryInfo{AmountGiB: 16, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "gibibyte hour"}}
	m5 := MemoryInfo{AmountGiB: 160, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "giBibyte hour"}}
	m6 := MemoryInfo{AmountGiB: 160, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "mebibite hour"}}

	gb := float64(1000 * 1000 * 1000)
	gib := float64(1024 * 1024 * 1024)

	tests := []struct {
		name  string
		mem   MemoryInfo
		price float64
		err   error
	}{
		{"gigabyte_unit", m1, float64(6980000) / nano * 100 * gib / gb, nil},
		{"pebibyte_unit", m2, float64(44856000) / nano * 50 / (1024 * 1024), nil},
		{"tebibyte_unit", m3, float64(1121733) / nano * 320 / 1024, nil},
		{"gibibyte_unit", m4, float64(2701000) / nano * 16, nil},
		{"wrong_unit_0", m5, 0, fmt.Errorf("unknown final unit giBibyte")},
		{"wrong_unit_1", m6, 0, fmt.Errorf("unknown final unit mebibite")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, err := test.mem.getTotalPrice()
			f1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
			f2 := err != nil && test.err != nil && err.Error() != test.err.Error()
			// Test fails if the error has a different value or message or if the return value is different than expected.
			if f1 || f2 || math.Abs(p-test.price) > epsilon {
				t.Errorf("{%+v}.getTotalPrice() = %f, %+v ; want %f, %+v", test.mem, p, err, test.price, test.err)
			}
		})
	}
}

func TestGetDelta(t *testing.T) {
	c1 := CoreInfo{Number: 4, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 12345}}
	m1 := MemoryInfo{AmountGiB: 1000, UnitPricing: PricingInfo{HourlyUnitPrice: 23455, UsageUnit: "gibibyte hour"}}
	i1 := ComputeInstance{Cores: c1, Memory: m1}

	c2 := CoreInfo{Number: 16, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 12345}}
	m2 := MemoryInfo{AmountGiB: 500, UnitPricing: PricingInfo{HourlyUnitPrice: 23455, UsageUnit: "gibibyte hour"}}
	i2 := ComputeInstance{Cores: c2, Memory: m2}

	c3 := CoreInfo{Number: 32, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 785678}}
	m3 := MemoryInfo{AmountGiB: 2000, UnitPricing: PricingInfo{HourlyUnitPrice: 235977, UsageUnit: "gigbyte hour"}}
	i3 := ComputeInstance{Name: "test", MachineType: "n1-standard-1", Cores: c3, Memory: m3}

	tests := []struct {
		name  string
		state ComputeInstanceState
		dcore float64
		dmem  float64
		err   error
	}{
		{"create", ComputeInstanceState{Before: nil, After: &i1}, 4 * 12345, 1000 * 23455, nil},
		{"destroy", ComputeInstanceState{Before: &i1, After: nil}, -4 * 12345, -1000 * 23455, nil},
		{"wrong_before", ComputeInstanceState{Before: &i3, After: &i2}, 0, 0, fmt.Errorf("test(n1-standard-1): unknown final unit gigbyte")},
		{"wrong_after", ComputeInstanceState{Before: &i1, After: &i3}, 0, 0, fmt.Errorf("test(n1-standard-1): unknown final unit gigbyte")},
		{"update_0", ComputeInstanceState{Before: &i1, After: &i2}, (16 - 4) * 12345, (500 - 1000) * 23455, nil},
		{"update_1", ComputeInstanceState{Before: &i2, After: &i1}, -(16 - 4) * 12345, -(500 - 1000) * 23455, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dcore, dmem, err := test.state.getDelta()
			f1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
			f2 := err != nil && test.err != nil && err.Error() != test.err.Error()
			// Test fails if the error value is different or with different messages or if the return values differs from the expected ones.
			if f1 || f2 || math.Abs(dcore-test.dcore/nano) > epsilon || math.Abs(dmem-test.dmem/nano) > epsilon {
				t.Errorf("%+v.getDelta() = %f, %f, %s ; want %f, %f, %s",
					test.state, dcore, dmem, err, test.dcore, test.dmem, test.err)
			}
		})
	}
}
