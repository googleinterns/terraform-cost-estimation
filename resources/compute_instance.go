package resources

import (
	"context"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

// PricingInfo stores the information from the billing API.
type PricingInfo struct {
	UsageUnit       string
	HourlyUnitPrice int64
	CurrencyType    string
	CurrencyUnit    string
}

// CoreInfo stores CPU core details.
type CoreInfo struct {
	Type          string
	ResourceGroup string
	UsageType     string
	Number        int
	UnitPricing   PricingInfo
}

func (core *CoreInfo) getPricingInfo() PricingInfo {
	return core.UnitPricing
}

func (core *CoreInfo) isMatch(sku *billingpb.Sku, region string) bool {
	cond1 := billing.FitsDescription(sku, []string{core.Type + " ", "Instance Core"}, []string{})
	cond2 := billing.FitsCategory(sku, "Compute Engine", "Compute", core.ResourceGroup, core.UsageType)
	cond3 := billing.FitsRegion(sku, region)
	return cond1 && cond2 && cond3
}

func (core *CoreInfo) completePricingInfo(ctx context.Context,
	getSKUs func(context.Context) ([]*billingpb.Sku, error), region string) error {
	sku, err := getSKU(ctx, core, getSKUs, region)

	if err != nil {
		return err
	}

	usageUnit, hourlyUnitPrice, currencyType, currencyUnit := billing.GetPricingInfo(sku)
	core.UnitPricing = PricingInfo{usageUnit, hourlyUnitPrice, currencyType, currencyUnit}
	return nil
}

// MemoryInfo stores memory details.
type MemoryInfo struct {
	Type          string
	ResourceGroup string
	UsageType     string
	AmountGB      float64
	UnitPricing   PricingInfo
}

func (mem *MemoryInfo) getPricingInfo() PricingInfo {
	return mem.UnitPricing
}

func (mem *MemoryInfo) isMatch(sku *billingpb.Sku, region string) bool {
	cond1 := billing.FitsDescription(sku, []string{mem.Type + " ", "Instance Ram"}, []string{})
	cond2 := billing.FitsCategory(sku, "Compute Engine", "Compute", mem.ResourceGroup, mem.UsageType)
	cond3 := billing.FitsRegion(sku, region)
	return cond1 && cond2 && cond3
}

func (mem *MemoryInfo) completePricingInfo(ctx context.Context,
	getSKUs func(context.Context) ([]*billingpb.Sku, error), region string) error {
	sku, err := getSKU(ctx, mem, getSKUs, region)

	if err != nil {
		return err
	}

	usageUnit, hourlyUnitPrice, currencyType, currencyUnit := billing.GetPricingInfo(sku)
	mem.UnitPricing = PricingInfo{usageUnit, hourlyUnitPrice, currencyType, currencyUnit}
	return nil
}

// ComputeInstance stores information about the compute instance resource type.
type ComputeInstance struct {
	ID          string
	Name        string
	MachineType string
	Region      string
	Memory      MemoryInfo
	Cores       CoreInfo
}

// ExtractResource extracts the resource details from the JSON object
// and fills the necessary fields.
func (instance *ComputeInstance) ExtractResource(jsonObject interface{}) {
}

// CompletePricingInfo fills the pricing information fields.
func (instance *ComputeInstance) CompletePricingInfo(ctx context.Context) {
	instance.Cores.completePricingInfo(ctx, billing.GetSKUs, instance.Region)
	instance.Memory.completePricingInfo(ctx, billing.GetSKUs, instance.Region)
}

// PrintPricingInfo prints the cost estimation in a readable format.
func (instance *ComputeInstance) PrintPricingInfo() {
}
