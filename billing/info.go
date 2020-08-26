package billing

import (
	"context"
	"fmt"
	"strings"

	billing "cloud.google.com/go/billing/apiv1"
	"google.golang.org/api/iterator"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

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

func fitsCategory(sku *billingpb.Sku, serviceDisplayName, resourceFamily, usageType string) bool {
	c := sku.Category
	cond1 := c.ServiceDisplayName == serviceDisplayName && c.ResourceFamily == resourceFamily
	cond2 := c.UsageType == usageType
	return cond1 && cond2
}

func fitsRegion(sku *billingpb.Sku, region string) bool {
	if sku.ServiceRegions != nil {
		for _, r := range sku.ServiceRegions {
			if r == region {
				return true
			}
		}
	}
	return false
}

// GetPricingInfo returns the pricing information of an SKU.
func GetPricingInfo(sku *billingpb.Sku) (usageUnit string, hourlyUnitPrice int64, currencyType, currencyUnit string) {
	pExpr := sku.PricingInfo[0].PricingExpression
	usageUnit = pExpr.UsageUnitDescription
	unitPrice := pExpr.TieredRates[0].UnitPrice
	hourlyUnitPrice = int64(unitPrice.Nanos)
	currencyType = unitPrice.CurrencyCode
	currencyUnit = "nano"
	return
}

// GetSKUs returns the SKUs from the Compute Engine billing API or an error.
func GetSKUs(ctx context.Context) ([]*billingpb.Sku, error) {
	var skus []*billingpb.Sku

	c, err := billing.NewCloudCatalogClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &billingpb.ListSkusRequest{
		Parent: "services/6F81-5844-456A",
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
	if skus == nil || len(skus) == 0 {
		return nil, fmt.Errorf("SKU list must not be empty")
	}

	filtered := []*billingpb.Sku{}

	for _, sku := range skus {
		if fitsDescription(sku, contains, omits) {
			filtered = append(filtered, sku)
		}
	}

	if filtered == nil || len(filtered) == 0 {
		return nil, fmt.Errorf("no SKU with the specified description")
	}

	return filtered, nil
}

// CategoryFilter returns the SKUs with the specified category attributes.
func CategoryFilter(skus []*billingpb.Sku, serviceDisplayName,
	resourceFamily, usageType string) ([]*billingpb.Sku, error) {
	if skus == nil || len(skus) == 0 {
		return nil, fmt.Errorf("SKU list must not be empty")
	}

	filtered := []*billingpb.Sku{}

	for _, sku := range skus {
		if fitsCategory(sku, serviceDisplayName, resourceFamily, usageType) {
			filtered = append(filtered, sku)
		}
	}

	if filtered == nil || len(filtered) == 0 {
		return nil, fmt.Errorf("no SKU from the specified category")
	}

	return filtered, nil
}

// RegionFilter returns the SKUs from the specified region.
func RegionFilter(skus []*billingpb.Sku, region string) ([]*billingpb.Sku, error) {
	if skus == nil || len(skus) == 0 {
		return nil, fmt.Errorf("SKU list must not be empty")
	}

	filtered := []*billingpb.Sku{}

	for _, sku := range skus {
		if fitsRegion(sku, region) {
			filtered = append(filtered, sku)
		}
	}

	if filtered == nil || len(filtered) == 0 {
		return nil, fmt.Errorf("region '" + region + "' is invalid")
	}

	return filtered, nil
}
