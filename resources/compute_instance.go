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

// Description holds information about information the SKU
// description contains/omits (Preemptible, Custom, Type etc.).
type Description struct {
	Contains []string
	Omits    []string
}

func (d *Description) fill(machineType, usageType string) error {
	anythingButN1 := []string{"N2", "N2D", "E2", "Compute", "Memory", "Sole Tenancy"}

	if usageType == "Preemptible" {
		d.Contains = append(d.Contains, "Preemptible")
	} else {
		d.Omits = append(d.Omits, "Preemptible")
	}

	if strings.HasPrefix(usageType, "Commit") {
		d.Contains = append(d.Contains, "Commitment")
		if strings.Contains(machineType, "n1") {
			d.Omits = append(d.Omits, "N1")
			d.Omits = append(d.Omits, anythingButN1...)
		}
	} else {
		d.Omits = append(d.Omits, "Commitment")
	}

	if strings.Contains(machineType, "custom") {
		d.Contains = append(d.Contains, "Custom")
	} else {
		d.Omits = append(d.Omits, "Custom")
	}

	if strings.HasPrefix(machineType, "custom") {
		d.Omits = append(d.Omits, "N1")
		d.Omits = append(d.Omits, anythingButN1...)
	} else {

		switch {
		case strings.HasPrefix(machineType, "c2-"):
			d.Contains = append(d.Contains, "Compute")
		case strings.HasPrefix(machineType, "m1-") || strings.HasPrefix(machineType, "m2-"):
			d.Contains = append(d.Contains, "Memory")
		case strings.HasPrefix(machineType, "n1-mega") || strings.HasPrefix(machineType, "n1-ultra"):
			d.Contains = append(d.Contains, "Memory")
		case strings.HasPrefix(machineType, "n1-"):
			if !strings.HasPrefix(usageType, "Commit") {
				d.Contains = append(d.Contains, "N1")
			}
		default:
			i := strings.Index(machineType, "-")
			if i < 0 {
				return fmt.Errorf("wrong machine type format")
			}

			d.Contains = append(d.Contains, strings.ToUpper(machineType[:i]))
		}
	}

	return nil
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
	return core.ResourceGroup == sku.Category.ResourceGroup && !strings.Contains(sku.Description, "Ram")
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
	return mem.ResourceGroup == sku.Category.ResourceGroup && !strings.Contains(sku.Description, "Core")
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
	Zone        string
	UsageType   string
	Memory      MemoryInfo
	Cores       CoreInfo
}

// NewComputeInstance builds a compute instance with the specified fields
// and fills the other resource details.
// Returns a pointer to a ComputeInstance structure.
func NewComputeInstance(id, name, machineType, zone, usageType string) (*ComputeInstance, error) {
	instance := new(ComputeInstance)

	instance.ID = id
	instance.MachineType = machineType
	instance.Zone = zone

	i := strings.LastIndex(zone, "-")
	if i < 0 {
		return nil, fmt.Errorf("invalid zone format")
	}
	instance.Region = zone[:i]

	instance.UsageType = usageType
	err := instance.Description.fill(machineType, usageType)
	if err != nil {
		return nil, err
	}

	return instance, nil
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
