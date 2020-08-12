package resources

import (
	"context"
	"fmt"

	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

//Resource is the interface of a general resource (ComputeInstance,...).
type Resource interface {
	ExtractResource(jsonResourceInfo interface{})
	CompletePricingInfo(context.Context)
	PrintPricingInfo()
}

// skuObject is the interface for the general resource in the billing catalog.
type skuObject interface {
	isMatch(sku *billingpb.Sku, region string) bool
	completePricingInfo(ctx context.Context,
		getSKUs func(context.Context) ([]*billingpb.Sku, error), region string) error
	getPricingInfo() PricingInfo
}

func getSKU(ctx context.Context, obj skuObject, getSKUs func(context.Context) ([]*billingpb.Sku, error),
	region string) (*billingpb.Sku, error) {
	skus, err := getSKUs(ctx)

	if err != nil {
		return nil, err
	}

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
