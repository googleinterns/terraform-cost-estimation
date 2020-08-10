package catalog

import "testing"

var (
	pN2CoreUS = "Preemptible N2 Instance Core running in Americas"
	n1CoreUS  = "Predefined N1 Instance Core running in Americas"

	pN2RAMUS = "Preemptible N2 Instance Ram running in Americas"
	n2RAMUS  = "N2 Instance Ram running in Americas"

	computeEngine = "Compute Engine"
	compute       = "Compute"
	cpu           = "CPU"
	ram           = "RAM"
	preemptible   = "Preemptible"
	demand        = "OnDemand"
	usRegions     = []string{"us-central1", "us-east1", "us-west1"}

	core1 = MakeSKU(pN2CoreUS, computeEngine, compute, cpu, preemptible, usRegions, "hour", "USD", 7649000)
	core2 = MakeSKU(n1CoreUS, computeEngine, compute, cpu, demand, usRegions, "hour", "USD", 546789)

	ram1 = MakeSKU(pN2RAMUS, computeEngine, compute, ram, preemptible, usRegions, "gibibyte hour", "USD", 444600)
	ram2 = MakeSKU(n2RAMUS, computeEngine, compute, ram, demand, usRegions, "gibibyte hour", "USD", 6567435)
)

func TestIsMatch(t *testing.T) {
	var tests = []struct {
		sku         ComputeEngineSKU
		description []string
		region      string
		match       bool
	}{
		{core1, []string{"N2"}, "us-central1", true},
		{core1, []string{"N2", "Instance Core"}, "us-east1", true},
		{core1, []string{"N2", "Instance Core", "Preemptible"}, "us-east1", true},
		{core1, []string{"N2"}, "us-central2", false},
		{core1, []string{"N1", "Instance Core", "Preemptible"}, "us-central1", false},
		{ram1, []string{"N2"}, "us-central1", true},
		{ram1, []string{"N1"}, "us-central1", false},
		{ram1, []string{"N2", "Preemptible"}, "us-central2", false},
		{ram1, []string{"N2", "Preemptible"}, "us-west1", true},
		{ram1, []string{"N2", "Preemptible", "Instance Ram"}, "us-central1", true},
		{ram1, []string{"N2", "Preemptible", "Americas"}, "us-central1", true},
	}

	for _, test := range tests {
		match := test.sku.IsMatch(test.description, test.region)
		if match != test.match {
			t.Errorf("%+v.IsMatch(%+v, %+v) = %t ; want %t",
				test.sku, test.description, test.region, match, test.match)
		}
	}
}

func TestGetPricingInfo(t *testing.T) {
	var tests = []struct {
		sku                  ComputeEngineSKU
		usageUnitDescription string
		currencyCode         string
		nanos                int64
	}{
		{core1, "hour", "USD", 7649000},
		{core2, "hour", "USD", 546789},
		{ram1, "gibibyte hour", "USD", 444600},
		{ram2, "gibibyte hour", "USD", 6567435},
	}

	for _, test := range tests {
		usage, currency, nanos := test.sku.GetPricingInfo()
		if usage != test.usageUnitDescription || currency != test.currencyCode || nanos != test.nanos {
			t.Errorf("%+v.GetPricingInfo() = %s, %s, %v ; want %s, %s, %v",
				test.sku, usage, currency, nanos, test.usageUnitDescription, test.currencyCode, test.nanos)
		}
	}
}
