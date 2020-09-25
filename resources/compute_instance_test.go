package resources

import (
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"reflect"
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

func mapToPricingInfo(skus []*billingpb.Sku, correctTierRate func(*billingpb.PricingExpression_TierRate) bool) (mapped []PricingInfo) {
	for _, sku := range skus {
		p := PricingInfo{}
		p.fillHourlyBase(sku, correctTierRate)
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

	p := mapToPricingInfo(skus, func(*billingpb.PricingExpression_TierRate) bool { return true })

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
			if !reflect.DeepEqual(err, test.err) || test.pricing != test.skuObj.getPricingInfo() {
				t.Errorf("{%+v}.completePricingInfo(%+v) -> %+v, %+v; want %+v, %+v",
					test.skuObj, mapToDescription(test.skus), test.skuObj.getPricingInfo(), err, test.pricing, test.err)
			}
		})
	}
}

func TestCoreGetTotalPrice(t *testing.T) {
	c1 := CoreInfo{Number: 2, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 0.06}}
	c2 := CoreInfo{Number: 4, Fractional: 0.125, UnitPricing: PricingInfo{HourlyUnitPrice: 0.44}}
	c3 := CoreInfo{Number: 32, Fractional: 0.5, UnitPricing: PricingInfo{HourlyUnitPrice: 0.101}}
	c4 := CoreInfo{Number: 16, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 2.7}}

	tests := []struct {
		name  string
		core  CoreInfo
		price float64
	}{
		{"no_fractional_0", c1, 0.06 * 2},
		{"no_fractiona_1", c4, 2.7 * 16},
		{"fractional_0", c2, 0.44 * 4 * 0.125},
		{"fractional_1", c3, 0.101 * 32 * 0.5},
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
	m1 := MemoryInfo{AmountGiB: 100, UnitPricing: PricingInfo{HourlyUnitPrice: 0.06, UsageUnit: "gigabyte"}}
	m2 := MemoryInfo{AmountGiB: 50, UnitPricing: PricingInfo{HourlyUnitPrice: 0.44, UsageUnit: "pebibyte"}}
	m3 := MemoryInfo{AmountGiB: 320, UnitPricing: PricingInfo{HourlyUnitPrice: 0.101, UsageUnit: "tebibyte"}}
	m4 := MemoryInfo{AmountGiB: 16, UnitPricing: PricingInfo{HourlyUnitPrice: 2.7, UsageUnit: "gibibyte"}}

	gb := float64(1000 * 1000 * 1000)
	gib := float64(1024 * 1024 * 1024)

	tests := []struct {
		name  string
		mem   MemoryInfo
		price float64
	}{
		{"gigabyte_unit", m1, 0.06 * 100 * gib / gb},
		{"pebibyte_unit", m2, 0.44 * 50 / (1024 * 1024)},
		{"tebibyte_unit", m3, 0.101 * 320 / 1024},
		{"gibibyte_unit", m4, 2.7 * 16},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if p := test.mem.getTotalPrice(); math.Abs(p-test.price) > epsilon {
				t.Errorf("{%+v}.getTotalPrice() = %f ; want %f", test.mem, p, test.price)
			}
		})
	}
}

func TestGetDeltas(t *testing.T) {
	c1 := CoreInfo{Number: 4, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 0.12345}}
	m1 := MemoryInfo{AmountGiB: 1000, UnitPricing: PricingInfo{HourlyUnitPrice: 0.23455, UsageUnit: "gibibyte"}}
	i1 := ComputeInstance{Cores: c1, Memory: m1}

	c2 := CoreInfo{Number: 16, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 0.12345}}
	m2 := MemoryInfo{AmountGiB: 500, UnitPricing: PricingInfo{HourlyUnitPrice: 0.23455, UsageUnit: "gibibyte"}}
	i2 := ComputeInstance{Cores: c2, Memory: m2}

	tests := []struct {
		name  string
		state ComputeInstanceState
		dcore float64
		dmem  float64
	}{
		{"create", ComputeInstanceState{Before: nil, After: &i1}, 4 * 0.12345, 1000 * 0.23455},
		{"destroy", ComputeInstanceState{Before: &i1, After: nil}, -4 * 0.12345, -1000 * 0.23455},
		{"update_0", ComputeInstanceState{Before: &i1, After: &i2}, (16 - 4) * 0.12345, (500 - 1000) * 0.23455},
		{"update_1", ComputeInstanceState{Before: &i2, After: &i1}, -(16 - 4) * 0.12345, -(500 - 1000) * 0.23455},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if dcore, dmem := test.state.getDeltas(); math.Abs(dcore-test.dcore) > epsilon || math.Abs(dmem-test.dmem) > epsilon {
				t.Errorf("%+v.getDelta() = %f, %f; want %f, %f",
					test.state, dcore, dmem, test.dcore, test.dmem)
			}
		})
	}
}
