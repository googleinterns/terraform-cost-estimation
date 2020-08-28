package billing

import (
	"context"
	"strings"

	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

type computeEngineCatalog struct {
	service       string
	coreInstances []*billingpb.Sku
	RAMInstances  []*billingpb.Sku
}

func newComputeEngineCatalog() *computeEngineCatalog {
	c := new(computeEngineCatalog)
	c.service = "services/6F81-5844-456A"
	return c
}

var computeEngineCatalogPtr *computeEngineCatalog

func (catalog *computeEngineCatalog) assignSKUCategories(skus []*billingpb.Sku) {
	for _, sku := range skus {
		c := sku.Category
		if c.ServiceDisplayName == "Compute Engine" && c.ResourceFamily == "Compute" {
			if c.ResourceGroup == "CPU" || (c.ResourceGroup == "N1Standard" && !strings.Contains(sku.Description, "Ram")) {
				catalog.coreInstances = append(catalog.coreInstances, sku)
			}

			if c.ResourceGroup == "RAM" || (c.ResourceGroup == "N1Standard" && !strings.Contains(sku.Description, "Core")) {
				catalog.RAMInstances = append(catalog.RAMInstances, sku)
			}
		}
	}
}

// GetCoreSKUs returns the Core Instance SKUs from the billing API.
func GetCoreSKUs(ctx context.Context) ([]*billingpb.Sku, error) {
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
	return computeEngineCatalogPtr.coreInstances, nil
}

// GetRAMSKUs returns the Ram Instance SKUs from the billing API.
func GetRAMSKUs(ctx context.Context) ([]*billingpb.Sku, error) {
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
	return computeEngineCatalogPtr.RAMInstances, nil
}
