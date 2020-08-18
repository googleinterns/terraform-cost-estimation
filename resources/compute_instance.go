package resources

import (
	"context"
	"fmt"
	"os"
	"strings"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
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

func (core *CoreInfo) isMatch(sku *billingpb.Sku) bool {
	if core.Type == "" {
		return false
	}
	cond1 := billing.FitsDescription(sku, append(core.Description.Contains, core.Type+" ", "Instance Core"), core.Description.Omits)
	cond2 := billing.FitsCategory(sku, "Compute Engine", "Compute", core.ResourceGroup, core.UsageType)
	return cond1 && cond2
}

func (core *CoreInfo) completePricingInfo(skus []*billingpb.Sku) error {
	sku := getSKU(core, skus)

	if sku == nil {
		return fmt.Errorf("could not find core pricing information")
	}

	usageUnit, hourlyUnitPrice, currencyType, currencyUnit := billing.GetPricingInfo(sku)
	core.UnitPricing = PricingInfo{usageUnit, hourlyUnitPrice, currencyType, currencyUnit}
	return nil
}

func (core *CoreInfo) getTotalPrice() float64 {
	nano := float64(1000 * 1000 * 1000)
	return float64(core.UnitPricing.HourlyUnitPrice*int64(core.Number)) / nano
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

func (mem *MemoryInfo) isMatch(sku *billingpb.Sku) bool {
	if mem.Type == "" {
		return false
	}
	cond1 := billing.FitsDescription(sku, append(mem.Description.Contains, mem.Type+" ", "Instance Ram"), mem.Description.Omits)
	cond2 := billing.FitsCategory(sku, "Compute Engine", "Compute", mem.ResourceGroup, mem.UsageType)
	return cond1 && cond2
}

func (mem *MemoryInfo) completePricingInfo(skus []*billingpb.Sku) error {
	sku := getSKU(mem, skus)

	if sku == nil {
		return fmt.Errorf("could not find memory pricing information")
	}

	usageUnit, hourlyUnitPrice, currencyType, currencyUnit := billing.GetPricingInfo(sku)
	mem.UnitPricing = PricingInfo{usageUnit, hourlyUnitPrice, currencyType, currencyUnit}
	return nil
}

func (mem *MemoryInfo) getTotalPrice() (float64, error) {
	nano := float64(1000 * 1000 * 1000)
	unitType := strings.Split(mem.UnitPricing.UsageUnit, " ")[0]
	unitsNum, err := conv.Convert("gb", mem.AmountGB, unitType)

	if err != nil {
		return 0, err
	}

	return float64(mem.UnitPricing.HourlyUnitPrice) * unitsNum / nano, nil
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

// CompletePricingInfo fills the pricing information fields.
func (instance *ComputeInstance) CompletePricingInfo(ctx context.Context) error {

	skus, err := billing.GetSKUs(ctx)
	if err != nil {
		return fmt.Errorf("an error occurred while looking for pricing information")
	}

	filtered, err := billing.RegionFilter(skus, instance.Region)

	if err != nil {
		return err
	}

	err1 := instance.Cores.completePricingInfo(filtered)
	if err1 != nil {
		return err1
	}

	err2 := instance.Memory.completePricingInfo(filtered)
	if err2 != nil {
		return err2
	}

	return nil
}

// ComputeInstanceState holds the before and after states of a compute instance
// and the action performed (created, destroyed etc.)
type ComputeInstanceState struct {
	Before *ComputeInstance
	After  *ComputeInstance
	Action string
}

// CompletePricingInfo completes pricing information of both before and after states.
func (state *ComputeInstanceState) CompletePricingInfo(ctx context.Context) error {
	err1 := state.Before.CompletePricingInfo(ctx)
	if err1 != nil {
		return err1
	}

	err2 := state.After.CompletePricingInfo(ctx)
	if err2 != nil {
		return err2
	}
	return nil
}

// PrintPricingInfo outputs the pricing estimation in a file/terminal.
func (state *ComputeInstanceState) PrintPricingInfo(f *os.File) {

}
