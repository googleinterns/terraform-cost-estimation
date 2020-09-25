package resources

import (
	"fmt"
	conv "github.com/googleinterns/terraform-cost-estimation/memconverter"
	"reflect"
	"sort"
	"strings"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
	"github.com/googleinterns/terraform-cost-estimation/io/js"
	"github.com/jedib0t/go-pretty/v6/table"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

const (
	nano            = float64(1000 * 1000 * 1000)
	epsilon         = 1e-10
	hourlyToMonthly = float64(24 * 30)
	hourlyToYearly  = float64(24 * 365)
)

func generalChange(initial, final string) string {
	if initial != final {
		return initial + " -> " + final
	}
	return initial
}

func zonesChange(z1, z2 []string) string {
	sort.Strings(z1)
	sort.Strings(z2)
	if reflect.DeepEqual(z1, z2) {
		return strings.Join(z1, ", ")
	}
	return strings.Join(z1, ", ") + " -> " + strings.Join(z2, ", ")
}

func filterSKUs(skus []*billingpb.Sku, region string, d Description) ([]*billingpb.Sku, error) {
	filtered, err := billing.RegionFilter(skus, region)
	if err != nil {
		return nil, err
	}

	filtered, err = billing.DescriptionFilter(filtered, d.Contains, d.Omits)
	if err != nil {
		return nil, err
	}
	return filtered, nil
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

func completeInstanceOut(r *ComputeInstance) (*js.InstancePricing, error) {
	core, mem, t, err := getMemCoreInfo(r)
	if err != nil {
		return nil, err
	}

	rOut := &js.InstancePricing{
		Cpu: js.Pricing{
			UnitCost:  core[0],
			NumUnits:  core[1],
			TotalCost: core[2],
		},
		Ram: js.Pricing{
			UnitCost:  mem[0],
			NumUnits:  mem[1],
			TotalCost: mem[2],
		},
		TotalCost: t,
	}
	return rOut, nil
}
