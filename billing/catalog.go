package billing

import (
	"context"
	"fmt"
	"strings"

	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

// ComputeEngineCatalog holds the information from the billing catalog for Compute Engine SKUs.
type ComputeEngineCatalog struct {
	service       string
	coreInstances map[string][]*billingpb.Sku
	ramInstances  map[string][]*billingpb.Sku
}

// NewComputeEngineCatalog creates a catalog instance, calls the billing API and stores its response by category and usage type.
func NewComputeEngineCatalog(ctx context.Context) (*ComputeEngineCatalog, error) {
	c := new(ComputeEngineCatalog)
	c.service = "services/6F81-5844-456A"
	c.coreInstances = map[string][]*billingpb.Sku{}
	c.ramInstances = map[string][]*billingpb.Sku{}

	skus, err := GetSKUs(ctx, c.service)
	if err != nil {
		return nil, err
	}
	c.assignSKUCategories(skus)

	return c, nil
}

func emptyComputeEngineCatalog() *ComputeEngineCatalog {
	c := new(ComputeEngineCatalog)
	c.service = "services/6F81-5844-456A"
	c.coreInstances = map[string][]*billingpb.Sku{}
	c.ramInstances = map[string][]*billingpb.Sku{}
	return c
}

func (catalog *ComputeEngineCatalog) assignSKUCategories(skus []*billingpb.Sku) {
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
				if _, ok := catalog.ramInstances[c.UsageType]; !ok {
					catalog.ramInstances[c.UsageType] = nil
				}
				catalog.ramInstances[c.UsageType] = append(catalog.ramInstances[c.UsageType], sku)
			}
		}
	}
}

// GetCoreSKUs returns the Core Instance SKUs from the billing API.
func (catalog *ComputeEngineCatalog) GetCoreSKUs(usageType string) ([]*billingpb.Sku, error) {
	skus, ok := catalog.coreInstances[usageType]
	if !ok {
		return nil, fmt.Errorf("found no core SKU of this usage type")
	}
	return skus, nil
}

// GetRAMSKUs returns the Ram Instance SKUs from the billing API.
func (catalog *ComputeEngineCatalog) GetRAMSKUs(usageType string) ([]*billingpb.Sku, error) {
	skus, ok := catalog.ramInstances[usageType]
	if !ok {
		return nil, fmt.Errorf("found no RAM SKU of this usage type")
	}
	return skus, nil
}
