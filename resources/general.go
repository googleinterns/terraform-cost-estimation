package resources

import (
	"os"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	"github.com/googleinterns/terraform-cost-estimation/io/web"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

const (
	nano            = float64(1000 * 1000 * 1000)
	epsilon         = 1e-10
	hourlyToMonthly = float64(24 * 30)
	hourlyToYearly  = float64(24 * 365)
)

// PricingInfo stores the information from the billing API.
type PricingInfo struct {
	UsageUnit       string
	HourlyUnitPrice float64
	CurrencyType    string
}

func (p *PricingInfo) fillHourlyBase(sku *billingpb.Sku) {
	p.UsageUnit, p.HourlyUnitPrice, p.CurrencyType = billing.PricingInfo(sku)
}

func (p *PricingInfo) fillMonthlyBase(sku *billingpb.Sku) {
	usageUnit, monthly, currencyType := billing.PricingInfo(sku)
	p.UsageUnit = usageUnit
	p.HourlyUnitPrice = monthly / hourlyToMonthly
	p.CurrencyType = currencyType
}

//ResourceState is the interface of a general before/after resource state(ComputeInstance,...).
type ResourceState interface {
	CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error
	WritePricingInfo(f *os.File)
	GetWebTables(stateNum int) *web.PricingTypeTables
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
