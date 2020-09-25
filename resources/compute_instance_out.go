package resources

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// ToTable creates a table.Table and fills it with the pricing information from ComputeInstanceState.
func (state *ComputeInstanceState) ToTable(colorful bool) (*table.Table, error) {
	before, after, err := syncInstances(state.Before, state.After)
	if err != nil {
		return nil, err
	}

	t := &table.Table{}
	autoMerge := table.RowConfig{AutoMerge: true}
	t.AppendRow(initRow("Name", before.Name, after.Name, false), autoMerge)
	t.AppendRow(initRow("ID", before.ID, after.ID, true), autoMerge)
	t.AppendRow(initRow("Zone", before.Zone, after.Zone, false), autoMerge)
	t.AppendRow(initRow("Machine type", before.MachineType, after.MachineType, true), autoMerge)
	t.AppendRow(initRow("Action", state.Action, state.Action, false), autoMerge)
	h := "Pricing Information\n(USD/h)"
	t.AppendRow(table.Row{h, h, h, h, h}, autoMerge)
	core1, mem1, t1, err := getMemCoreInfo(state.Before)
	if err != nil {
		return nil, err
	}
	core2, mem2, t2, err := getMemCoreInfo(state.After)
	if err != nil {
		return nil, err
	}
	t1Str := fmt.Sprintf("%.6f", t1)
	// Add " " in the end of string to avoid unwanted auto-merging in the table package.
	t2Str := fmt.Sprintf("%.6f ", t2)
	t.AppendRow(table.Row{" ", " ", "CPU", "RAM", "Total"}, autoMerge)
	t.AppendRows([]table.Row{
		{"Before", "Cost\nper\nunit", core1[0], mem1[0], t1Str},
		{"Before", "Number\nof\nunits", core1[1] + " ", mem1[1] + " ", t1Str},
		{"Before", "Units\ncost", core1[2], mem1[2], t1Str},
		{"After", "Cost\nper\nunit", core2[0] + " ", mem2[0] + " ", t2Str},
		{"After", "Number\nof\nunits", core2[1], mem2[1], t2Str},
		{"After", "Units\ncost", core2[2] + " ", mem2[2] + " ", t2Str},
	})

	dCore, dMem := state.getDeltas()

	color := text.FgGreen
	change := "No change"
	if dTotal := dCore + dMem; dTotal < 0 {
		change = "Down (↓)"
		color = text.FgRed
	} else if dTotal > 0 {
		change = "Up (↑)"
	}
	if colorful {
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 5, AutoMerge: true, ColorsFooter: text.Colors{color}},
		})
	} else {
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 5, AutoMerge: true},
		})
	}
	t.AppendFooter(table.Row{"DELTA", change, dCore, dMem, dCore + dMem})
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	return t, nil
}

// GetSummaryTable returns the table with brief cost changes info about all Compute Instance resources.
func GetSummaryTable(states []ResourceState) *table.Table {
	t := &table.Table{}
	autoMerge := table.RowConfig{AutoMerge: true}

	dTotal := getTotalDelta(states)
	t.SetTitle(fmt.Sprintf("The total cost change for all Compute Instances is %.6f USD/hour.", dTotal))
	h := "Pricing Information\n(USD/h)"
	t.AppendRow(table.Row{h, h, h, h, h}, autoMerge)
	t.AppendRow(table.Row{"Instance name", "Instance ID", "Machine type", "Action", "Delta"})
	for _, s := range states {
		if row, err := s.getSummaryRow(); err == nil {
			t.AppendRow(row)
		} else {
			log.Printf("Error: %v", err)
		}
	}
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	return t
}

// OutputPricing writes pricing information about each resource and summary.
func OutputPricing(states []ResourceState, f *os.File) {
	var colorful bool
	if f == os.Stdout {
		colorful = true
	}
	f.Write([]byte(GetSummaryTable(states).Render() + "\n\n"))
	f.Write([]byte("\n List of all Resources:\n\n"))
	for _, s := range states {
		if s != nil {
			t, err := s.ToTable(colorful)
			if err == nil {
				f.Write([]byte(t.Render() + "\n\n\n"))
			} else {
				log.Printf("Error: %v", err)
			}
		}
	}
}

// getTotalDelta returns the cost change of all Compute Instance resources.
func getTotalDelta(states []ResourceState) float64 {
	var t float64
	for _, s := range states {
		t += s.getDelta()
	}
	return t
}

// getMemCoreInfo returns two arrays with resource's core and memory information and the totalCost.
func getMemCoreInfo(r *ComputeInstance) (core, mem []string, t float64, err error) {
	if r == nil {
		return []string{"-", "0", "0"}, []string{"-", "0", "0"}, 0, nil
	}

	core = append(core, fmt.Sprintf("%.6f", float64(r.Cores.UnitPricing.HourlyUnitPrice)))
	core = append(core, fmt.Sprintf("%d", r.Cores.Number))
	core = append(core, fmt.Sprintf("%.6f", float64(r.Cores.getTotalPrice())))

	mem = append(mem, fmt.Sprintf("%.6f", float64(r.Memory.UnitPricing.HourlyUnitPrice)))
	unitType := strings.Split(r.Memory.UnitPricing.UsageUnit, " ")[0]
	memNum, err := conv.Convert("gib", r.Memory.AmountGiB, unitType)
	if err != nil {
		return nil, nil, 0, err
	}
	mem = append(mem, fmt.Sprintf("%.2f", memNum))
	p := r.Memory.getTotalPrice()
	mem = append(mem, fmt.Sprintf("%.6f", p))
	return core, mem, r.Cores.getTotalPrice() + p, nil
}

// syncInstances replace nils in state's before and after to be able to use them.
func syncInstances(before, after *ComputeInstance) (*ComputeInstance, *ComputeInstance, error) {
	if after == nil && before == nil {
		return nil, nil, fmt.Errorf("After and Before can't be nil at the same time.")
	}
	if after == nil {
		return before, before, nil
	}
	if before == nil {
		return after, after, nil
	}
	return before, after, nil
}

// initRow creates a sufficient row for the certain field in state struct depending on before and after are the same or different.
// If end == true add " " in the end of string to avoid unwanted auto-merging in the table package.
func initRow(h, before, after string, end bool) (row table.Row) {
	var s string
	switch {
	case before == "" && after == "":
		s = "unknown"
	case before == "" || after == "":
		s = before + after
	case before == after:
		s = before
	default:
		s = before + " ->\n-> " + after
	}
	if end {
		s = s + " "
	}

	row = append(row, h)
	for i := 1; i < 5; i++ {
		row = append(row, s)
	}
	return row
}

// getSummaryRow() returns the row for SummaryTable to be outputted about the certain state.
func (state *ComputeInstanceState) getSummaryRow() (table.Row, error) {
	dCore, dMem := state.getDeltas()
	_, r, err := syncInstances(state.Before, state.After)
	if err != nil {
		return table.Row{}, err
	}
	return table.Row{r.Name, r.ID, r.MachineType, state.Action, fmt.Sprintf("%.6f", dCore+dMem)}, nil
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
	Instance_ID Change               `json:"instance_id"`
	Zone        Change               `json:"zone"`
	MachineType Change               `json:"machine_type"`
	CpuType     Change               `json:"cpu_type"`
	RamType     Change               `json:"ram_type"`
	Action      string               `json:"action"`
	Pricing     InstanceStatePricing `json:"pricing_info"`
}

func (out *ComputeInstanceStateOut) addToJSONTableList(json *JsonOutput) {
	json.ComputeInstancesPricing = append(json.ComputeInstancesPricing, out)
}

// ComputeDiskStateOut contains ComputeDiskState information to be outputted.
type ComputeDiskStateOut struct {
	Name        Change           `json:"name"`
	Instance_ID Change           `json:"instance_id"`
	Zones       Change           `json:"zones"`
	DiskType    Change           `json:"disk_type"`
	Action      string           `json:"action"`
	Pricing     DiskStatePricing `json:"pricing_info"`
}

func (out *ComputeDiskStateOut) addToJSONTableList(json *JsonOutput) {
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

// ToStateOut creates ComputeInstanceStateOut from state struct to render output in json format.
func (state *ComputeInstanceState) ToStateOut() (JSONOut, error) {
	before, after, err := syncInstances(state.Before, state.After)
	if err != nil {
		return nil, err
	}
	out := &ComputeInstanceStateOut{
		Name:        Change{before.Name, after.Name},
		Instance_ID: Change{before.ID, after.ID},
		Zone:        Change{before.Zone, after.Zone},
		MachineType: Change{before.MachineType, after.MachineType},
		CpuType:     Change{before.Cores.Type, after.Cores.Type},
		RamType:     Change{before.Memory.Type, after.Memory.Type},
		Action:      state.Action,
	}

	dCore, dMem := state.getDeltas()
	beforeOut, err := completeResourceOut(state.Before)
	if err != nil {
		return nil, err
	}
	afterOut, err := completeResourceOut(state.After)
	if err != nil {
		return nil, err
	}
	pricing := InstanceStatePricing{
		Before:   beforeOut,
		After:    afterOut,
		DeltaCpu: dCore,
		DeltaRam: dMem,
		Delta:    dCore + dMem,
	}
	out.Pricing = pricing
	return out, nil
}

// RenderJson returns the string with json output struct for all resources.
func RenderJson(states []ResourceState) (string, error) {
	out := JsonOutput{}
	out.Delta = getTotalDelta(states)
	out.PricingUnit = "USD/hour"
	for _, state := range states {
		s, err := state.ToStateOut()
		if err == nil || s != nil {
			s.addToJSONTableList(&out)
		}
	}
	jsonString, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(jsonString), err
}

// GenerateJsonOut generates a json file with the pricing information of the specified resources.
func GenerateJsonOut(f *os.File, res []ResourceState) error {
	jsonString, err := RenderJson(res)
	if err != nil {
		return nil
	}
	if _, err = io.WriteString(f, jsonString); err != nil {
		return err
	}
	return nil
}

func completeResourceOut(r *ComputeInstance) (*InstancePricing, error) {
	core, mem, t, err := getMemCoreInfo(r)
	if err != nil {
		return nil, err
	}

	rOut := &InstancePricing{
		Cpu: Pricing{
			UnitCost:  core[0],
			NumUnits:  core[1],
			TotalCost: core[2],
		},
		Ram: Pricing{
			UnitCost:  mem[0],
			NumUnits:  mem[1],
			TotalCost: mem[2],
		},
		TotalCost: t,
	}
	return rOut, nil
}
