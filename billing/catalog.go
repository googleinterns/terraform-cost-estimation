package billing

import (
	"context"
	"strings"

	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

type computeEngineCatalog struct {
	service       string
	coreInstances map[string][]*billingpb.Sku
	RAMInstances  map[string][]*billingpb.Sku
}

func newComputeEngineCatalog() *computeEngineCatalog {
	c := new(computeEngineCatalog)
	c.service = "services/6F81-5844-456A"
	c.coreInstances = map[string][]*billingpb.Sku{}
	c.RAMInstances = map[string][]*billingpb.Sku{}
	return c
}

var computeEngineCatalogPtr *computeEngineCatalog

func (catalog *computeEngineCatalog) assignSKUCategories(skus []*billingpb.Sku) {
	for _, sku := range skus {
		c := sku.Category
		if c.ServiceDisplayName == "Compute Engine" && c.ResourceFamily == "Compute" {
			if c.ResourceGroup == "CPU" || (c.ResourceGroup == "N1Standard" && !strings.Contains(sku.Description, "Ram")) {
				if _, ok := catalog.coreInstances[c.UsageType]; !ok {
					catalog.coreInstances[c.UsageType] = nil
				}
				catalog.coreInstances[c.UsageType] = append(catalog.coreInstances[c.UsageType], sku)
			}

			if c.ResourceGroup == "RAM" || (c.ResourceGroup == "N1Standard" && !strings.Contains(sku.Description, "Core")) {
				if _, ok := catalog.RAMInstances[c.UsageType]; !ok {
					catalog.RAMInstances[c.UsageType] = nil
				}
				catalog.RAMInstances[c.UsageType] = append(catalog.RAMInstances[c.UsageType], sku)
			}
		}
	}
}

// GetCoreSKUs returns the Core Instance SKUs from the billing API.
func GetCoreSKUs(ctx context.Context, usageType string) ([]*billingpb.Sku, error) {
	if computeEngineCatalogPtr == nil {
		computeEngineCatalogPtr = newComputeEngineCatalog()
	}

	if computeEngineCatalogPtr.coreInstances == nil || len(computeEngineCatalogPtr.coreInstances) == 0 {
		skus, err := GetSKUs(ctx, computeEngineCatalogPtr.service)
		if err != nil {
			return nil, err
		}
		computeEngineCatalogPtr.assignSKUCategories(skus)
	}
	return computeEngineCatalogPtr.coreInstances[usageType], nil
}

// GetRAMSKUs returns the Ram Instance SKUs from the billing API.
func GetRAMSKUs(ctx context.Context, usageType string) ([]*billingpb.Sku, error) {
	if computeEngineCatalogPtr == nil {
		computeEngineCatalogPtr = newComputeEngineCatalog()
	}

	if computeEngineCatalogPtr.RAMInstances == nil || len(computeEngineCatalogPtr.RAMInstances) == 0 {
		skus, err := GetSKUs(ctx, computeEngineCatalogPtr.service)
		if err != nil {
			return nil, err
		}
		computeEngineCatalogPtr.assignSKUCategories(skus)
	}
	return computeEngineCatalogPtr.RAMInstances[usageType], nil
}
