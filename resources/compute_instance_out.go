package resources

import (
	"encoding/json"
	"fmt"
	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"strings"
)

// The number of columns in output table for ComputeInstance.
const numColumns = 5

// ToTable creates a table.Table and fills it with the pricing information from
// ComputeInstanceState.
func (state *ComputeInstanceState) ToTable() (*table.Table, error) {
	if state == nil {
		return nil, nil
	}
	t := &table.Table{}
	autoMerge := table.RowConfig{AutoMerge: true}

	before := state.Before
	after := state.After
	if after == nil {
		after = before
	} else if before == nil {
		before = after
	}
	if after == nil && before == nil {
		return nil, fmt.Errorf("After and Before can't be nil at the same time.")
	}
	t.AppendRow(initRow("Name", before.Name, after.Name, false), autoMerge)
	t.AppendRow(initRow("Instance_ID", before.ID, after.ID, true), autoMerge)
	t.AppendRow(initRow("Zone", before.Zone, after.Zone, false), autoMerge)
	t.AppendRow(initRow("Machine type", before.MachineType, after.MachineType, true), autoMerge)
	//t.AppendRow(initRow("Core type", before.Core.Type, after.Core.Type, false), autoMerge)
	//t.AppendRow(initRow("Memory type", before.Mem.Type, after.Mem.Type, true), autoMerge)

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
	// Add " " in the end of string to avoid unwanted auto-merging in the
	// table package.
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

/*
// GetTotalSummary returns the table with brief cost changes info about all
// resources.
func GetTotalSummary(states []*ComputeInstanceState) (*table.Table) {
	t := &table.Table{}
	dTotal, dCore, dMem := GetTotalDelta(states)
	//In header total cost change for all resources.
	t.AppendHeader("")

	t.AppendRow(addHeader("Pricing Information"))
	t.AppendRow(table.Row{"Name", "MachineType", "Action", "Change", "Delta"})
	for _, r := range states {
		//in GetSummary return array {Name , MachineType, Action, Change, dCore+dMem}
		t.AppendRow(table.Row{r.GetSummary()})
	}
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	return t
}
*/

// Pricing contains the details of core and memory
// pricing information to be outputted.
type Pricing struct {
	UnitCost  float64 `json:"cost_per_unit"`
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
	Before   InstancePricingOut `json:"before"`
	After    InstancePricingOut `json:"after"`
	DeltaCpu float64            `json:"cpu_cost_change"`
	DeltaMem float64            `json:"memory_cost_change"`
	Delta    float64            `json:"cost_change"`
}

// ComputeInstanceStateOut contains ComputeInstanceState
// information to be outputted.
type ComputeInstanceStateOut struct {
	Name        string         `json:"name"`
	Instance_ID string         `json:"instance_id"`
	Zone        string         `json:"zone"`
	MachineType string         `json:"machine_type"`
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
	if state == nil {
		return nil, nil
	}

	before := state.Before
	after := state.After
	if after == nil {
		after = before
	} else if before == nil {
		before = after
	}
	if after == nil && before == nil {
		return nil, fmt.Errorf("After and Before can't be nil at the same time.")
	}
	out := &ComputeInstanceStateOut{
		Name:        getComponent(before.Name, after.Name),
		Instance_ID: getComponent(before.ID, after.ID),
		Zone:        getComponent(before.Zone, after.Zone),
		MachineType: getComponent(before.MachineType, after.MachineType),
	}

	core1, mem1, t1, err := getMemCoreInfo(state.Before)
	if err != nil {
		return nil, err
	}
	core2, mem2, t2, err := getMemCoreInfo(state.After)
	if err != nil {
		return nil, err
	}
	dCore, dMem, err := state.getDelta()
	if err != nil {
		return nil, err
	}
	//TODO add for cses when "-"
	beforeOut := InstancePricingOut{
		Cpu: Pricing{
			UnitCost:  core1[0].(float64),
			NumUnits:  core1[1].(int),
			TotalCost: core1[2].(float64),
		},
		Memory: Pricing{
			UnitCost:  mem1[0].(float64),
			NumUnits:  mem1[1].(int),
			TotalCost: mem1[2].(float64),
		},
		TotalCost: t1,
	}
	afterOut := InstancePricingOut{
		Cpu: Pricing{
			UnitCost:  core2[0].(float64),
			NumUnits:  core2[1].(int),
			TotalCost: core2[2].(float64),
		},
		Memory: Pricing{
			UnitCost:  mem2[0].(float64),
			NumUnits:  mem2[1].(int),
			TotalCost: mem2[2].(float64),
		},
		TotalCost: t2,
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
	jsonString, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(jsonString), err
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
