package catalog

import "strings"

type category struct {
	serviceDisplayName string
	resourceFamily     string
	resourceGroup      string
	usageType          string
}

type unitPrice struct {
	currencyCode string
	nanos        int64
}

type tieredRate struct {
	unitPrice unitPrice
}

type pricingExpression struct {
	usageUnitDescription string
	tieredRates          []tieredRate
}

type pricingInfo struct {
	pricingExpression pricingExpression
}

// ComputeEngineSKU holds the relevant fields from an SKU element from the list
// inside a billing catalog page for Compute Engine.
type ComputeEngineSKU struct {
	description    string
	category       category
	serviceRegions []string
	pricingInfo    pricingInfo
}

// ComputeEnginePage holds the fields from a billing
// catalog page regarding the Compute Engine SKUs.
type ComputeEnginePage struct {
	skus          []ComputeEngineSKU
	nextPageToken string
}

// MakeSKU builds a new Compute Engine SKU from scratch for testing purposes (1 tiered rate).
func MakeSKU(description, serviceDisplayName, resourceFamily, resourceGroup, usageType string, serviceRegions []string,
	usageUnitDescription string, currencyCode string, nanos int64) ComputeEngineSKU {

	category := category{serviceDisplayName, resourceFamily, resourceGroup, usageType}
	tieredRates := []tieredRate{tieredRate{unitPrice{currencyCode, nanos}}}
	pricingInfo := pricingInfo{pricingExpression{usageUnitDescription, tieredRates}}

	return ComputeEngineSKU{description, category, serviceRegions, pricingInfo}
}

// IsMatch checks if the SKU matches the description and region requirements.
func (sku *ComputeEngineSKU) IsMatch(description []string, region string) bool {
	for _, d := range description {
		if !strings.Contains(sku.description, d) {
			return false
		}
	}

	for _, r := range sku.serviceRegions {
		if r == region {
			return true
		}
	}

	return false
}

// GetPricingInfo returns the values of fields:
func (sku *ComputeEngineSKU) GetPricingInfo() (string, string, int64) {
	usageUnitDescription := sku.pricingInfo.pricingExpression.usageUnitDescription
	unitPrice := sku.pricingInfo.pricingExpression.tieredRates[0].unitPrice

	return usageUnitDescription, unitPrice.currencyCode, unitPrice.nanos
}
