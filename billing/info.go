package billing

import (
	"context"
	"fmt"
	"strings"

	billing "cloud.google.com/go/billing/apiv1"
	"google.golang.org/api/iterator"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

const nano = float64(1000 * 1000 * 1000)

func fitsDescription(sku *billingpb.Sku, contains, omits []string) bool {
	if contains != nil {
		for _, d := range contains {
			if !strings.Contains(sku.Description, d) {
				return false
			}
		}
	}

	if omits != nil {
		for _, d := range omits {
			if strings.Contains(sku.Description, d) {
				return false
			}
		}
	}

	return true
}

func fitsRegion(sku *billingpb.Sku, region string) bool {
	if len(sku.ServiceRegions) == 0 {
		return false
	}

	if sku.ServiceRegions[0] == "global" {
		return true
	}

	for _, r := range sku.ServiceRegions {
		if r == region {
			return true
		}
	}
	return false
}

// PricingInfo returns the pricing information of an SKU.
func PricingInfo(sku *billingpb.Sku, correctTieredRate func(*billingpb.PricingExpression_TierRate) bool) (usageUnit string,
	pricePerUnit float64, currencyType string) {

	pExpr := sku.PricingInfo[0].PricingExpression
	usageUnit = strings.Split(pExpr.UsageUnitDescription, " ")[0]

	var tr *billingpb.PricingExpression_TierRate
	for i := len(pExpr.TieredRates) - 1; i >= 0; i-- {
		if correctTieredRate(pExpr.TieredRates[i]) {
			tr = pExpr.TieredRates[i]
			break
		}
	}
	if tr == nil {
		return
	}

	pricePerUnit = float64(tr.UnitPrice.Nanos) / nano
	currencyType = tr.UnitPrice.CurrencyCode
	return
}

// GetSKUs returns the SKUs from the billing API for the specific service or an error.
func GetSKUs(ctx context.Context, service string) ([]*billingpb.Sku, error) {
	var skus []*billingpb.Sku

	c, err := billing.NewCloudCatalogClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &billingpb.ListSkusRequest{
		Parent: service,
	}

	it := c.ListSkus(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		skus = append(skus, resp)
	}
	return skus, nil
}

// DescriptionFilter returns the SKUs that meet the description requirements.
func DescriptionFilter(skus []*billingpb.Sku, contains, omits []string) ([]*billingpb.Sku, error) {
	if len(skus) == 0 {
		return nil, fmt.Errorf("SKU list must not be empty")
	}

	filtered := []*billingpb.Sku{}

	for _, sku := range skus {
		if fitsDescription(sku, contains, omits) {
			filtered = append(filtered, sku)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no SKU with the specified description")
	}

	return filtered, nil
}

// RegionFilter returns the SKUs from the specified region.
func RegionFilter(skus []*billingpb.Sku, region string) ([]*billingpb.Sku, error) {
	if len(skus) == 0 {
		return nil, fmt.Errorf("SKU list must not be empty")
	}

	filtered := []*billingpb.Sku{}

	for _, sku := range skus {
		if fitsRegion(sku, region) {
			filtered = append(filtered, sku)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("region '" + region + "' is invalid")
	}

	return filtered, nil
}
