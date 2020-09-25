package js

// JSONOut is a general interface of a JSON output.
type JSONOut interface {
	AddToJSONTableList(*JsonOutput)
}

// JsonOutput contains relevant information resources and cost changes in a file.
type JsonOutput struct {
	Delta                   float64                    `json:"cost_change"`
	PricingUnit             string                     `json:"pricing_unit"`
	ComputeInstancesPricing []*ComputeInstanceStateOut `json:"instances_pricing_info"`
	ComputeDisksPricing     []*ComputeDiskStateOut     `json:"disks_pricing_info"`
}

// ComputeInstanceStateOut contains ComputeInstanceState information to be outputted.
type ComputeInstanceStateOut struct {
	Name        Change               `json:"name"`
	InstanceID  Change               `json:"instance_id"`
	Zone        Change               `json:"zone"`
	MachineType Change               `json:"machine_type"`
	CpuType     Change               `json:"cpu_type"`
	RamType     Change               `json:"ram_type"`
	Action      string               `json:"action"`
	Pricing     InstanceStatePricing `json:"pricing_info"`
}

func (out *ComputeInstanceStateOut) AddToJSONTableList(json *JsonOutput) {
	json.ComputeInstancesPricing = append(json.ComputeInstancesPricing, out)
}

// ComputeDiskStateOut contains ComputeDiskState information to be outputted.
type ComputeDiskStateOut struct {
	Name     Change           `json:"name"`
	ID       Change           `json:"id"`
	Zones    Change           `json:"zones"`
	DiskType Change           `json:"disk_type"`
	Action   string           `json:"action"`
	Pricing  DiskStatePricing `json:"pricing_info"`
}

func (out *ComputeDiskStateOut) AddToJSONTableList(json *JsonOutput) {
	json.ComputeDisksPricing = append(json.ComputeDisksPricing, out)
}

// InstanceStatePricing contains ComputeInstanceState pricing info to be outputted.
type InstanceStatePricing struct {
	Before   *InstancePricing `json:"before"`
	After    *InstancePricing `json:"after"`
	DeltaCpu float64          `json:"cpu_cost_change"`
	DeltaRam float64          `json:"ram_cost_change"`
	Delta    float64          `json:"cost_change"`
}

// DiskStatePricing contains ComputeDiskState pricing info to be outputted.
type DiskStatePricing struct {
	Before *DiskPricing `json:"before"`
	After  *DiskPricing `json:"after"`
	Delta  float64      `json:"cost_change"`
}

// InstancePricing contains ComputeInstance pricing info to be outputted.
type InstancePricing struct {
	Cpu       Pricing `json:"cpu"`
	Ram       Pricing `json:"ram"`
	TotalCost float64 `json:"total_cost"`
}

// DiskPricing contains ComputeDisk pricing info to be outputted.
type DiskPricing struct {
	Disk Pricing `json:"disk"`
}

// Pricing contains the pricing details about a certain  component.
type Pricing struct {
	//if the cost of unit is unknown we use string "-"
	UnitCost  string `json:"cost_per_unit"`
	NumUnits  string `json:"number_of_units"`
	TotalCost string `json:"cost_of_units"`
}

// Change contains before and after value of the certain field.
type Change struct {
	Before string `json:"before"`
	After  string `json:"after"`
}
