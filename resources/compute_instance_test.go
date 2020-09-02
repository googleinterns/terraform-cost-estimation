package resources

import (
	"fmt"
	"math"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

var (
	str1 = `
	{
		"name": "services/6F81-5844-456A/skus/0048-21CE-74C3",
		"skuId": "0048-21CE-74C3",
		"description": "Preemptible N2 Instance Core running in Americas",
		"category": {
		  "serviceDisplayName": "Compute Engine",
		  "resourceFamily": "Compute",
		  "resourceGroup": "CPU",
		  "usageType": "Preemptible"
		},
		"serviceRegions": [
		  "us-central1",
		  "us-west1",
		  "us-east1"
		],
		"pricingInfo": [
        {
          "summary": "",
          "pricingExpression": {
            "usageUnit": "h",
            "usageUnitDescription": "hour",
            "baseUnit": "s",
            "baseUnitDescription": "second",
            "baseUnitConversionFactor": 3600,
            "displayQuantity": 1,
            "tieredRates": [
              {
                "startUsageAmount": 0,
                "unitPrice": {
                  "currencyCode": "USD",
                  "units": "0",
                  "nanos": 6980000
                }
              }
            ]
          },
          "currencyConversionRate": 1,
          "effectiveTime": "2020-08-05T01:48:54.819Z"
        }
      ]
	}
	`
	str2 = `
	{
		"description": "E2 Custom Instance Core running in Sao Paulo",
		"category": {
			"serviceDisplayName": "Compute Engine",
			"resourceFamily": "Compute",
			"resourceGroup": "CPU",
			"usageType": "OnDemand"
		},
		"serviceRegions": [
			"southamerica-east1",
			"southamerica-west1"
		],
		"pricingInfo": [
        {
          "summary": "",
          "pricingExpression": {
            "usageUnit": "h",
            "usageUnitDescription": "hour",
            "baseUnit": "s",
            "baseUnitDescription": "second",
            "baseUnitConversionFactor": 3600,
            "displayQuantity": 1,
            "tieredRates": [
              {
                "startUsageAmount": 0,
                "unitPrice": {
                  "currencyCode": "USD",
                  "units": "0",
                  "nanos": 44856000
                }
              }
            ]
          },
          "currencyConversionRate": 1,
          "effectiveTime": "2020-08-05T01:48:54.819Z"
        }
      ]
	}
	`
	str3 = `
	{
		"description": "Preemptible N2 Custom Instance Ram running in Sao Paulo",
		"category": {
			"serviceDisplayName": "Compute Engine",
			"resourceFamily": "Compute",
			"resourceGroup": "RAM",
			"usageType": "Preemptible"
		},
		"serviceRegions": [
			"southamerica-east1"
		],
		"pricingInfo": [
        {
          "summary": "",
          "pricingExpression": {
            "usageUnit": "GiBy.h",
            "usageUnitDescription": "gibibyte hour",
            "baseUnit": "By.s",
            "baseUnitDescription": "byte second",
            "baseUnitConversionFactor": 3865470566400,
            "displayQuantity": 1,
            "tieredRates": [
              {
                "startUsageAmount": 0,
                "unitPrice": {
                  "currencyCode": "USD",
                  "units": "0",
                  "nanos": 1121733
                }
              }
            ]
          },
          "currencyConversionRate": 1,
          "effectiveTime": "2020-08-05T01:48:54.819Z"
        }
      ]
	}
	`
	str4 = `
	{
		"description": "N1 Instance Ram running in Americas",
		"category": {
			"serviceDisplayName": "Compute Engine",
			"resourceFamily": "Compute",
			"resourceGroup": "N1Standard",
			"usageType": "OnDemand"
		},
		"serviceRegions": [
			"us-west1",
			"us-east1"
		],
		"pricingInfo": [
        {
          "summary": "",
          "pricingExpression": {
            "usageUnit": "GiBy.h",
            "usageUnitDescription": "gibibyte hour",
            "baseUnit": "By.s",
            "baseUnitDescription": "byte second",
            "baseUnitConversionFactor": 3865470566400,
            "displayQuantity": 1,
            "tieredRates": [
              {
                "startUsageAmount": 0,
                "unitPrice": {
                  "currencyCode": "USD",
                  "units": "0",
                  "nanos": 2701000
                }
              }
            ]
          },
          "currencyConversionRate": 1,
          "effectiveTime": "2020-08-05T01:48:54.819Z"
        }
      ]
	}
	`
	str5 = `
	{
		"name": "services/6F81-5844-456A/skus/0450-45CE-C078",
		"skuId": "0450-45CE-C078",
		"description": "N2D AMD Instance Core running in Netherlands",
		"category": {
		  "serviceDisplayName": "Compute Engine",
		  "resourceFamily": "Compute",
		  "resourceGroup": "CPU",
		  "usageType": "OnDemand"
		},
		"serviceRegions": [
		  "europe-west4"
		],
		"pricingInfo": [
		  {
			"summary": "",
			"pricingExpression": {
			  "usageUnit": "h",
			  "usageUnitDescription": "hour",
			  "baseUnit": "s",
			  "baseUnitDescription": "second",
			  "baseUnitConversionFactor": 3600,
			  "displayQuantity": 1,
			  "tieredRates": [
				{
				  "startUsageAmount": 0,
				  "unitPrice": {
					"currencyCode": "USD",
					"units": "0",
					"nanos": 30278000
				  }
				}
			  ]
			},
			"currencyConversionRate": 1,
			"effectiveTime": "2020-08-05T01:48:54.819Z"
		  }
		]
	}
	`
)

func mapToDescription(skus []*billingpb.Sku) (mapped []string) {
	for _, sku := range skus {
		mapped = append(mapped, sku.Description)
	}
	return
}

func TestCompletePricingInfo(t *testing.T) {
	sku1 := new(billingpb.Sku)
	sku2 := new(billingpb.Sku)
	sku3 := new(billingpb.Sku)
	sku4 := new(billingpb.Sku)
	sku5 := new(billingpb.Sku)

	jsonpb.UnmarshalString(str1, sku1)
	jsonpb.UnmarshalString(str2, sku2)
	jsonpb.UnmarshalString(str3, sku3)
	jsonpb.UnmarshalString(str4, sku4)
	jsonpb.UnmarshalString(str5, sku5)

	core1 := CoreInfo{ResourceGroup: "CPU", Number: 4, Fractional: 1, UnitPricing: PricingInfo{}}
	core2 := CoreInfo{ResourceGroup: "CPU", Number: 8, Fractional: 1, UnitPricing: PricingInfo{}}
	core3 := core1
	mem1 := MemoryInfo{ResourceGroup: "RAM", AmountGiB: 100, UnitPricing: PricingInfo{}}
	mem2 := MemoryInfo{ResourceGroup: "N1Standard", AmountGiB: 150, UnitPricing: PricingInfo{}}
	mem3 := mem1

	tests := []struct {
		skuObj  skuObject
		skus    []*billingpb.Sku
		pricing PricingInfo
		err     error
	}{
		{&core1, []*billingpb.Sku{sku1, sku3, sku4}, PricingInfo{UsageUnit: "hour", HourlyUnitPrice: 6980000, CurrencyType: "USD", CurrencyUnit: "nano"}, nil},
		{&core2, []*billingpb.Sku{sku2, sku3, sku4}, PricingInfo{UsageUnit: "hour", HourlyUnitPrice: 44856000, CurrencyType: "USD", CurrencyUnit: "nano"}, nil},
		{&mem1, []*billingpb.Sku{sku1, sku2, sku3, sku4, sku5}, PricingInfo{UsageUnit: "gibibyte hour", HourlyUnitPrice: 1121733, CurrencyType: "USD", CurrencyUnit: "nano"}, nil},
		{&mem2, []*billingpb.Sku{sku1, sku2, sku3, sku4, sku5}, PricingInfo{UsageUnit: "gibibyte hour", HourlyUnitPrice: 2701000, CurrencyType: "USD", CurrencyUnit: "nano"}, nil},
		{&core3, []*billingpb.Sku{sku3, sku4}, PricingInfo{}, fmt.Errorf("could not find core pricing information")},
		{&mem3, []*billingpb.Sku{sku1, sku2, sku5}, PricingInfo{}, fmt.Errorf("could not find memory pricing information")},
	}

	for _, test := range tests {
		err := test.skuObj.completePricingInfo(test.skus)
		fail1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
		fail2 := err != nil && test.err != nil && err.Error() != test.err.Error()
		fail3 := test.pricing != test.skuObj.getPricingInfo()

		if fail1 || fail2 || fail3 {
			t.Errorf("{%+v}.completePricingInfo(%+v) -> %+v, %+v; want %+v, %+v",
				test.skuObj, mapToDescription(test.skus), test.skuObj.getPricingInfo(), err, test.pricing, test.err)
		}
	}
}

func TestCoreGetTotalPrice(t *testing.T) {
	c1 := CoreInfo{Number: 2, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000}}
	c2 := CoreInfo{Number: 4, Fractional: 0.125, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000}}
	c3 := CoreInfo{Number: 32, Fractional: 0.5, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733}}
	c4 := CoreInfo{Number: 16, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000}}

	tests := []struct {
		core  CoreInfo
		price float64
	}{
		{c1, float64(6980000) * 2 / nano},
		{c2, float64(44856000) * 4 / nano * 0.125},
		{c3, float64(1121733) * 32 / nano * 0.5},
		{c4, float64(2701000) * 16 / nano},
	}

	for _, test := range tests {
		actual := test.core.getTotalPrice()
		if math.Abs(actual-test.price) > epsilon {
			t.Errorf("{%+v}.getTotalPrice() = %f ; want %f", test.core, actual, test.price)
		}
	}
}

func TestMemGetTotalPrice(t *testing.T) {
	m1 := MemoryInfo{AmountGiB: 100, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000, UsageUnit: "gigabyte hour"}}
	m2 := MemoryInfo{AmountGiB: 50, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000, UsageUnit: "pebibyte hour"}}
	m3 := MemoryInfo{AmountGiB: 320, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733, UsageUnit: "tebibyte hour"}}
	m4 := MemoryInfo{AmountGiB: 16, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "gibibyte hour"}}
	m5 := MemoryInfo{AmountGiB: 160, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "giBibyte hour"}}

	gb := float64(1000 * 1000 * 1000)
	gib := float64(1024 * 1024 * 1024)

	tests := []struct {
		mem   MemoryInfo
		price float64
		err   error
	}{
		{m1, float64(6980000) / nano * 100 * gib / gb, nil},
		{m2, float64(44856000) / nano * 50 / (1024 * 1024), nil},
		{m3, float64(1121733) / nano * 320 / 1024, nil},
		{m4, float64(2701000) / nano * 16, nil},
		{m5, 0, fmt.Errorf("unknown final unit giBibyte")},
	}

	for _, test := range tests {
		actual, err := test.mem.getTotalPrice()
		fail1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
		fail2 := err != nil && test.err != nil && err.Error() != test.err.Error()
		if math.Abs(actual-test.price) > epsilon || fail1 || fail2 {
			t.Errorf("{%+v}.getTotalPrice() = %f, %+v ; want %f, %+v", test.mem, actual, err, test.price, test.err)
		}
	}
}

func TestGetDelta(t *testing.T) {
	core1 := CoreInfo{Number: 4, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 12345}}
	mem1 := MemoryInfo{AmountGiB: 1000, UnitPricing: PricingInfo{HourlyUnitPrice: 23455, UsageUnit: "gibibyte hour"}}
	instance1 := ComputeInstance{Cores: core1, Memory: mem1}

	core2 := CoreInfo{Number: 16, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 12345}}
	mem2 := MemoryInfo{AmountGiB: 500, UnitPricing: PricingInfo{HourlyUnitPrice: 23455, UsageUnit: "gibibyte hour"}}
	instance2 := ComputeInstance{Cores: core2, Memory: mem2}

	core3 := CoreInfo{Number: 8, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 67458}}
	mem3 := MemoryInfo{AmountGiB: 100, UnitPricing: PricingInfo{HourlyUnitPrice: 78996, UsageUnit: "gibibyte hour"}}
	instance3 := ComputeInstance{Cores: core3, Memory: mem3}

	core4 := CoreInfo{Number: 32, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 785678}}
	mem4 := MemoryInfo{AmountGiB: 2000, UnitPricing: PricingInfo{HourlyUnitPrice: 235977, UsageUnit: "gibibyte hour"}}
	instance4 := ComputeInstance{Cores: core4, Memory: mem4}

	badCore := CoreInfo{Number: 32, Fractional: 1, UnitPricing: PricingInfo{HourlyUnitPrice: 785678}}
	badMem := MemoryInfo{AmountGiB: 2000, UnitPricing: PricingInfo{HourlyUnitPrice: 235977, UsageUnit: "gigbyte hour"}}
	badInstance := ComputeInstance{Name: "test", MachineType: "n1-standard-1", Cores: badCore, Memory: badMem}

	tests := []struct {
		state ComputeInstanceState
		dcore float64
		dmem  float64
		err   error
	}{
		{ComputeInstanceState{Before: &instance1, After: &instance2, Action: "test1"}, (16 - 4) * 12345, (500 - 1000) * 23455, nil},
		{ComputeInstanceState{Before: &instance2, After: &instance3, Action: "test2"}, 8*67458 - 16*12345, 100*78996 - 500*23455, nil},
		{ComputeInstanceState{Before: &instance3, After: &instance4, Action: "test3"}, 32*785678 - 8*67458, 2000*235977 - 100*78996, nil},
		{ComputeInstanceState{Before: &instance2, After: &instance2, Action: "test4"}, 0, 0, nil},
		{ComputeInstanceState{Before: &instance1, After: &badInstance, Action: "test6"}, 0, 0, fmt.Errorf("test(n1-standard-1): unknown final unit gigbyte")},
	}

	for _, test := range tests {
		dcore, dmem, err := test.state.getDelta()
		fail1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
		fail2 := err != nil && test.err != nil && err.Error() != test.err.Error()
		if math.Abs(dcore-test.dcore/nano) > epsilon || math.Abs(dmem-test.dmem/nano) > epsilon || fail1 || fail2 {
			t.Errorf("<%s state>.getDelta() = %f, %f, %s ; want %f, %f, %s",
				test.state.Action, dcore, dmem, err, test.dcore, test.dmem, test.err)
		}
	}
}
