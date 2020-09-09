package billing

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

var (
	skuStr1 = `
		{
			"name": "services/6F81-5844-456A/skus/000F-E31B-1D6F",
			"skuId": "000F-E31B-1D6F",
			"description": "N1 Predefined Instance Ram running in Zurich",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "Compute",
				"resourceGroup": "N1Standard",
				"usageType": "OnDemand"
			}
		}`
	skuStr2 = `
		{
			"name": "services/6F81-5844-456A/skus/0012-B7F2-DD14",
			"skuId": "0012-B7F2-DD14",
			"description": "Preemptible Compute optimized Ram running in Montreal",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "Compute",
				"resourceGroup": "RAM",
				"usageType": "Preemptible"
			}
		}`
	skuStr3 = `
		{
			"name": "services/6F81-5844-456A/skus/0013-863C-A2FF",
			"skuId": "0013-863C-A2FF",
			"description": "Licensing Fee for SQL Server 2016 Standard on VM with 18 VCPU",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "License",
				"resourceGroup": "SQLServer2016Standard",
				"usageType": "OnDemand"
			}
		}`
	skuStr4 = `
		{
			"name": "services/6F81-5844-456A/skus/0014-939F-88A0",
			"skuId": "0014-939F-88A0",
			"description": "Licensing Fee for Windows Server 2012 BYOL (CPU cost)",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "License",
				"resourceGroup": "Google",
				"usageType": "OnDemand"
			}
		}`
	skuStr5 = `
		{
			"name": "services/6F81-5844-456A/skus/001D-204A-23DA",
			"skuId": "001D-204A-23DA",
			"description": "Commitment v1: Cpu in Montreal for 1 Year",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "Compute",
				"resourceGroup": "CPU",
				"usageType": "Commit1Yr"
			}
		}`
	skuStr6 = `
		{
			"name": "services/6F81-5844-456A/skus/0026-A923-AA09",
			"skuId": "0026-A923-AA09",
			"description": "Sole Tenancy Instance Ram running in Jakarta",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "Compute",
				"resourceGroup": "RAM",
				"usageType": "OnDemand"
			}
		}`
	skuStr7 = `
		{
			"name": "services/6F81-5844-456A/skus/002C-E0AF-213B",
			"skuId": "002C-E0AF-213B",
			"description": "Licensing Fee for SQL Server 2012 SP3 Express Edition on Windows Server 2012 R2 on VM with 12 VCPU",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "License",
				"resourceGroup": "Cognosys",
				"usageType": "OnDemand"
			}
		}`
	skuStr8 = `
		{
			"name": "services/6F81-5844-456A/skus/0031-0209-F39B",
			"skuId": "0031-0209-F39B",
			"description": "Network Vpn Inter Region Ingress from Hong Kong to Sydney",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "Network",
				"resourceGroup": "VPNInterregionIngress",
				"usageType": "OnDemand"
			}
		}`
	skuStr9 = `
		{
			"name": "services/6F81-5844-456A/skus/0048-21CE-74C3",
			"skuId": "0048-21CE-74C3",
			"description": "Preemptible N2 Custom Instance Core running in Sao Paulo",
			"category": {
				"serviceDisplayName": "Compute Engine",
				"resourceFamily": "Compute",
				"resourceGroup": "CPU",
				"usageType": "Preemptible"
			}
		}`
	skuStr10 = `
		{
			"name": "services/6F81-5844-456A/skus/0143-7EA7-329F",
			"skuId": "0143-7EA7-329F",
			"description": "N1 Predefined Instance Ram running in Singapore",
			"category": {
			  "serviceDisplayName": "Compute Engine",
			  "resourceFamily": "Compute",
			  "resourceGroup": "N1Standard",
			  "usageType": "OnDemand"
			}
		}`
	skuStr11 = `
		{
			"name": "services/6F81-5844-456A/skus/023F-CB27-DC68",
			"skuId": "023F-CB27-DC68",
			"description": "Preemptible N1 Predefined Instance Core running in Virginia",
			"category": {
			  "serviceDisplayName": "Compute Engine",
			  "resourceFamily": "Compute",
			  "resourceGroup": "N1Standard",
			  "usageType": "Preemptible"
			}
		}`
)

func show(c *ComputeEngineCatalog) (s string) {
	s = "Cores:"
	for k, v := range c.coreInstances {
		s += "\n" + k + ": "
		for _, sku := range v {
			s += sku.Description + "; "
		}
	}

	s += "RAM:"
	for k, v := range c.ramInstances {
		s += "\n" + k + ": "
		for _, sku := range v {
			s += sku.Description + "; "
		}
	}
	return
}

func TestAssignSKUCategories(t *testing.T) {
	sku1 := new(billingpb.Sku)
	sku2 := new(billingpb.Sku)
	sku3 := new(billingpb.Sku)
	sku4 := new(billingpb.Sku)
	sku5 := new(billingpb.Sku)
	sku6 := new(billingpb.Sku)
	sku7 := new(billingpb.Sku)
	sku8 := new(billingpb.Sku)
	sku9 := new(billingpb.Sku)
	sku10 := new(billingpb.Sku)
	sku11 := new(billingpb.Sku)

	jsonpb.UnmarshalString(skuStr1, sku1)
	jsonpb.UnmarshalString(skuStr2, sku2)
	jsonpb.UnmarshalString(skuStr3, sku3)
	jsonpb.UnmarshalString(skuStr4, sku4)
	jsonpb.UnmarshalString(skuStr5, sku5)
	jsonpb.UnmarshalString(skuStr6, sku6)
	jsonpb.UnmarshalString(skuStr7, sku7)
	jsonpb.UnmarshalString(skuStr8, sku8)
	jsonpb.UnmarshalString(skuStr9, sku9)
	jsonpb.UnmarshalString(skuStr10, sku10)
	jsonpb.UnmarshalString(skuStr11, sku11)

	c1 := emptyComputeEngineCatalog()
	c2 := emptyComputeEngineCatalog()
	c3 := emptyComputeEngineCatalog()
	c4 := emptyComputeEngineCatalog()

	c2.coreInstances["Preemptible"] = []*billingpb.Sku{sku11}
	c2.ramInstances["OnDemand"] = []*billingpb.Sku{sku1, sku10}

	c3.coreInstances["Commit1Yr"] = []*billingpb.Sku{sku5}
	c3.coreInstances["Preemptible"] = []*billingpb.Sku{sku9}
	c3.ramInstances["Preemptible"] = []*billingpb.Sku{sku2}
	c3.ramInstances["OnDemand"] = []*billingpb.Sku{sku6}

	c4.coreInstances["Commit1Yr"] = []*billingpb.Sku{sku5}
	c4.coreInstances["Preemptible"] = []*billingpb.Sku{sku9, sku11}
	c4.ramInstances["OnDemand"] = []*billingpb.Sku{sku1, sku6, sku10}
	c4.ramInstances["Preemptible"] = []*billingpb.Sku{sku2}

	tests := []struct {
		skus    []*billingpb.Sku
		catalog *ComputeEngineCatalog
	}{
		{[]*billingpb.Sku{sku3, sku4, sku7, sku8}, c1},
		{[]*billingpb.Sku{sku1, sku10, sku11}, c2},
		{[]*billingpb.Sku{sku2, sku5, sku6, sku9}, c3},
		{[]*billingpb.Sku{sku1, sku2, sku3, sku4, sku5, sku6, sku7, sku8, sku9, sku10, sku11}, c4},
	}

	for i, test := range tests {
		catalog := emptyComputeEngineCatalog()
		catalog.assignSKUCategories(test.skus)

		if !reflect.DeepEqual(*catalog, *test.catalog) {
			t.Errorf("test%+v: catalog.assignSKUCategories(skus) -> \n%+v,;\n want\n %+v", i, show(catalog), show(test.catalog))
		}
	}
}
