package billing

import (
	"math"
	"testing"

	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

func TestFitsDescription(t *testing.T) {
	skus, err := readSKUs()
	if err != nil {
		t.Fatal("Failed to read SKU JSON files")
	}

	tests := []struct {
		name     string
		sku      *billingpb.Sku
		contains []string
		omits    []string
		ok       bool
	}{
		{"all_ok_0", skus[0], []string{"N1"}, []string{"N2"}, true},
		{"all_ok_1", skus[2], []string{"Licensing Fee"}, []string{"2017"}, true},
		{"all_ok_2", skus[4], []string{"Commitment", "Cpu"}, []string{"Preemptible"}, true},
		{"all_ok_3", skus[7], []string{"Network", "Vpn"}, []string{}, true},
		{"wrong_contains", skus[0], []string{"N1", "Preemptible"}, []string{}, false},
		{"wrong_omits", skus[6], []string{"Licensing Fee", "SQL"}, []string{"2012"}, false},
		{"wrong_contains_and_omits", skus[6], []string{"Licensing Fee", "SQL", "2013"}, []string{"2012"}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if ok := fitsDescription(test.sku, test.contains, test.omits); ok != test.ok {
				t.Errorf("sku.Description = %s, fitsDescription(sku, %+v, %+v) = %t; want %t",
					test.sku.Description, test.contains, test.omits, ok, test.ok)
			}
		})
	}
}

func TestFitsRegion(t *testing.T) {
	skus, err := readSKUs()
	if err != nil {
		t.Fatal("Failed to read SKU JSON files")
	}

	tests := []struct {
		name   string
		sku    *billingpb.Sku
		region string
		ok     bool
	}{
		{"single_region_0", skus[9], "asia-southeast1", true},
		{"single_region_1", skus[9], "asia-east1", false},
		{"more_regions_0", skus[0], "europe-west1", true},
		{"more_regions_1", skus[0], "europe-west3", true},
		{"more_regions_2", skus[0], "europe-west6", true},
		{"more_regions_3", skus[0], "europe-west5", false},
		{"more_regions_4", skus[0], "europe-east6", false},
		{"global", skus[3], "europe-north1", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if ok := fitsRegion(test.sku, test.region); ok != test.ok {
				t.Errorf("sku.Description = %s, FitsRegion(sku, %s) = %t; want %t",
					test.sku.Description, test.region, ok, test.ok)
			}
		})
	}
}

func TestGetPricingInfo(t *testing.T) {
	const epsilon = 1e-10
	skus, err := readSKUs()
	if err != nil {
		t.Fatal("Failed to read SKU JSON files")
	}

	tests := []struct {
		name         string
		sku          *billingpb.Sku
		f            func(*billingpb.PricingExpression_TierRate) bool
		usageUnit    string
		pricePerUnit float64
		currencyType string
	}{
		{"no_pricing", skus[6], func(*billingpb.PricingExpression_TierRate) bool { return true }, "hour", 0, ""},
		{"one_pricing", skus[0], func(*billingpb.PricingExpression_TierRate) bool { return true }, "gibibyte", 5928000 / nano, "USD"},
		{"more_pricing", skus[5], func(*billingpb.PricingExpression_TierRate) bool { return true }, "gibibyte", 5226000 / nano, "USD"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			usageUnit, pricePerUnit, currencyType := PricingInfo(test.sku, test.f)
			// Test fails if any return value is different than the expected one.
			if usageUnit != test.usageUnit || math.Abs(pricePerUnit-test.pricePerUnit) > epsilon || currencyType != test.currencyType {
				t.Errorf("GetPricingInfo(sku) = %+v, %+v, %+v; want %+v, %+v, %+v",
					usageUnit, pricePerUnit, currencyType,
					test.usageUnit, test.pricePerUnit, test.currencyType)
			}
		})
	}
}
