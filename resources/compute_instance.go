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

// CoreInfo stores CPU core details.
type CoreInfo struct {
	ResourceGroup string
	Number        int
	UnitPricing   PricingInfo
}

func (core *CoreInfo) getPricingInfo() PricingInfo {
	return core.UnitPricing
}

func (core *CoreInfo) isMatch(sku *billingpb.Sku) bool {
	return core.ResourceGroup == sku.Category.ResourceGroup
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
	ResourceGroup string
	AmountGB      float64
	UnitPricing   PricingInfo
}

func (mem *MemoryInfo) getPricingInfo() PricingInfo {
	return mem.UnitPricing
}

func (mem *MemoryInfo) isMatch(sku *billingpb.Sku) bool {
	return mem.ResourceGroup == sku.Category.ResourceGroup
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
	Description Description
	Region      string
	UsageType   string
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

	filtered, err = billing.DescriptionFilter(filtered, instance.Description.Contains, instance.Description.Omits)
	if err != nil {
		return err
	}

	filtered, err = billing.CategoryFilter(filtered, "Compute Instance", "Compute", instance.UsageType)
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

func (state *ComputeInstanceState) getDelta() (DCore, DMem float64, err error) {
	var core1, mem1, core2, mem2 float64
	if state.Before != nil {
		core1 = state.Before.Cores.getTotalPrice()
		mem1, err = state.Before.Memory.getTotalPrice()
		if err != nil {
			return 0, 0, err
		}
	}

	if state.After != nil {
		core2 = state.After.Cores.getTotalPrice()
		mem2, err = state.After.Memory.getTotalPrice()
		if err != nil {
			return 0, 0, err
		}
	}

	return core2 - core1, mem2 - mem1, nil
}

// PrintPricingInfo outputs the pricing estimation in a file/terminal.
func (state *ComputeInstanceState) PrintPricingInfo(f *os.File) {

}
