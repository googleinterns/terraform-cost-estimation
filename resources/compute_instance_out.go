package resources

import (
	"encoding/json"
	"fmt"
	"io"
	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"os"
	"strings"
)

// The number of columns in output table for ComputeInstance.
const numColumns = 5

// ToTable creates a table.Table and fills it with the pricing information from
// ComputeInstanceState.
func (state *ComputeInstanceState) ToTable() (*table.Table, error) {
	t := &table.Table{}
	autoMerge := table.RowConfig{AutoMerge: true}
	before, after, err := syncInstances(state.Before, state.After)
	if err != nil {
		return nil, err
	}

	t.AppendRow(initRow("Name", before.Name, after.Name, false), autoMerge)
	t.AppendRow(initRow("Instance_ID", before.ID, after.ID, true), autoMerge)
	t.AppendRow(initRow("Zone", before.Zone, after.Zone, false), autoMerge)
	t.AppendRow(initRow("Machine type", before.MachineType, after.MachineType, true), autoMerge)
	t.AppendRow(initRow("Cpu type", before.Cores.Type, after.Cores.Type, false), autoMerge)
	t.AppendRow(initRow("Memory type", before.Memory.Type, after.Memory.Type, true), autoMerge)
	t.AppendRow(initRow("Action", state.Action, state.Action, false), autoMerge)

	t.AppendRow(addHeader("Pricing\nInformation"), autoMerge)

	core1, mem1, t1, err := getMemCoreInfo(state.Before)
	if err != nil {
		return nil, err
	}
	core2, mem2, t2, err := getMemCoreInfo(state.After)
	if err != nil {
		return nil, err
	}

	t1Str := fmt.Sprintf("%f", t1)
	// Add " " in the end of string to avoid unwanted auto-merging in the table package.
	t2Str := fmt.Sprintf("%f ", t2)

	t.AppendRow(table.Row{" ", " ", "Cpu", "Mem", "Total"}, autoMerge)
	t.AppendRows([]table.Row{
		{"Before", "Cost\nper\nunit", core1[0], mem1[0], t1Str},
		{"Before", "Number\nof\nunits", core1[1], mem1[1], t1Str},
		{"Before", "Units\ncost", core1[2], mem1[2], t1Str},
		{"After", "Cost\nper\nunit", core2[0], mem2[0], t2Str},
		{"After", "Number\nof\nunits", core2[1], mem2[1], t2Str},
		{"After", "Units\ncost", core2[2], mem2[2], t2Str},
	})

	dCore, dMem, err := state.getDelta()
	if err != nil {
		return nil, err
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 5, AutoMerge: true, ColorsFooter: text.Colors{text.FgGreen}},
	})
	change := "No change"
	if dTotal := dCore + dMem; dTotal < 0 {
		change = "Down (↓)"
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 5, AutoMerge: true, ColorsFooter: text.Colors{text.FgRed}},
		})
	} else if dTotal > 0 {
		change = "Up (↑)"
	}

	t.AppendFooter(table.Row{"DELTA", change, dCore, dMem, dCore + dMem})
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	return t, nil
}

// GetTotalSummary returns the table with brief cost changes info about all
// resources.
func GetTotalSummary(states []*ComputeInstanceState) *table.Table {
	t := &table.Table{}
	dTotal, _, _ := GetTotalDelta(states)
	t.AppendHeader(table.Row{fmt.Sprintf("The total cost change for all resources is %f.", dTotal)})

	t.AppendRow(addHeader("Pricing Information"))
	t.AppendRow(table.Row{"Name", "MachineType", "Action", "Change", "Delta"})
	for _, state := range states {
		if row, err := state.getSummary(); err == nil {
			t.AppendRow(row)
		}
	}
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	return t
}

// getSummary() returns the row for SummaryTable to be outputted about the certain state.
func (state *ComputeInstanceState) getSummary() (table.Row, error) {
	dCore, dMem, err := state.getDelta()
	if err != nil {
		return table.Row{}, err
	}

	change := "No change"
	if dTotal := dCore + dMem; dTotal > 0 {
		change = "Up (↑)"
	} else if dTotal < 0 {
		change = "Down (↓)"
	}

	var r *ComputeInstance
	if state.After == nil {
		r = state.Before
	} else {
		r = state.After
	}

	if r == nil {
		return table.Row{}, fmt.Errorf("Before and After resources can't be nil at the  same time.")
	}
	return table.Row{r.Name, r.MachineType, state.Action, change, dCore + dMem}, nil
}

// OutputPricing writes pricing information about each resource and summary for all
// resources of the array.
func OutputPricing(states []*ComputeInstanceState, f *os.File) {
	if f == nil {
		return
	}
	f.Write([]byte(GetTotalSummary(states).Render()))
	for _, s := range states {
		if s != nil {
			t, err := s.ToTable()
			if err == nil {
				f.Write([]byte(t.Render()))
			}
		}
	}
}

// Pricing contains the details of core and memory
// pricing information to be outputted.
type Pricing struct {
	//if the cost of unit is unknown we use string "-"
	UnitCost  string  `json:"cost_per_unit"`
	NumUnits  int     `json:"number_of_units"`
	TotalCost float64 `json:"cost_of_units"`
}

// InstancePricingOut contains ComputeInstance pricing
// information to be outputted.
type InstancePricingOut struct {
	Cpu       Pricing `json:"cpu"`
	Memory    Pricing `json:"memory"`
	TotalCost float64 `json:"total_cost"`
}

// PricingInfoOut contains ComputeInstanceState pricing
// information to be outputted.
type PricingInfoOut struct {
	Before   *InstancePricingOut `json:"before"`
	After    *InstancePricingOut `json:"after"`
	DeltaCpu float64             `json:"cpu_cost_change"`
	DeltaMem float64             `json:"memory_cost_change"`
	Delta    float64             `json:"cost_change"`
}

// ComputeInstanceStateOut contains ComputeInstanceState
// information to be outputted.
type ComputeInstanceStateOut struct {
	Name        string         `json:"name"`
	Instance_ID string         `json:"instance_id"`
	Zone        string         `json:"zone"`
	MachineType string         `json:"machine_type"`
	CpuType     string         `json:"cpu_type"`
	MemType     string         `json:"memory_type"`
	Action      string         `json:"action"`
	Pricing     PricingInfoOut `json:"pricing_info"`
}

// JsonOutput contains relevant information resources
// and cost changes in a file.
type JsonOutput struct {
	Delta            float64                    `json:"cost_change"`
	DeltaMem         float64                    `json:"memory_cost_change"`
	DeltaCpu         float64                    `json:"cpu_cost_change"`
	ResourcesPricing []*ComputeInstanceStateOut `json:"resources_pricing_info"`
}

// ToStateOut extracts ComputeInstanceStateOut from state struct to render output in json
// format.
func (state *ComputeInstanceState) ToStateOut() (*ComputeInstanceStateOut, error) {
	before, after, err := syncInstances(state.Before, state.After)
	if err != nil {
		return nil, err
	}
	out := &ComputeInstanceStateOut{
		Name:        getComponent(before.Name, after.Name),
		Instance_ID: getComponent(before.ID, after.ID),
		Zone:        getComponent(before.Zone, after.Zone),
		MachineType: getComponent(before.MachineType, after.MachineType),
		CpuType:     getComponent(before.Cores.Type, after.Cores.Type),
		MemType:     getComponent(before.Memory.Type, after.Memory.Type),
	}

	dCore, dMem, err := state.getDelta()
	if err != nil {
		return nil, err
	}
	beforeOut, err := completeResourceOut(state.Before)
	if err != nil {
		return nil, err
	}
	afterOut, err := completeResourceOut(state.After)
	if err != nil {
		return nil, err
	}
	pricing := PricingInfoOut{
		Before:   beforeOut,
		After:    afterOut,
		DeltaCpu: dCore,
		DeltaMem: dMem,
		Delta:    dCore + dMem,
	}
	out.Pricing = pricing
	return out, nil
}

func completeResourceOut(r *ComputeInstance) (*InstancePricingOut, error) {
	core, mem, t, err := getMemCoreInfo(r)
	if err != nil {
		return nil, err
	}

	rOut := &InstancePricingOut{
		Cpu: Pricing{
			NumUnits:  core[1].(int),
			TotalCost: core[2].(float64),
		},
		Memory: Pricing{
			NumUnits:  mem[1].(int),
			TotalCost: mem[2].(float64),
		},
		TotalCost: t,
	}

	if core[0] == "-" {
		rOut.Cpu.UnitCost = "-"
	} else {
		rOut.Cpu.UnitCost = fmt.Sprintf("%f", core[0])
	}

	if mem[0] == "-" {
		rOut.Memory.UnitCost = "-"
	} else {
		rOut.Memory.UnitCost = fmt.Sprintf("%f", mem[0])
	}
	return rOut, nil
}

// RederJson returns ComputeInstanceState as a JSON string.
func (state *ComputeInstanceState) RederStateJson() (string, error) {
	out, err := state.ToStateOut()
	if err != nil || out == nil {
		return "", err
	}
	jsonString, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

// RenderJson returns the string with json output struct for all
// resources.
func RenderJson(states []*ComputeInstanceState) (string, error) {
	out := JsonOutput{}
	out.Delta, out.DeltaMem, out.DeltaCpu = GetTotalDelta(states)
	var r []*ComputeInstanceStateOut
	for _, state := range states {
		s, err := state.ToStateOut()
		if err == nil || s != nil {
			r = append(r, s)
		}
	}
	out.ResourcesPricing = r
	jsonString, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(jsonString), err
}

// GenerateJsonOut generates a json file with the pricing information of the specified resources.
func GenerateJsonOut(outputPath string, res []*ComputeInstanceState) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	jsonString, err := RenderJson(res)
	if err != nil {
		return nil
	}

	if _, err = io.WriteString(f, jsonString); err != nil {
		return err
	}

	return nil
}

// GetTotalDelta returns the cost changes of all resources.
func GetTotalDelta(states []*ComputeInstanceState) (dTotal, dCore, dMem float64) {
	for _, state := range states {
		core, mem, err := state.getDelta()
		if err == nil {
			dCore = core + dCore
			dMem = mem + dMem
		}
	}
	dTotal = dCore + dMem
	return dTotal, dCore, dMem
}

// getComponent returns the string to be outputted
// instead of before and after.
func getComponent(before, after string) string {
	var s string
	if before == "" && after == "" {
		s = "unknown"
	} else if before == "" || after == "" {
		s = before + after
	} else if before == after {
		s = before
	} else {
		s = before + after
	}
	return s
}

func syncInstances(before, after *ComputeInstance) (*ComputeInstance, *ComputeInstance, error) {
	if after == nil {
		after = before
	} else if before == nil {
		before = after
	}
	if after == nil && before == nil {
		return nil, nil, fmt.Errorf("After and Before can't be nil at the same time.")
	}
	return before, after, nil
}

// initRow creates a sufficient row for the cases when the certain field
// in state struct for before and after are the same or different.
// If end == true add " " in the end of string to avoid unwanted auto-merging
// in the table package.
func initRow(header, before, after string, end bool) (row table.Row) {
	row = append(row, header)
	s := getComponent(before, after)
	if before != "" && after != "" && before != after {
		s = before + " ->\n-> " + after
	}
	if end {
		s = s + " "
	}
	for i := 1; i < numColumns; i++ {
		row = append(row, s)
	}
	return row
}

// getMemCoreInfo returns two arrays with resource's core and memory
// information and the totalCost.
func getMemCoreInfo(r *ComputeInstance) (core, mem []interface{}, t float64, err error) {
	if r == nil {
		core = []interface{}{"-", 0, 0}
		mem = []interface{}{"-", 0, 0}
		return core, mem, 0, nil
	}
	core = append(core, r.Cores.UnitPricing.HourlyUnitPrice)
	core = append(core, r.Cores.Number)
	core = append(core, r.Cores.getTotalPrice())

	mem = append(mem, r.Memory.UnitPricing.HourlyUnitPrice)
	unitType := strings.Split(r.Memory.UnitPricing.UsageUnit, " ")[0]
	memNum, err := conv.Convert("gib", r.Memory.AmountGiB, unitType)
	if err != nil {
		return nil, nil, 0, err
	}
	mem = append(mem, memNum)
	p, err := r.Memory.getTotalPrice()
	if err != nil {
		return nil, nil, 0, err
	}
	mem = append(mem, p)
	return core, mem, r.Cores.getTotalPrice() + p, nil
}

// addHeader creates a row with the same string in every column.
func addHeader(header string) (row table.Row) {
	for i := 0; i < numColumns; i++ {
		row = append(row, header)
	}
	return row
}
