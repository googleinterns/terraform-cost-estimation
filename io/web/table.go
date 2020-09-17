package web

// Table holds the HTML table of pricing information for a resource.
type Table struct {
	Index             int
	Type              string
	Header            [2]string
	GeneralRows       [][2]string
	PricingComponents []string
	PricingInfo       [][7]float64
	Total             [3]float64
}

// PricingTypeTables holds the HTML tables of hourly, monthly and yearly pricing information for a resource.
type PricingTypeTables struct {
	Hourly  Table
	Monthly Table
	Yearly  Table
}

// AddComputeInstanceGeneralInfo fills the table with general information about the resource change.
func (t *Table) AddComputeInstanceGeneralInfo(name, ID, action, machineType, zone, cpuType, memType string) {
	t.Header = [2]string{"Name", name}
	t.GeneralRows = [][2]string{
		{"ID", ID},
		{"Action", action},
		{"Machine Type", machineType},
		{"Zone", zone},
		{"CPU Type", cpuType},
		{"RAM Type", memType},
	}
}

// AddComputeInstancePricing fills the table with the pricing information section for all billing components.
func (t *Table) AddComputeInstancePricing(cpuCostPerUnit1, cpuCostPerUnit2, cpuUnits1, cpuUnits2,
	memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2 float64) {
	t.PricingComponents = []string{"CPU", "RAM"}

	cpuTot1 := cpuCostPerUnit1 * cpuUnits1
	cpuTot2 := cpuCostPerUnit2 * cpuUnits2
	memTot1 := memCostPerUnit1 * memUnits1
	memTot2 := memCostPerUnit2 * memUnits2
	dCPU := cpuTot2 - cpuTot1
	dMem := memTot2 - memTot1

	t.PricingInfo = [][7]float64{
		{cpuCostPerUnit1, cpuUnits1, cpuTot1, cpuCostPerUnit2, cpuUnits2, cpuTot2, dCPU},
		{memCostPerUnit1, memUnits1, memTot1, memCostPerUnit2, memUnits2, memTot2, dMem},
	}
	t.Total = [3]float64{cpuTot1 + memTot1, cpuTot2 + memTot2, dCPU + dMem}
}
