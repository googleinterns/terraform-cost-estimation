package web

import "fmt"

// Table holds the HTML table of pricing information for a resource.
type Table struct {
	Index       int
	Type        string
	Header      [2]string
	GeneralRows [][2]string
	PricingInfo [][8]string
	Total       [3]string
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
func (t *Table) AddComputeInstancePricing(priceUnit string, cpuCostPerUnit1, cpuCostPerUnit2 float64, cpuUnits1, cpuUnits2 int,
	memCostPerUnit1, memCostPerUnit2, memUnits1, memUnits2 float64) {

	cpuTot1 := cpuCostPerUnit1 * float64(cpuUnits1)
	cpuTot2 := cpuCostPerUnit2 * float64(cpuUnits2)
	memTot1 := memCostPerUnit1 * memUnits1
	memTot2 := memCostPerUnit2 * memUnits2
	dCPU := cpuTot2 - cpuTot1
	dMem := memTot2 - memTot1

	f1 := func(x float64) string { return fmt.Sprintf("%.6f USD/%s", x, priceUnit) }
	f2 := func(x float64) string { return fmt.Sprintf("%.2f", x) }
	f3 := func(x int) string { return fmt.Sprintf("%d", x) }

	t.PricingInfo = [][8]string{
		{"CPU", f1(cpuCostPerUnit1), f3(cpuUnits1), f1(cpuTot1), f1(cpuCostPerUnit2), f3(cpuUnits2), f1(cpuTot2), f1(dCPU)},
		{"RAM", f1(memCostPerUnit1), f2(memUnits1), f1(memTot1), f1(memCostPerUnit2), f2(memUnits2), f1(memTot2), f1(dMem)},
	}
	t.Total = [3]string{f1(cpuTot1 + memTot1), f1(cpuTot2 + memTot2), f1(dCPU + dMem)}
}

// AddComputeDiskGeneralInfo fills the table with general information about the resource change.
func (t *Table) AddComputeDiskGeneralInfo(name, id, action, diskType, zones, image, snapshot string) {
	t.Header = [2]string{"Name", name}
	t.GeneralRows = [][2]string{
		{"ID", id},
		{"Action", action},
		{"Disk Type", diskType},
		{"Zones", zones},
		{"Image", image},
		{"Snapshot", snapshot},
	}
}

// AddComputeDiskPricing fills the table with the pricing information section for all billing components.
func (t *Table) AddComputeDiskPricing(priceUnit string, costPerUnit1, costPerUnit2 float64, units1, units2 int64, delta float64) {
	f1 := func(x float64) string { return fmt.Sprintf("%.6f USD/%s", x, priceUnit) }
	f2 := func(x int64) string { return fmt.Sprintf("%d", x) }

	tot1 := costPerUnit1 * float64(units1)
	tot2 := costPerUnit2 * float64(units2)

	t.PricingInfo = [][8]string{
		{"Disk", f1(costPerUnit1), f2(units1), f1(tot1), f1(costPerUnit2), f2(units2), f1(tot2), f1(delta)},
	}
	t.Total = [3]string{f1(tot1), f1(tot2), f1(delta)}
}
