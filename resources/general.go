package resources

import (
	"context"
	"fmt"

	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

//Resource is the interface of a general resource (ComputeInstance,...).
type Resource interface {
	ExtractResource(jsonResourceInfo interface{})
	CompletePricingInfo(context.Context) error
	PrintPricingInfo()
}

// skuObject is the interface for SKU types (core, memory etc.)
// that can be looked up in the billing catalog.
type skuObject interface {
	isMatch(sku *billingpb.Sku, region string) bool
	completePricingInfo(skus []*billingpb.Sku, region string) error
	getPricingInfo() PricingInfo
}

func getSKU(skus []*billingpb.Sku, obj skuObject, region string) (*billingpb.Sku, error) {
	if skus == nil || len(skus) == 0 {
		return nil, fmt.Errorf("could not find SKUs")
	}

	for _, sku := range skus {
		if obj.isMatch(sku, region) {
			return sku, nil
		}
	}
	return nil, fmt.Errorf("could not find SKU type")
}
