package billing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

var (
	str1 = []byte(`
	{
		"skuId": "0048-21CE-74C3",
		"serviceProviderName": "Google"
	}
	`)
	str2 = []byte(`{"description": "N2 Custom Instance Core running in Sao Paulo","category": {"serviceDisplayName": "Compute Engine","resourceFamily": "Compute","resourceGroup": "CPU","usageType": "OnDemand"},"serviceRegions": ["southamerica-east1"]}`)
	str3 = []byte(`{"description": "Preemptible N2 Custom Instance Ram running in Sao Paulo","category": {"serviceDisplayName": "Compute Engine","resourceFamily": "Compute","resourceGroup": "RAM","usageType": "Preemptible"},"serviceRegions": ["southamerica-east1"]}`)
	str4 = []byte(`{"description": "N1 Instance Ram running in Americas","category": {"serviceDisplayName": "Compute Engine","resourceFamily": "Compute","resourceGroup": "N1Standard","usageType": "OnDemand"},"serviceRegions": ["us-central1","us-west1", "us-east1"]}`)
)

func TestFitsDescription(t *testing.T) {
	var sku1, sku2, sku3, sku4 *billingpb.Sku
	json.Unmarshal(str1, &sku1)
	a, b := json.Marshal(sku1)
	fmt.Println(string(a), b)
	json.Unmarshal(str2, &sku2)
	json.Unmarshal(str3, &sku3)
	json.Unmarshal(str4, &sku4)

	tests := []struct {
		sku      *billingpb.Sku
		contains []string
		omits    []string
		ok       bool
	}{
		{sku1, []string{"Preemptible", "Instance Core"}, []string{"Custom"}, true},
		{sku2, []string{"Preemptible", "Instance Core"}, []string{"Custom"}, false},
		{sku3, []string{"Preemptible", "Instance Ram", "N2"}, []string{"Custom"}, false},
		{sku4, []string{"N1", "Instance Ram"}, []string{"Custom"}, true},
		{sku4, []string{"N1", "Instance Ram", "Custom"}, []string{}, false},
		{sku3, []string{"N2", "Instance Ram", "Custom"}, []string{}, true},
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

	content, _ := ioutil.ReadFile("skus.json")
	var sku billingpb.Sku
	json.Unmarshal([]byte(content), &sku)
	fmt.Println(sku)

	var sku1, sku2, sku3, sku4 billingpb.Sku
	json.Unmarshal(str1, &sku1)
	json.Unmarshal(str2, &sku2)
	json.Unmarshal(str3, &sku3)
	json.Unmarshal(str4, &sku4)

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
		{&sku4, "Compute Engine", "Compute Engine", "N1Standard", "OnDemand", true},
	}

	for _, test := range tests {
		actual := FitsCategory(test.sku, test.serviceDisplayName, test.resourceFamily, test.resourceGroup, test.usageType)
		if actual != test.ok {
			t.Errorf("sku.Description = %s, FitsCategory(sku, %s, %s, %s, %s) = %t; want %t", test.sku.Description,
				test.serviceDisplayName, test.resourceFamily, test.resourceGroup, test.usageType, actual, test.ok)
		}
	}
}
