package resources

import (
	"fmt"
	"os"
	"strings"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
	cd "github.com/googleinterns/terraform-cost-estimation/resources/classdetail"
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
		if !strings.HasPrefix(machineType, "e2") {
			d.Contains = append(d.Contains, "Custom")
		}
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
			d.Omits = append(d.Omits, "Upgrade")
		case strings.HasPrefix(machineType, "n1-mega") || strings.HasPrefix(machineType, "n1-ultra"):
			d.Contains = append(d.Contains, "Memory")
			d.Omits = append(d.Omits, "Upgrade")
		case strings.HasPrefix(machineType, "n1-") || strings.HasPrefix(machineType, "f1-") || strings.HasPrefix(machineType, "g1-"):
			if !strings.HasPrefix(usageType, "Commit") {
				d.Contains = append(d.Contains, "N1")
			}
		default:
			i := strings.Index(machineType, "-")
			if i < 0 {
				return fmt.Errorf("wrong machine type format")
			}

			d.Contains = append(d.Contains, strings.ToUpper(machineType[:i])+" ")
		}
	}

	return nil
}

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
	sku := getSKU(core, skus)

	if sku == nil {
		return fmt.Errorf("could not find core pricing information")
	}

	usageUnit, hourlyUnitPrice, currencyType, currencyUnit := billing.GetPricingInfo(sku)
	core.UnitPricing = PricingInfo{usageUnit, hourlyUnitPrice, currencyType, currencyUnit}
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
	sku := getSKU(mem, skus)

	if sku == nil {
		return fmt.Errorf("could not find memory pricing information")
	}

	usageUnit, hourlyUnitPrice, currencyType, currencyUnit := billing.GetPricingInfo(sku)
	mem.UnitPricing = PricingInfo{usageUnit, hourlyUnitPrice, currencyType, currencyUnit}
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
	err := instance.Description.fill(machineType, usageType)
	if err != nil {
		return nil, err
	}

	instance.Cores.Number, instance.Memory.AmountGiB, err = cd.GetMachineDetails(machineType)
	if err != nil {
		return nil, err
	}

	instance.Memory.Extended = strings.Contains(machineType, "custom") && strings.HasSuffix(machineType, "-ext")

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
	cores, err1 := catalog.GetCoreSKUs(instance.UsageType)
	if err1 != nil {
		return err1
	}

	mem, err2 := catalog.GetRAMSKUs(instance.UsageType)
	if err2 != nil {
		return err2
	}

	filteredCores, err1 := instance.filterSKUs(cores)
	if err1 != nil {
		return err1
	}

	filteredRAM, err2 := instance.filterSKUs(mem)
	if err2 != nil {
		return err2
	}

	err1 = instance.Cores.completePricingInfo(filteredCores)
	if err1 != nil {
		return err1
	}

	err2 = instance.Memory.completePricingInfo(filteredRAM)
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
func (state *ComputeInstanceState) CompletePricingInfo(catalog *billing.ComputeEngineCatalog) error {
	if state.Before != nil {
		err1 := state.Before.CompletePricingInfo(catalog)
		if err1 != nil {
			return fmt.Errorf(state.Before.Name + "(" + state.After.MachineType + ")" + ": " + err1.Error())
		}
	}

	if state.After != nil {
		err2 := state.After.CompletePricingInfo(catalog)
		if err2 != nil {
			return fmt.Errorf(state.After.Name + "(" + state.After.MachineType + ")" + ": " + err2.Error())
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
			return 0, 0, fmt.Errorf(state.Before.Name + "(" + state.After.MachineType + ")" + ": " + err.Error())
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

func (state *ComputeInstanceState) getGeneralChanges() (name, ID, action, cpuType, memType string) {
	action = state.Action

	switch {
	case state.Before == nil:
		name = state.After.Name

		if state.After.ID == "" {
			ID = "unknown"
		}
		cpuType = state.After.Cores.Type
		memType = state.After.Memory.Type

	case state.After == nil:
		name = state.Before.Name

		if state.Before.ID == "" {
			ID = "unknown"
		}
		cpuType = state.Before.Cores.Type
		memType = state.Before.Memory.Type

	default:
		name = state.Before.Name + " -> " + state.After.Name
		if state.After.ID == "" {
			ID = "unknown"
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

func (state *ComputeInstanceState) getCostChanges() (cpuCostPerUnit1, cpuCostPerUnit2 float64,
	cpuUnits1, cpuUnits2 int, memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2 float64) {

	if state.Before != nil {
		cpuCostPerUnit1 = float64(state.Before.Cores.UnitPricing.HourlyUnitPrice) / nano
		cpuUnits1 = state.Before.Cores.Number
		memCostPerUnit1 = float64(state.Before.Memory.UnitPricing.HourlyUnitPrice) / nano
		memUnits1, _ = conv.Convert("gib", state.Before.Memory.AmountGiB, state.Before.Memory.UnitPricing.UsageUnit)
	}

	if state.After != nil {
		cpuCostPerUnit2 = float64(state.After.Cores.UnitPricing.HourlyUnitPrice) / nano
		cpuUnits2 = state.After.Cores.Number
		memCostPerUnit2 = float64(state.After.Memory.UnitPricing.HourlyUnitPrice) / nano
		memUnits2, _ = conv.Convert("gib", state.After.Memory.AmountGiB, strings.Split(state.After.Memory.UnitPricing.UsageUnit, " ")[0])
	}

	return
}

// GetWebTables returns html pricing information table strings to be displayed in a web page.
func (state *ComputeInstanceState) GetWebTables(stateNum int) (hourly, monthly, yearly string) {
	name, ID, action, cpuType, memType := state.getGeneralChanges()
	cpuCostPerUnit1, cpuCostPerUnit2, cpuUnits1, cpuUnits2, memCostPerUnit1,
		memCostPerUnit2, memUnits1, memUnits2 := state.getCostChanges()

	format := `
		<table class="table table-bordered" style="table-layout: fixed;">
            <thead class="table-info">
                <tr style="cursor: pointer;" data-toggle="collapse" data-target="#table_%+v_%s">
					<th colspan="1" style="width: 12.5%%;">Name</th>
					<th colspan="7" style="width:87.5%%;">%s</th></td>
                </tr>
            </thead>
            <tbody class="hide-table-padding collapse" id="table_%+v_%s">
				<tr>
                    <td colspan="1">Instance ID</td>
                    <td colspan="7">%s</td>
                </tr>
                <tr>
                    <td colspan="1">Action</td>
                    <td colspan="7">%s</td>
                </tr>
                <tr>
                    <td colspan="1">CPU type</td>
                    <td colspan="7">%s</td>
                </tr>
                <tr>
                    <td colspan="1">RAM Type</td>
                    <td colspan="7">%s</td>
                </tr>
                <tr>
                    <td colspan="8">Pricing information</td>
                </tr>
                <tr>
                    <td colspan="1"></td>
                    <td colspan="3">Before</td>
                    <td colspan="3">After</td>
                    <td colspan="1">Delta</td>
                </tr>
                <tr>
                    <td colspan="1"></td>
                    <td colspan="1">Cost per Unit</td>
                    <td colspan="1">Number of units</td>
                    <td colspan="1">Cost of units</td>
                    <td colspan="1">Cost per Unit</td>
                    <td colspan="1">Number of units</td>
                    <td colspan="1">Cost per Unit</td>
                    <td colspan="1">Cost of units</td>
                </tr>
                <tr>
                    <td colspan="1">CPU</td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v </td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v </td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v USD </td>
                </tr>        
                <tr>
                    <td colspan="1">RAM</td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v </td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v </td>
                    <td colspan="1"> %+v USD </td>
                    <td colspan="1"> %+v USD </td>
                </tr>
                <tr>
                    <td colspan="1">Total Cost</td>
                    <td colspan="3"> %+v USD </td>
                    <td colspan="3"> %+v USD </td>
                    <td colspan="1"> %+v USD </td>

                </tr>
            </tbody>
        </table>
	`
	cpuTotal1 := cpuCostPerUnit1 * float64(cpuUnits1)
	cpuTotal2 := cpuCostPerUnit2 * float64(cpuUnits2)
	memTotal1 := memCostPerUnit1 * memUnits1
	memTotal2 := memCostPerUnit2 * memUnits2

	hourlyToMonthly := float64(24 * 30)
	hourlyToYearly := float64(24 * 365)

	hourly = fmt.Sprintf(format, stateNum, "hourly", name, stateNum, "hourly", ID, action, cpuType, memType,
		cpuCostPerUnit1, cpuUnits1, cpuTotal1, cpuCostPerUnit2, cpuUnits2, cpuTotal2, cpuTotal2-cpuTotal1,
		memCostPerUnit1, memUnits1, memTotal1, memCostPerUnit2, memUnits2, memTotal2, memTotal2-memTotal1,
		cpuTotal1+memTotal1, cpuTotal2+memTotal2, cpuTotal2-cpuTotal1+memTotal2-memTotal1)

	monthly = fmt.Sprintf(format, stateNum, "monthly", name, stateNum, "monthly", ID, action, cpuType, memType,
		cpuCostPerUnit1*hourlyToMonthly, cpuUnits1, cpuTotal1*hourlyToMonthly, cpuCostPerUnit2*hourlyToMonthly, cpuUnits2, cpuTotal2*hourlyToMonthly, (cpuTotal2-cpuTotal1)*hourlyToMonthly,
		memCostPerUnit1*hourlyToMonthly, memUnits1, memTotal1*hourlyToMonthly, memCostPerUnit2*hourlyToMonthly, memUnits2, memTotal2*hourlyToMonthly, (memTotal2-memTotal1)*hourlyToMonthly,
		(cpuTotal1+memTotal1)*hourlyToMonthly, (cpuTotal2+memTotal2)*hourlyToMonthly, (cpuTotal2-cpuTotal1+memTotal2-memTotal1)*hourlyToMonthly)

	yearly = fmt.Sprintf(format, stateNum, "yearly", name, stateNum, "yearly", ID, action, cpuType, memType,
		cpuCostPerUnit1*hourlyToYearly, cpuUnits1, cpuTotal1*hourlyToYearly, cpuCostPerUnit2*hourlyToYearly, cpuUnits2, cpuTotal2*hourlyToYearly, (cpuTotal2-cpuTotal1)*hourlyToYearly,
		memCostPerUnit1*hourlyToYearly, memUnits1, memTotal1*hourlyToYearly, memCostPerUnit2*hourlyToYearly, memUnits2, memTotal2*hourlyToYearly, (memTotal2-memTotal1)*hourlyToYearly,
		(cpuTotal1+memTotal1)*hourlyToYearly, (cpuTotal2+memTotal2)*hourlyToYearly, (cpuTotal2-cpuTotal1+memTotal2-memTotal1)*hourlyToYearly)
	return
}
