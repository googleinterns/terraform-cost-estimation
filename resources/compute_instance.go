package resources

import (
	"fmt"
	"os"
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

	core.UnitPricing.fillInfo(sku)
	core.Type = sku.Description
	return nil
}

func (core *CoreInfo) getTotalPrice() float64 {
	return float64(core.UnitPricing.HourlyUnitPrice*int64(core.Number)) / nano * core.Fractional
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

	mem.UnitPricing.fillInfo(sku)
	mem.Type = sku.Description
	return nil
}

func (mem *MemoryInfo) getTotalPrice() (float64, error) {
	unitType := strings.Split(mem.UnitPricing.UsageUnit, " ")[0]
	unitsNum, err := conv.Convert("gib", mem.AmountGiB, unitType)

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
	instance.Name = name
	instance.MachineType = machineType
	instance.Zone = zone

	i := strings.LastIndex(zone, "-")
	if i < 0 {
		return nil, fmt.Errorf("invalid zone format")
	}
	instance.Region = zone[:i]

	instance.UsageType = usageType
	err := instance.Description.fillForComputeInstance(machineType, usageType)
	if err != nil {
		return nil, err
	}

	instance.Cores.Number, instance.Memory.AmountGiB, err = cd.GetMachineDetails(machineType)
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

	instance.Cores.Fractional = cd.GetMachineFractionalCore(machineType)

	return instance, nil
}

func (instance *ComputeInstance) filterSKUs(skus []*billingpb.Sku) ([]*billingpb.Sku, error) {
	filtered, err := billing.RegionFilter(skus, instance.Region)
	if err != nil {
		return nil, err
	}

	filtered, err = billing.DescriptionFilter(filtered, instance.Description.Contains, instance.Description.Omits)
	if err != nil {
		return nil, err
	}
	return filtered, nil
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

	filteredCores, err := instance.filterSKUs(cores)
	if err != nil {
		return err
	}

	filteredRAM, err := instance.filterSKUs(mem)
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

// ComputeInstanceState holds the before and after states of a compute instance
// and the action performed (created, destroyed etc.)
type ComputeInstanceState struct {
	Before *ComputeInstance
	After  *ComputeInstance
	Action string
}

// CompletePricingInfo completes pricing information of both before and after states.
func (state *ComputeInstanceState) CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error {
	if state.Before != nil {
		if err := state.Before.CompletePricingInfo(catalog); err != nil {
			return fmt.Errorf(state.Before.Name + "(" + state.After.MachineType + ")" + ": " + err.Error())
		}
	}

	if state.After != nil {
		if err := state.After.CompletePricingInfo(catalog); err != nil {
			return fmt.Errorf(state.After.Name + "(" + state.After.MachineType + ")" + ": " + err.Error())
		}
	}
	return nil
}

func (state *ComputeInstanceState) getDelta() (DCore, DMem float64, err error) {
	var core1, mem1, core2, mem2 float64
	if state.Before != nil {
		core1 = state.Before.Cores.getTotalPrice()
		mem1, err = state.Before.Memory.getTotalPrice()
		if err != nil {
			return 0, 0, fmt.Errorf(state.Before.Name + "(" + state.Before.MachineType + ")" + ": " + err.Error())
		}
	}

	if state.After != nil {
		core2 = state.After.Cores.getTotalPrice()
		mem2, err = state.After.Memory.getTotalPrice()
		if err != nil {
			return 0, 0, fmt.Errorf(state.After.Name + "(" + state.After.MachineType + ")" + ": " + err.Error())
		}
	}

	return core2 - core1, mem2 - mem1, nil
}

// WritePricingInfo outputs the pricing estimation in a file/terminal.
func (state *ComputeInstanceState) WritePricingInfo(f *os.File) {
	if f == nil {
		return
	}
	a := state.After
	c := a.Cores.getTotalPrice()
	m, _ := a.Memory.getTotalPrice()
	f.Write([]byte(fmt.Sprintf("%s -> Cores: %+v, Memory: %+v, Total: %+v\n\n", a.MachineType, c, m, c+m)))
}

// GetSummary returns a summary of the cost change for a compute instance state.
func (state *ComputeInstanceState) GetSummary() string {
	format := "Name: %s, Machine Type: %s, Action: %s, Total Cost: %f USD/hour, %s by %f USD/hour\n"
	var instance *ComputeInstance
	var change string

	dCore, dMem, err := state.getDelta()
	if err != nil {
		if state.After == nil {
			return "Could not make summary for compute instance initially named " + state.Before.Name + "\n"
		}
		return "Could not make summary for compute instance finally named " + state.After.Name + "\n"
	}

	if dCore+dMem < 0 {
		change = "Down"
	} else {
		change = "Up"
	}

	if state.After == nil {
		instance = state.Before
	} else {
		instance = state.After
	}

	c := instance.Cores.getTotalPrice()
	m, _ := instance.Memory.getTotalPrice()

	return fmt.Sprintf(format, instance.Name, instance.MachineType, state.Action, c+m, change, dCore+dMem)
}

func (state *ComputeInstanceState) getGeneralChanges() (name, ID, action,
	machineType, zone, cpuType, memType string) {
	action = state.Action

	switch {
	case state.Before == nil:
		name = state.After.Name

		if state.After.ID == "" {
			ID = "unknown"
		}
		machineType = state.After.MachineType
		zone = state.After.Zone
		cpuType = state.After.Cores.Type
		memType = state.After.Memory.Type

	case state.After == nil:
		name = state.Before.Name

		if state.Before.ID == "" {
			ID = "unknown"
		}
		machineType = state.Before.MachineType
		zone = state.Before.Zone
		cpuType = state.Before.Cores.Type
		memType = state.Before.Memory.Type

	default:
		name = state.Before.Name + " -> " + state.After.Name
		if state.After.ID == "" {
			ID = "unknown"
		}

		if state.Before.MachineType != state.After.MachineType {
			machineType = state.Before.MachineType + " -> " + state.After.MachineType
		} else {
			machineType = state.Before.MachineType
		}

		if state.Before.Zone != state.After.Zone {
			zone = state.Before.Zone + " -> " + state.After.Zone
		} else {
			zone = state.Before.Zone
		}

		if state.Before.Cores.Type != state.After.Cores.Type {
			cpuType = state.Before.Cores.Type + " -> " + state.After.Cores.Type
			memType = state.Before.Memory.Type + " -> " + state.After.Memory.Type
		} else {
			cpuType = state.Before.Cores.Type
			memType = state.Before.Memory.Type
		}
	}
	return
}

func (state *ComputeInstanceState) getCostChanges() (cpuCostPerUnit1, cpuCostPerUnit2, cpuUnits1, cpuUnits2,
	memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2 float64) {

	if state.Before != nil {
		cpuCostPerUnit1 = float64(state.Before.Cores.UnitPricing.HourlyUnitPrice) / nano
		cpuUnits1 = float64(state.Before.Cores.Number)
		memCostPerUnit1 = float64(state.Before.Memory.UnitPricing.HourlyUnitPrice) / nano
		memUnits1, _ = conv.Convert("gib", state.Before.Memory.AmountGiB, state.Before.Memory.UnitPricing.UsageUnit)
	}

	if state.After != nil {
		cpuCostPerUnit2 = float64(state.After.Cores.UnitPricing.HourlyUnitPrice) / nano
		cpuUnits2 = float64(state.After.Cores.Number)
		memCostPerUnit2 = float64(state.After.Memory.UnitPricing.HourlyUnitPrice) / nano
		memUnits2, _ = conv.Convert("gib", state.After.Memory.AmountGiB, strings.Split(state.After.Memory.UnitPricing.UsageUnit, " ")[0])
	}

	return
}

// GetWebTables returns html pricing information table strings to be displayed in a web page.
func (state *ComputeInstanceState) GetWebTables(stateNum int) *web.PricingTypeTables {
	name, ID, action, machineType, zone, cpuType, memType := state.getGeneralChanges()
	cpuCostPerUnit1, cpuCostPerUnit2, cpuUnits1, cpuUnits2,
		memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2 := state.getCostChanges()

	hourlyToMonthly := float64(24 * 30)
	hourlyToYearly := float64(24 * 365)

	h := web.Table{Index: stateNum, Type: "hourly"}
	h.AddComputeInstanceGeneralInfo(name, ID, action, machineType, zone, cpuType, memType)
	h.AddComputeInstancePricing(cpuCostPerUnit1, cpuCostPerUnit2, cpuUnits1, cpuUnits2,
		memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2)

	m := web.Table{Index: stateNum, Type: "monthly"}
	m.AddComputeInstanceGeneralInfo(name, ID, action, machineType, zone, cpuType, memType)
	m.AddComputeInstancePricing(cpuCostPerUnit1*hourlyToMonthly, cpuCostPerUnit2*hourlyToMonthly, cpuUnits1, cpuUnits2,
		memCostPerUnit1*hourlyToMonthly, memCostPerUnit2*hourlyToMonthly, memUnits1, memUnits2)

	y := web.Table{Index: stateNum, Type: "yearly"}
	y.AddComputeInstanceGeneralInfo(name, ID, action, machineType, zone, cpuType, memType)
	y.AddComputeInstancePricing(cpuCostPerUnit1*hourlyToYearly, cpuCostPerUnit2*hourlyToYearly, cpuUnits1, cpuUnits2,
		memCostPerUnit1*hourlyToYearly, memCostPerUnit2*hourlyToYearly, memUnits1, memUnits2)

	return &web.PricingTypeTables{Hourly: h, Monthly: m, Yearly: y}
}
