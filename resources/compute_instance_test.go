package resources

import (
	"context"
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

	core1 = CoreInfo{"N2", "CPU", "Preemptible", 4, PricingInfo{}}
	core2 = CoreInfo{"E2", "CPU", "OnDemand", 8, PricingInfo{}}
	mem1  = MemoryInfo{"N2", "RAM", "Preemptible", 100, PricingInfo{}}
	mem2  = MemoryInfo{"N1", "N1Standard", "OnDemand", 150, PricingInfo{}}

	badCore = CoreInfo{"N2", "CPU", "OnDemand", 8, PricingInfo{}}
)

func fakeGetSKUs(context.Context) ([]*billingpb.Sku, error) {
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

	return []*billingpb.Sku{sku1, sku2, sku3, sku4, sku5}, nil
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
		region string
		ok     bool
	}{
		{&sku1, &core1, "us-east1", true},
		{&sku1, &core1, "us-central1", true},
		{&sku3, &core1, "southamerica-east1", false},
		{&sku2, &core1, "southamerica-east1", false},
		{&sku2, &core2, "southamerica-west1", true},
		{&sku2, &core2, "us-east1", false},
		{&sku4, &core2, "us-east1", false},
		{&sku3, &mem1, "southamerica-east1", true},
		{&sku3, &mem1, "southamerica-west1", false},
		{&sku4, &mem2, "us-west1", true},
	}

	for _, test := range tests {
		actual := test.skuObj.isMatch(test.sku, test.region)
		if actual != test.ok {
			t.Errorf("sku.Description = %s, {%+v}.isMath(sku, %s) = %t; want %t",
				test.sku.Description, test.skuObj, test.region, actual, test.ok)
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
		region  string
		pricing PricingInfo
		err     error
	}{
		{&core1, "us-west1", PricingInfo{"hour", 6980000, "USD", "nano"}, nil},
		{&core2, "southamerica-east1", PricingInfo{"hour", 44856000, "USD", "nano"}, nil},
		{&mem1, "southamerica-east1", PricingInfo{"gibibyte hour", 1121733, "USD", "nano"}, nil},
		{&mem2, "us-east1", PricingInfo{"gibibyte hour", 2701000, "USD", "nano"}, nil},
		{&badCore, "southamerica-east1", PricingInfo{}, fmt.Errorf("could not find SKU type")},
		{&badCore, "us-west1", PricingInfo{}, fmt.Errorf("could not find SKU type")},
		{&badCore, "us-east1", PricingInfo{}, fmt.Errorf("could not find SKU type")},
	}

	for _, test := range tests {
		err := test.skuObj.completePricingInfo(context.Background(), fakeGetSKUs, test.region)
		fail1 := (err == nil && test.err != nil) || (err != nil && test.err == nil)
		fail2 := err != nil && test.err != nil && err.Error() != test.err.Error()
		fail3 := test.pricing != test.skuObj.getPricingInfo()

		if fail1 || fail2 || fail3 {
			t.Errorf("{%+v}.completePricingInfo(context, fakeGetSKUs, %s) -> %+v, %+v; want %+v, %+v",
				test.skuObj, test.region, test.skuObj.getPricingInfo(), err, test.pricing, test.err)
		}
	}
}
