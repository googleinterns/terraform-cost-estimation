package billing

import (
	"context"
	"fmt"
	"strings"

	billing "cloud.google.com/go/billing/apiv1"
	"google.golang.org/api/iterator"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

// FitsDescription checks if an SKU description fits the requirements.
func FitsDescription(sku *billingpb.Sku, contains, omits []string) bool {
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

// FitsCategory checks if an SKU has the requested category attributes.
func FitsCategory(sku *billingpb.Sku, serviceDisplayName, resourceFamily, resourceGroup, usageType string) bool {
	c := sku.Category
	cond1 := c.ServiceDisplayName == serviceDisplayName && c.ResourceFamily == resourceFamily
	if !cond1 {
		fmt.Println(c.ServiceDisplayName == serviceDisplayName, c.ServiceDisplayName, resourceFamily)
	}
	cond2 := c.ResourceGroup == resourceGroup && c.UsageType == usageType
	return cond1 && cond2
}

// FitsRegion checks if the SKU is available in a specific region.
func FitsRegion(sku *billingpb.Sku, region string) bool {
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
	pExpr := sku.PricingInfo[0].GetPricingExpression()
	usageUnit = pExpr.GetUsageUnitDescription()
	unitPrice := pExpr.GetTieredRates()[0].GetUnitPrice()
	hourlyUnitPrice = int64(unitPrice.GetNanos())
	currencyType = unitPrice.GetCurrencyCode()
	currencyUnit = "nano"
	return
}

// GetSKUs returns the SKUs from the Compute Engine billing API or an error.
func GetSKUs() ([]*billingpb.Sku, error) {
	var skus []*billingpb.Sku

	ctx := context.Background()
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
