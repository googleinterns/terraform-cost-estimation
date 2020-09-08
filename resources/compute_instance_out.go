package resources

import (
	"encoding/json"
	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
	table "github.com/jedib0t/go-pretty/table"
	//text "github.com/jedib0t/go-pretty/text"
	"fmt"
	"os"
	"strings"
)

//------------------Render console/plain text output-------------------

// The number of columns in the output table for ComputeInstance.
const (
	numColumns = 5
)

// addRow creates a sufficient row for the cases when the certain field
// in state struct for before and after are the same or different.
func addRow(header, before, after string) (row table.Row) {
	row = append(row, header)
	var s string
	if before == "" && after == "" {
		s = "unknown"
	} else if before == "" || after == "" {
		s = before + after
	} else if before == after {
		s = before
	} else {
		s = before + " ->\n-> " + after
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
		return core, mem, 0, err
	}
	mem = append(mem, memNum)
	mem = append(mem, r.Memory.getTotalPrice())

	t = r.Cores.getTotalPrice() + r.Memory.getTotalPrice()
	return core, mem, t, nil
}

// addHeader cretes a row with the same string in every column.
func addHeader(header string) (row table.Row) {
	for i := 0; i < numColumns; i++ {
		row = append(row, header)
	}
	return row
}

// StateToTable creates a table.Table and fills it with the pricing information from
// ComputeInstanceState.
func (state *ComputeInstanceState) StateToTable() (table.Table, error) {
	if state == nil {
		return table.Table{}, nil
	}
	t := table.Table{}
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

	if state.Before != nil && state.After != nil {
		t.AppendRow(addRow("Name", state.Before.Name, state.After.Name), rowConfigAutoMerge)
		t.AppendRow(addRow("Instance_ID", state.Before.ID, state.After.ID), rowConfigAutoMerge)
		t.AppendRow(addRow("Zone", state.Before.Zone, state.After.Zone), rowConfigAutoMerge)
		t.AppendRow(addRow("Machine type", state.Before.MachineType, state.After.MachineType))
	} else if state.Before != nil {
		t.AppendRow(addRow("Name", state.Before.Name, state.Before.Name), rowConfigAutoMerge)
		t.AppendRow(addRow("Instance_ID", state.Before.ID, state.Before.ID), rowConfigAutoMerge)
		t.AppendRow(addRow("Zone", state.Before.Zone, state.Before.Zone), rowConfigAutoMerge)
		t.AppendRow(addRow("Machine type", state.Before.MachineType, state.Before.MachineType), rowConfigAutoMerge)
	} else if state.After != nil {
		t.AppendRow(addRow("Name", state.After.Name, state.After.Name), rowConfigAutoMerge)
		t.AppendRow(addRow("Instance_ID", state.After.ID, state.After.ID), rowConfigAutoMerge)
		t.AppendRow(addRow("Zone", state.After.Zone, state.After.Zone), rowConfigAutoMerge)
		t.AppendRow(addRow("Machine type", state.After.MachineType, state.After.MachineType), rowConfigAutoMerge)
	}
	t.AppendRow(addRow("Action", state.Action, state.Action), rowConfigAutoMerge)
	t.AppendRow(addHeader("Pricing\nInformation"), rowConfigAutoMerge)

	core1, mem1, t1, err := getMemCoreInfo(state.Before)
	if err != nil {
		return t, err
	}
	core2, mem2, t2, err := getMemCoreInfo(state.After)
	if err != nil {
		return t, err
	}
	dCore, dMem, err := state.getDelta()
	if err != nil {
		return t, err
	}
	t.AppendRow(table.Row{" ", " ", "Cpu", "Mem", "Total"}, rowConfigAutoMerge)
	t.AppendRows([]table.Row{
		{"Before", "Cost\nper\nunit", core1[0], mem1[0], t1},
		{"Before", "Number\nof\nunits", core1[1], mem1[1], t1},
		{"Before", "Units\ncost", core1[2], mem1[2], t1},
		{"After", "Cost\nper\nunit", core2[0], mem2[0], t2},
		{"After", "Number\nof\nunits", core2[1], mem2[1], t2},
		{"After", "Units\ncost", core2[2], mem2[2], t2},
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		//TODO merge Total in Before and After
	})

	dCoreStr := fmt.Sprintf("%f", dCore)
	dMemStr := fmt.Sprintf("%f", dMem)
	dTotalStr := fmt.Sprintf("%f", dCore+dMem)
	change := "No change"
	if dCore+dMem < 0 {
		change = "Down"
	} else if dCore+dMem > 0 {
		change = "Up"
	}
	t.AppendRow(table.Row{"DELTA", change, dCoreStr, dMemStr, dTotalStr})
	SetOutputStyle(&t)
	return t, nil
}

// SetOutputStyle adds to table a style formatting.
func SetOutputStyle(t *table.Table) {
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
}

// RenderPricingTable prints Pricing Table of
// ComputeInstanceState.
func RenderPricingTable(r *ComputeInstanceState) {
	t, _ := r.StateToTable()
	t.SetOutputMirror(os.Stdout)
	//add to file
	t.Render()
}

//--------------Render JSON output----------------

// Pricing contains the details about core and memory
// pricing information to be outputed.
type Pricing struct {
	UnitCost  float64 `json:"cost_per_unit"`
	NumUnits  int     `json:"number_of_units"`
	TotalCost float64 `json:"cost_of_units"`
}

// InstancePricingOut contains ComputeInstance pricing
// information to be outputed.
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
	DeltaCpu float64            `json:"cpu_delta"`
	DeltaMem float64            `json:"memory_delta"`
	Delta    float64            `json:"delta"`
}

// ComputeInstanceStateOut contains ComputeInstanceState
// information to be outputted.
type ComputeInstanceStateOut struct {
	Name        string         `json:"name"`
	Instance_ID string         `json:"instance_id"`
	Zone        string         `json:"zone"`
	MachineType string         `json:"machine_type"`
	Action      string         `json:"action"`
	Pricing     PricingInfoOut `json:"pricing_information"`
}

// getField compares the field in Before and After and returns the string to
// be added to ComputeInstanceStateOut.
func getField(before, after string) string {
	if before == "" && after == "" {
		return "unknown"
	} else if before == "" || after == "" {
		return before + after
	} else if before == after {
		return before
	} else {
		return before + " -> " + after
	}
}

//TODO complete the function
func StateOut(state *ComputeInstanceState) (*ComputeInstanceStateOut, error) {
	if state == nil {
		return nil, nil
	}

	out := new(ComputeInstanceStateOut)
	// modify the filling depending on action
	return out, nil
}

// RederJson returns ComputeInstanceState as a JSON string.
func (state *ComputeInstanceState) RederJson() (string, error) {
	out, err := StateOut(state)
	if err != nil || out == nil {
		return "", err
	}
	jsonString, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

// Looks ugly in this lib. TODO use the html template.
// RenderHTMLTable returns pricing information in HTML table format as a string.
func (state *ComputeInstanceState) RenderHTMLTable() (string, error) {
	table, err := state.StateToTable()
	if err != nil {
		return "", err
	}
	return table.RenderHTML(), nil
}
