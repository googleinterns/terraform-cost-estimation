package resources

import (
	"context"
	"fmt"

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

// Description holds information about additional information the SKU
// description contains/omits (Preemptible, Custom, Predefined etc.).
type Description struct {
	Contains []string
	Omits    []string
}

// BuildDescription returns a Description structure based on the description of an SKU.
func BuildDescription(custom, preemptible, predefined bool) (d Description) {
	if custom {
		d.Contains = append(d.Contains, "Custom")
	} else {
		d.Omits = append(d.Omits, "Custom")
	}

	if preemptible {
		d.Contains = append(d.Contains, "Preemptible")
	} else {
		d.Omits = append(d.Omits, "Preemptible")
	}

	if predefined {
		d.Contains = append(d.Contains, "Predefined")
	} else {
		d.Omits = append(d.Omits, "Predefined")
	}
	return
}

// CoreInfo stores CPU core details.
type CoreInfo struct {
	Type          string
	Description   Description
	ResourceGroup string
	UsageType     string
	Number        int
	UnitPricing   PricingInfo
}

func (core *CoreInfo) getPricingInfo() PricingInfo {
	return core.UnitPricing
}

func (core *CoreInfo) isMatch(sku *billingpb.Sku, region string) bool {
	if core.Type == "" {
		return false
	}
	cond1 := billing.FitsDescription(sku, append(core.Description.Contains, core.Type+" ", "Instance Core"), core.Description.Omits)
	cond2 := billing.FitsCategory(sku, "Compute Engine", "Compute", core.ResourceGroup, core.UsageType)
	cond3 := billing.FitsRegion(sku, region)
	return cond1 && cond2 && cond3
}

func (core *CoreInfo) completePricingInfo(skus []*billingpb.Sku, region string) error {
	sku, err := getSKU(skus, core, region)

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
	Description   Description
	ResourceGroup string
	UsageType     string
	AmountGB      float64
	UnitPricing   PricingInfo
}

func (mem *MemoryInfo) getPricingInfo() PricingInfo {
	return mem.UnitPricing
}

func (mem *MemoryInfo) isMatch(sku *billingpb.Sku, region string) bool {
	if mem.Type == "" {
		return false
	}
	cond1 := billing.FitsDescription(sku, append(mem.Description.Contains, mem.Type+" ", "Instance Ram"), mem.Description.Omits)
	cond2 := billing.FitsCategory(sku, "Compute Engine", "Compute", mem.ResourceGroup, mem.UsageType)
	cond3 := billing.FitsRegion(sku, region)
	return cond1 && cond2 && cond3
}

func (mem *MemoryInfo) completePricingInfo(skus []*billingpb.Sku, region string) error {
	sku, err := getSKU(skus, mem, region)

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
func (instance *ComputeInstance) CompletePricingInfo(ctx context.Context) error {

	skus, err := billing.GetSKUs(ctx)

	if err != nil {
		return fmt.Errorf("an error occurred while looking for pricing information")
	}

	err1 := instance.Cores.completePricingInfo(skus, instance.Region)
	if err1 != nil {
		return fmt.Errorf("could not find core pricing information")
	}

	err2 := instance.Memory.completePricingInfo(skus, instance.Region)
	if err2 != nil {
		return fmt.Errorf("could not find memory pricing information")
	}
	return nil
}

// PrintPricingInfo prints the cost estimation in a readable format.
func (instance *ComputeInstance) PrintPricingInfo() {
}
