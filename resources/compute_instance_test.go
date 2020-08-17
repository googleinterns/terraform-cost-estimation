package resources

import (
	"fmt"
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

	core1 = CoreInfo{"N2", Description{[]string{"Preemptible"}, []string{"Custom"}}, "CPU", "Preemptible", 4, PricingInfo{}}
	core2 = CoreInfo{"E2", Description{[]string{"Custom"}, []string{"Preemptible"}}, "CPU", "OnDemand", 8, PricingInfo{}}
	mem1  = MemoryInfo{"N2", Description{[]string{"Preemptible", "Custom"}, []string{}}, "RAM", "Preemptible", 100, PricingInfo{}}
	mem2  = MemoryInfo{"N1", Description{[]string{}, []string{"Preemptible", "Custom"}}, "N1Standard", "OnDemand", 150, PricingInfo{}}

	badCore = CoreInfo{"N2", Description{[]string{}, []string{"Preemptible", "Custom", "Predefiend"}}, "CPU", "OnDemand", 8, PricingInfo{}}
)

func fakeGetSKUs() []*billingpb.Sku {
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

	return []*billingpb.Sku{sku1, sku2, sku3, sku4, sku5}
}

func TestIsMatch(t *testing.T) {
	var sku1, sku2, sku3, sku4 billingpb.Sku
	jsonpb.UnmarshalString(str1, &sku1)
	jsonpb.UnmarshalString(str2, &sku2)
	jsonpb.UnmarshalString(str3, &sku3)
	jsonpb.UnmarshalString(str4, &sku4)

	tests := []struct {
		sku    *billingpb.Sku
		skuObj skuObject
		ok     bool
	}{
		{&sku1, &core1, true},
		{&sku3, &core1, false},
		{&sku2, &core1, false},
		{&sku2, &core2, true},
		{&sku4, &core2, false},
		{&sku3, &mem1, true},
		{&sku4, &mem2, true},
	}

	for _, test := range tests {
		actual := test.skuObj.isMatch(test.sku)
		if actual != test.ok {
			t.Errorf("sku.Description = %s, {%+v}.isMath(sku) = %t; want %t",
				test.sku.Description, test.skuObj, actual, test.ok)
		}
	}
}

func TestCompletePricingInfo(t *testing.T) {
	var sku1, sku2, sku3, sku4 billingpb.Sku
	jsonpb.UnmarshalString(str1, &sku1)
	jsonpb.UnmarshalString(str2, &sku2)
	jsonpb.UnmarshalString(str3, &sku3)
	jsonpb.UnmarshalString(str4, &sku4)

	tests := []struct {
		skuObj  skuObject
		pricing PricingInfo
		err     error
	}{
		{&core1, PricingInfo{"hour", 6980000, "USD", "nano"}, nil},
		{&core2, PricingInfo{"hour", 44856000, "USD", "nano"}, nil},
		{&mem1, PricingInfo{"gibibyte hour", 1121733, "USD", "nano"}, nil},
		{&mem2, PricingInfo{"gibibyte hour", 2701000, "USD", "nano"}, nil},
		{&badCore, PricingInfo{}, fmt.Errorf("could not find core pricing information")},
		{&badCore, PricingInfo{}, fmt.Errorf("could not find core pricing information")},
		{&badCore, PricingInfo{}, fmt.Errorf("could not find core pricing information")},
	}

	for _, test := range tests {
		err := test.skuObj.completePricingInfo(fakeGetSKUs())
		fail1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
		fail2 := err != nil && test.err != nil && err.Error() != test.err.Error()
		fail3 := test.pricing != test.skuObj.getPricingInfo()

		if fail1 || fail2 || fail3 {
			t.Errorf("{%+v}.completePricingInfo(skus) -> %+v, %+v; want %+v, %+v",
				test.skuObj, test.skuObj.getPricingInfo(), err, test.pricing, test.err)
		}
	}
}

func TestCoreGetTotalPrice(t *testing.T) {
	c1 := CoreInfo{Number: 2, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000}}
	c2 := CoreInfo{Number: 4, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000}}
	c3 := CoreInfo{Number: 32, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733}}
	c4 := CoreInfo{Number: 16, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000}}

	nano := float64(1000 * 1000 * 1000)

	tests := []struct {
		core  CoreInfo
		price float64
	}{
		{c1, float64(6980000) * 2 / nano},
		{c2, float64(44856000) * 4 / nano},
		{c3, float64(1121733) * 32 / nano},
		{c4, float64(2701000) * 16 / nano},
	}

	for _, test := range tests {
		actual := test.core.getTotalPrice()
		if actual != test.price {
			t.Errorf("{%+v}.getTotalPrice() = %f ; want %f", test.core, actual, test.price)
		}
	}
}

func TestMemGetTotalPrice(t *testing.T) {
	m1 := MemoryInfo{AmountGB: 100, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000, UsageUnit: "gigabyte hour"}}
	m2 := MemoryInfo{AmountGB: 50, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000, UsageUnit: "pebibyte hour"}}
	m3 := MemoryInfo{AmountGB: 320, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733, UsageUnit: "tebibyte hour"}}
	m4 := MemoryInfo{AmountGB: 16, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "gibibyte hour"}}
	m5 := MemoryInfo{AmountGB: 160, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "giBibyte hour"}}

	gb := float64(1000 * 1000 * 1000)
	gib := float64(1024 * 1024 * 1024)
	tib := gib * float64(1024)
	pib := tib * float64(1024)
	nano := gb

	tests := []struct {
		mem   MemoryInfo
		price float64
		err   error
	}{
		{m1, float64(6980000) / nano * 100, nil},
		{m2, float64(44856000) / nano * 50 * gb / pib, nil},
		{m3, float64(1121733) / nano * 320 * gb / tib, nil},
		{m4, float64(2701000) / nano * 16 * gb / gib, nil},
		{m5, 0, fmt.Errorf("unknown final unit giBibyte")},
	}

	for _, test := range tests {
		actual, err := test.mem.getTotalPrice()
		fail1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
		fail2 := err != nil && test.err != nil && err.Error() != test.err.Error()
		if actual != test.price || fail1 || fail2 {
			t.Errorf("{%+v}.getTotalPrice() = %f, %+v ; want %f, %+v", test.mem, actual, err, test.price, test.err)
		}
	}
}

func TestCoreGetTotalPrice(t *testing.T) {
	c1 := CoreInfo{Number: 2, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000}}
	c2 := CoreInfo{Number: 4, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000}}
	c3 := CoreInfo{Number: 32, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733}}
	c4 := CoreInfo{Number: 16, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000}}

	nano := float64(1000 * 1000 * 1000)

	tests := []struct {
		core  CoreInfo
		price float64
	}{
		{c1, float64(6980000) * 2 / nano},
		{c2, float64(44856000) * 4 / nano},
		{c3, float64(1121733) * 32 / nano},
		{c4, float64(2701000) * 16 / nano},
	}

	for _, test := range tests {
		actual := test.core.getTotalPrice()
		if actual != test.price {
			t.Errorf("{%+v}.getTotalPrice() = %f ; want %f", test.core, actual, test.price)
		}
	}
}

func TestMemGetTotalPrice(t *testing.T) {
	m1 := MemoryInfo{AmountGB: 100, UnitPricing: PricingInfo{HourlyUnitPrice: 6980000, UsageUnit: "gigabyte hour"}}
	m2 := MemoryInfo{AmountGB: 50, UnitPricing: PricingInfo{HourlyUnitPrice: 44856000, UsageUnit: "pebibyte hour"}}
	m3 := MemoryInfo{AmountGB: 320, UnitPricing: PricingInfo{HourlyUnitPrice: 1121733, UsageUnit: "tebibyte hour"}}
	m4 := MemoryInfo{AmountGB: 16, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "gibibyte hour"}}
	m5 := MemoryInfo{AmountGB: 160, UnitPricing: PricingInfo{HourlyUnitPrice: 2701000, UsageUnit: "giBibyte hour"}}

	gb := float64(1000 * 1000 * 1000)
	gib := float64(1024 * 1024 * 1024)
	tib := gib * float64(1024)
	pib := tib * float64(1024)
	nano := gb

	tests := []struct {
		mem   MemoryInfo
		price float64
		err   error
	}{
		{m1, float64(6980000) / nano * 100, nil},
		{m2, float64(44856000) / nano * 50 * gb / pib, nil},
		{m3, float64(1121733) / nano * 320 * gb / tib, nil},
		{m4, float64(2701000) / nano * 16 * gb / gib, nil},
		{m5, 0, fmt.Errorf("unknown final unit giBibyte")},
	}

	for _, test := range tests {
		actual, err := test.mem.getTotalPrice()
		fail1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
		fail2 := err != nil && test.err != nil && err.Error() != test.err.Error()
		if actual != test.price || fail1 || fail2 {
			t.Errorf("{%+v}.getTotalPrice() = %f, %+v ; want %f, %+v", test.mem, actual, err, test.price, test.err)
		}
	}
}
