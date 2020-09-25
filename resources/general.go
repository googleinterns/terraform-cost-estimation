package resources

import (
	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	"github.com/googleinterns/terraform-cost-estimation/io/web"
	"github.com/jedib0t/go-pretty/v6/table"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

// PricingInfo stores the information from the billing API.
type PricingInfo struct {
	UsageUnit       string
	HourlyUnitPrice float64
	CurrencyType    string
}

func (p *PricingInfo) fillHourlyBase(sku *billingpb.Sku, correctTieredRate func(*billingpb.PricingExpression_TierRate) bool) {
	p.UsageUnit, p.HourlyUnitPrice, p.CurrencyType = billing.PricingInfo(sku, correctTieredRate)
}

func (p *PricingInfo) fillMonthlyBase(sku *billingpb.Sku, correctTieredRate func(*billingpb.PricingExpression_TierRate) bool) {
	usageUnit, monthly, currencyType := billing.PricingInfo(sku, correctTieredRate)
	p.UsageUnit = usageUnit
	p.HourlyUnitPrice = monthly / hourlyToMonthly
	p.CurrencyType = currencyType
}

//ResourceState is the interface of a general before/after resource state(ComputeInstance,...).
type ResourceState interface {
	CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error
	getDelta() float64
	GetWebTables(stateNum int) *web.PricingTypeTables
	ToTable(colorful bool) (*table.Table, error)
	getSummaryRow() (table.Row, error)
	ToStateOut() (JSONOut, error)
}

// skuObject is the interface for CPU cores and RAM SKUs from the billing catalog.
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

// JSONOut is a general interface of a JSON output.
type JSONOut interface {
	addToJSONTableList(*JsonOutput)
}
