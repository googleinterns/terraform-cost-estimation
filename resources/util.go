package resources

import (
	"reflect"
	"sort"
	"strings"

	billing "github.com/googleinterns/terraform-cost-estimation/billing"
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
