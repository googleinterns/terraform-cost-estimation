package billing

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	billingpb "google.golang.org/genproto/googleapis/cloud/billing/v1"
)

func showCatalog(c *ComputeEngineCatalog) (s string) {
	s = "Cores:"
	for k, v := range c.coreInstances {
		s += "\n" + k + ": "
		for _, sku := range v {
			s += sku.Description + "; "
		}
	}

	s += "RAM:"
	for k, v := range c.ramInstances {
		s += "\n" + k + ": "
		for _, sku := range v {
			s += sku.Description + "; "
		}
	}
	return
}

func readSKU(path string) (*billingpb.Sku, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sku billingpb.Sku
	if err = jsonpb.UnmarshalString(string(data), &sku); err != nil {
		return nil, err
	}

	return &sku, nil
}

func readSKUs() ([]*billingpb.Sku, error) {
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/testdata/sku_%d.json"

	var skus []*billingpb.Sku
	for i := 0; i <= 10; i++ {
		sku, err := readSKU(fmt.Sprintf(inputPath, i))
		if err != nil {
			return nil, err
		}
		skus = append(skus, sku)
	}
	return skus, nil
}

func TestAssignSKUCategories(t *testing.T) {
	skus, err := readSKUs()
	if err != nil {
		t.Fatal("Failed to read SKU JSON files")
	}

	c1 := emptyComputeEngineCatalog()
	c2 := emptyComputeEngineCatalog()
	c3 := emptyComputeEngineCatalog()
	c4 := emptyComputeEngineCatalog()

	c2.coreInstances["Preemptible"] = []*billingpb.Sku{skus[10]}
	c2.ramInstances["OnDemand"] = []*billingpb.Sku{skus[0], skus[9]}

	c3.coreInstances["Commit1Yr"] = []*billingpb.Sku{skus[4]}
	c3.coreInstances["Preemptible"] = []*billingpb.Sku{skus[8]}
	c3.ramInstances["Preemptible"] = []*billingpb.Sku{skus[1]}
	c3.ramInstances["OnDemand"] = []*billingpb.Sku{skus[5]}

	c4.coreInstances["Commit1Yr"] = []*billingpb.Sku{skus[4]}
	c4.coreInstances["Preemptible"] = []*billingpb.Sku{skus[8], skus[10]}
	c4.ramInstances["OnDemand"] = []*billingpb.Sku{skus[0], skus[5], skus[9]}
	c4.ramInstances["Preemptible"] = []*billingpb.Sku{skus[1]}

	tests := []struct {
		name    string
		skus    []*billingpb.Sku
		catalog *ComputeEngineCatalog
	}{
		{"no_cpu_no_ram", []*billingpb.Sku{skus[2], skus[3], skus[6], skus[7]}, c1},
		{"different_n1standard", []*billingpb.Sku{skus[0], skus[9], skus[10]}, c2},
		{"different_usage_type", []*billingpb.Sku{skus[1], skus[4], skus[5], skus[8]}, c3},
		{"all_skus", []*billingpb.Sku{skus[0], skus[1], skus[2], skus[3], skus[4],
			skus[5], skus[6], skus[7], skus[8], skus[9], skus[10]}, c4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := emptyComputeEngineCatalog()
			catalog.assignSKUCategories(test.skus)

			if !reflect.DeepEqual(*catalog, *test.catalog) {
				t.Errorf("test: catalog.assignSKUCategories(skus) -> \n%s,;\n want\n %s",
					showCatalog(catalog), showCatalog(test.catalog))
			}
		})
	}
}
