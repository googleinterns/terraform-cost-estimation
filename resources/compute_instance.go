package resources

import (
	"fmt"
	"strings"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	"github.com/googleinterns/terraform-cost-estimation/io/web"
	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
	cd "github.com/googleinterns/terraform-cost-estimation/resources/classdetail"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

// CoreInfo stores CPU core details.
type CoreInfo struct {
	Type          string
	ResourceGroup string
	Number        int
	Fractional    float64
	UnitPricing   PricingInfo
}

func (core *CoreInfo) getPricingInfo() PricingInfo {
	return core.UnitPricing
}

func (core *CoreInfo) isMatch(sku *billingpb.Sku) bool {
	return core.ResourceGroup == sku.Category.ResourceGroup && !strings.Contains(sku.Description, "Ram")
}

func (core *CoreInfo) completePricingInfo(skus []*billingpb.Sku) error {
	sku := findMatchingSKU(core, skus)
	if sku == nil {
		return fmt.Errorf("could not find core pricing information")
	}

	core.UnitPricing.fillHourlyBase(sku, func(tr *billingpb.PricingExpression_TierRate) bool { return true })
	core.Type = sku.Description
	return nil
}

func (core *CoreInfo) getTotalPrice() float64 {
	return core.UnitPricing.HourlyUnitPrice * float64(core.Number) * core.Fractional
}

// MemoryInfo stores memory details.
type MemoryInfo struct {
	Type          string
	ResourceGroup string
	AmountGiB     float64
	Extended      bool
	UnitPricing   PricingInfo
}

func (mem *MemoryInfo) getPricingInfo() PricingInfo {
	return mem.UnitPricing
}

func (mem *MemoryInfo) isMatch(sku *billingpb.Sku) bool {
	return mem.ResourceGroup == sku.Category.ResourceGroup && !strings.Contains(sku.Description, "Core") &&
		strings.Contains(sku.Description, "Extended") == mem.Extended
}

func (mem *MemoryInfo) completePricingInfo(skus []*billingpb.Sku) error {
	sku := findMatchingSKU(mem, skus)
	if sku == nil {
		return fmt.Errorf("could not find memory pricing information")
	}

	mem.UnitPricing.fillHourlyBase(sku, func(tr *billingpb.PricingExpression_TierRate) bool { return true })
	// If the SKU memory unit is not supported, return error.
	if _, err := conv.Convert("gib", 0, mem.UnitPricing.UsageUnit); err != nil {
		return fmt.Errorf("memory unit of SKU is not supported")
	}

	mem.Type = sku.Description
	return nil
}

func (mem *MemoryInfo) getTotalPrice() float64 {
	unitsNum, _ := conv.Convert("gib", mem.AmountGiB, mem.UnitPricing.UsageUnit)
	return mem.UnitPricing.HourlyUnitPrice * unitsNum
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

// NewComputeInstance builds a compute instance with the specified fields nd fills the other resource details.
func NewComputeInstance(details *cd.ResourceDetail, id, name, machineType, zone, usageType string) (*ComputeInstance, error) {
	instance := &ComputeInstance{ID: id, Name: name, MachineType: machineType, Zone: zone, UsageType: usageType}

	i := strings.LastIndex(zone, "-")
	if i < 0 {
		return nil, fmt.Errorf("invalid zone format")
	}
	instance.Region = zone[:i]

	err := instance.Description.fillForComputeInstance(machineType, usageType)
	if err != nil {
		return nil, err
	}

	instance.Cores.Number, instance.Memory.AmountGiB, err = details.MachineDetails(machineType)
	if err != nil {
		return nil, err
	}

	instance.Memory.Extended = strings.Contains(machineType, "custom") && strings.HasSuffix(machineType, "-ext")

	// Only N1 predefined/preemptible type of cores/memory has N1Standard resource group.
	if (strings.HasPrefix(machineType, "n1-standard") || strings.HasPrefix(machineType, "n1-high") ||
		strings.HasPrefix(machineType, "f1-") || strings.HasPrefix(machineType, "g1-")) &&
		!strings.HasPrefix(usageType, "Commit") {
		instance.Memory.ResourceGroup = "N1Standard"
		instance.Cores.ResourceGroup = "N1Standard"
	} else {
		instance.Memory.ResourceGroup = "RAM"
		instance.Cores.ResourceGroup = "CPU"
	}

	instance.Cores.Fractional = details.MachineFractionalCore(machineType)

	return instance, nil
}

// CompletePricingInfo fills the pricing information fields.
func (instance *ComputeInstance) CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error {
	cores, err := catalog.GetCoreSKUs(instance.UsageType)
	if err != nil {
		return err
	}

	mem, err := catalog.GetRAMSKUs(instance.UsageType)
	if err != nil {
		return err
	}

	filteredCores, err := filterSKUs(cores, instance.Region, instance.Description)
	if err != nil {
		return err
	}

	filteredRAM, err := filterSKUs(mem, instance.Region, instance.Description)
	if err != nil {
		return err
	}

	if err = instance.Cores.completePricingInfo(filteredCores); err != nil {
		return err
	}

	if err = instance.Memory.completePricingInfo(filteredRAM); err != nil {
		return err
	}

	return nil
}

// ComputeInstanceState holds the before and after states of a compute instance and the action performed (created, destroyed etc.)
type ComputeInstanceState struct {
	Before *ComputeInstance
	After  *ComputeInstance
	Action string
}

// CompletePricingInfo completes pricing information of both before and after states.
func (state *ComputeInstanceState) CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error {
	if state.Before != nil {
		if err := state.Before.CompletePricingInfo(catalog); err != nil {
			return fmt.Errorf(state.Before.Name + "(" + state.Before.MachineType + ")" + ": " + err.Error())
		}
	}

	if state.After != nil {
		if err := state.After.CompletePricingInfo(catalog); err != nil {
			return fmt.Errorf(state.After.Name + "(" + state.After.MachineType + ")" + ": " + err.Error())
		}
	}
	return nil
}

func (state *ComputeInstanceState) getDeltas() (DCore, DMem float64) {
	var core1, mem1, core2, mem2 float64
	if state.Before != nil {
		core1 = state.Before.Cores.getTotalPrice()
		mem1 = state.Before.Memory.getTotalPrice()
	}

	if state.After != nil {
		core2 = state.After.Cores.getTotalPrice()
		mem2 = state.After.Memory.getTotalPrice()
	}

	return core2 - core1, mem2 - mem1
}

func (state *ComputeInstanceState) getDelta() float64 {
	dcore, dmem := state.getDeltas()
	return dcore + dmem
}

func (state *ComputeInstanceState) getGeneralChanges() (name, ID, action,
	machineType, zone, cpuType, memType string) {
	action = state.Action

	// Before and After can't be nil at the same time. Take return values from the non nil state or a combination of both.
	switch {
	case state.Before == nil:
		name = state.After.Name

		if state.After.ID == "" {
			ID = "unknown"
		} else {
			ID = state.After.ID
		}
		machineType = state.After.MachineType
		zone = state.After.Zone
		cpuType = state.After.Cores.Type
		memType = state.After.Memory.Type

	case state.After == nil:
		name = state.Before.Name
		ID = state.Before.ID
		machineType = state.Before.MachineType
		zone = state.Before.Zone
		cpuType = state.Before.Cores.Type
		memType = state.Before.Memory.Type

	default:
		name = generalChange(state.Before.Name, state.After.Name)
		ID = state.Before.ID
		machineType = generalChange(state.Before.MachineType, state.After.MachineType)
		zone = generalChange(state.Before.Zone, state.After.Zone)
		cpuType = generalChange(state.Before.Cores.Type, state.After.Cores.Type)
		memType = generalChange(state.Before.Memory.Type, state.After.Memory.Type)
	}
	return
}

func (state *ComputeInstanceState) getCostChanges() (cpuCostPerUnit1, cpuCostPerUnit2 float64, cpuUnits1, cpuUnits2 int,
	memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2 float64) {

	if state.Before != nil {
		cpuCostPerUnit1 = state.Before.Cores.UnitPricing.HourlyUnitPrice
		cpuUnits1 = state.Before.Cores.Number
		memCostPerUnit1 = state.Before.Memory.UnitPricing.HourlyUnitPrice
		memUnits1, _ = conv.Convert("gib", state.Before.Memory.AmountGiB, state.Before.Memory.UnitPricing.UsageUnit)
	}

	if state.After != nil {
		cpuCostPerUnit2 = state.After.Cores.UnitPricing.HourlyUnitPrice
		cpuUnits2 = state.After.Cores.Number
		memCostPerUnit2 = state.After.Memory.UnitPricing.HourlyUnitPrice
		memUnits2, _ = conv.Convert("gib", state.After.Memory.AmountGiB, state.After.Memory.UnitPricing.UsageUnit)
	}

	return
}

// GetWebTables returns html pricing information table with hourly, monthly and yearly pricing.
func (state *ComputeInstanceState) GetWebTables(stateNum int) *web.PricingTypeTables {
	name, ID, action, machineType, zone, cpuType, memType := state.getGeneralChanges()
	cpuCostPerUnit1, cpuCostPerUnit2, cpuUnits1, cpuUnits2,
		memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2 := state.getCostChanges()

	h := web.Table{Index: stateNum, Type: "hourly"}
	h.AddComputeInstanceGeneralInfo(name, ID, action, machineType, zone, cpuType, memType)
	h.AddComputeInstancePricing("hour", cpuCostPerUnit1, cpuCostPerUnit2, cpuUnits1, cpuUnits2,
		memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2)

	m := web.Table{Index: stateNum, Type: "monthly"}
	m.AddComputeInstanceGeneralInfo(name, ID, action, machineType, zone, cpuType, memType)
	m.AddComputeInstancePricing("month", cpuCostPerUnit1*hourlyToMonthly, cpuCostPerUnit2*hourlyToMonthly, cpuUnits1, cpuUnits2,
		memCostPerUnit1*hourlyToMonthly, memCostPerUnit2*hourlyToMonthly, memUnits1, memUnits2)

	y := web.Table{Index: stateNum, Type: "yearly"}
	y.AddComputeInstanceGeneralInfo(name, ID, action, machineType, zone, cpuType, memType)
	y.AddComputeInstancePricing("year", cpuCostPerUnit1*hourlyToYearly, cpuCostPerUnit2*hourlyToYearly, cpuUnits1, cpuUnits2,
		memCostPerUnit1*hourlyToYearly, memCostPerUnit2*hourlyToYearly, memUnits1, memUnits2)

	return &web.PricingTypeTables{Hourly: h, Monthly: m, Yearly: y}
}
