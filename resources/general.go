package resources

import (
	"os"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

const (
	nano    = float64(1000 * 1000 * 1000)
	epsilon = 1e-10
)

// PricingInfo stores the information from the billing API.
type PricingInfo struct {
	UsageUnit       string
	HourlyUnitPrice int64
	CurrencyType    string
	CurrencyUnit    string
}

func (p *PricingInfo) fillInfo(sku *billingpb.Sku) {
	usageUnit, hourlyUnitPrice, currencyType, currencyUnit := billing.GetPricingInfo(sku)
	p.UsageUnit = usageUnit
	p.HourlyUnitPrice = hourlyUnitPrice
	p.CurrencyType = currencyType
	p.CurrencyUnit = currencyUnit
}

//ResourceState is the interface of a general before/after resource state(ComputeInstance,...).
type ResourceState interface {
	CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error
	WritePricingInfo(f *os.File)
	GetWebTables(stateNum int) (hourly, monthly, yearly string)
	GetSummary() string
}

// skuObject is the interface for SKU types (core, memory etc.)
// that can be looked up in the billing catalog.
type skuObject interface {
	isMatch(sku *billingpb.Sku) bool
	completePricingInfo(skus []*billingpb.Sku) error
	getPricingInfo() PricingInfo
}

func findMatchingSKU(skuObj skuObject, skus []*billingpb.Sku) *billingpb.Sku {
	for _, sku := range skus {
		if skuObj.isMatch(sku) {
			return sku
		}
	}
	return nil
}
