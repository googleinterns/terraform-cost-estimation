package resources

// PricingInfo structure for storing the information from the billing API
type PricingInfo struct {
	ComponentUnit   string
	HourlyUnitPrice int64
	CurrencyType    string
	CurrencyUnit    string
}

// CoreInfo structure for CPU core details
type CoreInfo struct {
	Type        string
	Preemptible bool
	Number      int
	Pricing     PricingInfo
}

// MemoryInfo structure for memory details
type MemoryInfo struct {
	Type        string
	Preemptible bool
	Amount      float64
	UnitType    string
	Pricing     PricingInfo
}

// ComputeInstance structure for compute instance resource type
type ComputeInstance struct {
	ID          string
	Name        string
	MachineType string
	Region      string
	Memory      *MemoryInfo
	Cores       *CoreInfo
}

// ExtractResource method that extracts the resource details from JSON file
// and fills the necessary fields.
func (instance *ComputeInstance) ExtractResource(jsonObject interface{}) {
}

// CompletePricingInfo method that fills the pricing information fields.
func (instance *ComputeInstance) CompletePricingInfo() {
}

// PrintPricingInfo method that prints the cost estimation in a readable format.
func (instance *ComputeInstance) PrintPricingInfo() {
}
