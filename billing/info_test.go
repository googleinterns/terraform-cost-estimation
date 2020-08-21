package billing

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
		"description": "N2 Custom Instance Core running in Sao Paulo",
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
)

/*
func TestFitsDescription(t *testing.T) {
	var sku1, sku2, sku3, sku4 billingpb.Sku
	jsonpb.UnmarshalString(str1, &sku1)
	jsonpb.UnmarshalString(str2, &sku2)
	jsonpb.UnmarshalString(str3, &sku3)
	jsonpb.UnmarshalString(str4, &sku4)

	tests := []struct {
		sku      *billingpb.Sku
		contains []string
		omits    []string
		ok       bool
	}{
		{&sku1, []string{"Preemptible", "Instance Core"}, []string{"Custom"}, true},
		{&sku2, []string{"Preemptible", "Instance Core"}, []string{"Custom"}, false},
		{&sku3, []string{"Preemptible", "Instance Ram", "N2"}, []string{"Custom"}, false},
		{&sku4, []string{"N1", "Instance Ram"}, []string{"Custom"}, true},
		{&sku4, []string{"N1", "Instance Ram", "Custom"}, []string{}, false},
		{&sku3, []string{"N2", "Instance Ram", "Custom"}, []string{}, true},
	}

	for _, test := range tests {
		actual := FitsDescription(test.sku, test.contains, test.omits)
		if actual != test.ok {
			t.Errorf("sku.Description = %s, FitsDescription(sku, %+v, %+v) = %t; want %t",
				test.sku.Description, test.contains, test.omits, actual, test.ok)
		}
	}
}

func TestFitsCategory(t *testing.T) {
	var sku1, sku2, sku3, sku4 billingpb.Sku
	jsonpb.UnmarshalString(str1, &sku1)
	jsonpb.UnmarshalString(str2, &sku2)
	jsonpb.UnmarshalString(str3, &sku3)
	jsonpb.UnmarshalString(str4, &sku4)

	tests := []struct {
		sku                *billingpb.Sku
		serviceDisplayName string
		resourceFamily     string
		resourceGroup      string
		usageType          string
		ok                 bool
	}{
		{&sku1, "Compute Engine", "Compute", "CPU", "Preemptible", true},
		{&sku2, "VPN", "Compute", "CPU", "OnDemand", false},
		{&sku3, "Compute Engine", "Compute Engine", "RAM", "Preemptible", false},
		{&sku4, "Compute Engine", "Compute Engine", "CPU", "OnDemand", false},
		{&sku4, "Compute Engine", "Compute Engine", "N1Standard", "Preemptible", false},
		{&sku4, "Compute Engine", "Compute", "N1Standard", "OnDemand", true},
	}

	for _, test := range tests {
		actual := FitsCategory(test.sku, test.serviceDisplayName, test.resourceFamily, test.resourceGroup, test.usageType)
		if actual != test.ok {
			t.Errorf("sku.Description = %s, FitsCategory(sku, %s, %s, %s, %s) = %t; want %t", test.sku.Description,
				test.serviceDisplayName, test.resourceFamily, test.resourceGroup, test.usageType, actual, test.ok)
		}
	}
}

func TestFitsRegion(t *testing.T) {
	var sku1, sku2, sku3, sku4 billingpb.Sku
	jsonpb.UnmarshalString(str1, &sku1)
	jsonpb.UnmarshalString(str2, &sku2)
	jsonpb.UnmarshalString(str3, &sku3)
	jsonpb.UnmarshalString(str4, &sku4)

	tests := []struct {
		sku    *billingpb.Sku
		region string
		ok     bool
	}{
		{&sku1, "us-west1", true},
		{&sku1, "southamerica-east1", false},
		{&sku2, "southamerica-west1", true},
		{&sku2, "southamerica-west2", false},
		{&sku3, "southamerica-east1", true},
		{&sku3, "us-west1", false},
		{&sku4, "us-west1", true},
		{&sku4, "us-central1", false},
	}

	for _, test := range tests {
		actual := FitsRegion(test.sku, test.region)
		if actual != test.ok {
			t.Errorf("sku.Description = %s, FitsRegion(sku, %s) = %t; want %t",
				test.sku.Description, test.region, actual, test.ok)
		}
	}
}

func TestGetPricingInfo(t *testing.T) {
	var sku1, sku2, sku3, sku4 billingpb.Sku
	jsonpb.UnmarshalString(str1, &sku1)
	jsonpb.UnmarshalString(str2, &sku2)
	jsonpb.UnmarshalString(str3, &sku3)
	jsonpb.UnmarshalString(str4, &sku4)

	tests := []struct {
		sku             *billingpb.Sku
		usageUnit       string
		hourlyUnitPrice int64
		currencyType    string
		currencyUnit    string
	}{
		{&sku1, "hour", 6980000, "USD", "nano"},
		{&sku2, "hour", 44856000, "USD", "nano"},
		{&sku3, "gibibyte hour", 1121733, "USD", "nano"},
		{&sku4, "gibibyte hour", 2701000, "USD", "nano"},
	}

	for _, test := range tests {
		usageUnit, hourlyUnitPrice, currencyType, currencyUnit := GetPricingInfo(test.sku)
		fail1 := usageUnit != test.usageUnit || hourlyUnitPrice != test.hourlyUnitPrice
		fail2 := currencyType != test.currencyType || currencyUnit != test.currencyUnit
		if fail1 || fail2 {
			t.Errorf("sku.Description = %s, GetPricingInfo(sku) = %+v, %+v, %+v, %+v; want %+v, %+v, %+v, %+v",
				test.sku.Description, usageUnit, hourlyUnitPrice, currencyType, currencyUnit,
				test.usageUnit, test.hourlyUnitPrice, test.currencyType, test.currencyUnit)
		}
	}
}
*/
