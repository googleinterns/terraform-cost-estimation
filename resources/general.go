package resources

import (
	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

const (
	nano    = float64(1000 * 1000 * 1000)
	epsilon = 1e-10
)

//ResourceState is the interface of a general before/after resource state(ComputeInstance,...).
type ResourceState interface {
	CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error
	PrintPricingInfo()
	GetSummary() string
}

// skuObject is the interface for SKU types (core, memory etc.)
// that can be looked up in the billing catalog.
type skuObject interface {
	isMatch(sku *billingpb.Sku) bool
	completePricingInfo(skus []*billingpb.Sku) error
	getPricingInfo() PricingInfo
}

func getSKU(skuObj skuObject, skus []*billingpb.Sku) *billingpb.Sku {
	for _, sku := range skus {
		if skuObj.isMatch(sku) {
			return sku
		}
	}
	return nil
}
