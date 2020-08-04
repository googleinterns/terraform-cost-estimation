package resources

// PricingInfo stores the information from the billing API.
type PricingInfo struct {
	ComponentUnit   string
	HourlyUnitPrice int64
	CurrencyType    string
	CurrencyUnit    string
}

// CoreInfo stores CPU core details.
type CoreInfo struct {
	Type        string
	Preemptible bool
	Number      int
	Pricing     PricingInfo
}

// MemoryInfo stores memory details.
type MemoryInfo struct {
	Type        string
	Preemptible bool
	Amount      float64
	UnitType    string
	Pricing     PricingInfo
}

// ComputeInstance stores information about the compute instance resource type.
type ComputeInstance struct {
	ID          string
	Name        string
	MachineType string
	Region      string
	Memory      *MemoryInfo
	Cores       *CoreInfo
}

// ExtractResource extracts the resource details from the JSON object
// and fills the necessary fields.
func (instance *ComputeInstance) ExtractResource(jsonResource interface{}) {
}

// CompletePricingInfo fills the pricing information fields.
func (instance *ComputeInstance) CompletePricingInfo() {
}

// PrintPricingInfo prints the cost estimation in a readable format.
func (instance *ComputeInstance) PrintPricingInfo() {
}
