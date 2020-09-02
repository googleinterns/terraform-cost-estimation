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
		case strings.HasPrefix(machineType, "n1-"):
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
	return nil
}

func (core *CoreInfo) getTotalPrice() float64 {
	return float64(core.UnitPricing.HourlyUnitPrice*int64(core.Number)) / nano * core.Fractional
}

// MemoryInfo stores memory details.
type MemoryInfo struct {
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

	if (strings.HasPrefix(machineType, "n1-standard") || strings.HasPrefix(machineType, "n1-high")) && !strings.HasPrefix(usageType, "Commit") {
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

// PrintPricingInfo outputs the pricing estimation in a file/terminal.
func (state *ComputeInstanceState) PrintPricingInfo(f *os.File) {
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
