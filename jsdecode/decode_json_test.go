package jsdecode

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	resources "github.com/googleinterns/terraform-cost-estimation/resources"
	cd "github.com/googleinterns/terraform-cost-estimation/resources/classdetail"
	tfjson "github.com/hashicorp/terraform-json"
)

func TesttoComputeInstance(t *testing.T) {
	classDetails, err := cd.NewResourceDetail()
	if err != nil {
		t.Fatal(err.Error())
	}

	res1, _ := readResource("../testdata/compute_instances/resource1.json")
	res2, _ := readResource("../testdata/compute_instances/resource2.json")
	res3, _ := readResource("../testdata/compute_instances/resource3.json")
	res4, _ := readResource("../testdata/compute_instances/resource4.json")

	out1, _ := resources.NewComputeInstance(classDetails, "", "test", "n1-standard-1", "us-central1-a", "OnDemand")
	out2, _ := resources.NewComputeInstance(classDetails, "5889159656940809264", "test", "n1-standard-1", "us-central1-a", "Preemptible")
	out3, _ := resources.NewComputeInstance(classDetails, "", "test-us-east1-a-1", "n1-standard-1", "us-east1-a", "OnDemand")
	out4, _ := resources.NewComputeInstance(classDetails, "", "test-c2-standard-8", "c2-standard-8", "us-central1-a", "OnDemand")

	tests := []struct {
		in       interface{}
		expected *resources.ComputeInstance
	}{
		{
			res1,
			out1,
		},
		{
			res2,
			out2,
		},
		{
			res3,
			out3,
		},
		{
			res4,
			out4,
		},
	}

	for _, test := range tests {
		var actual *resources.ComputeInstance
		actual, err := toComputeInstance(classDetails, test.in)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", spew.Sdump(test.expected), spew.Sdump(actual))
		}
	}
}

func readResource(filePath string) (interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var res interface{}
	if err = json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func TesttoInstanceState(t *testing.T) {
	classDetails, err := cd.NewResourceDetail()
	if err != nil {
		t.Fatal(err.Error())
	}

	f, err := os.Open("../testdata/new-compute-instance/tfplan.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var plan *tfjson.Plan
	plan, err = ExtractPlanStruct(f)
	if err != nil || plan == nil {
		t.Fatal(err)
	}
	if plan.ResourceChanges == nil {
		t.Fatal(err)
	}

	after, _ := resources.NewComputeInstance(classDetails, "", "test", "n1-standard-1", "us-central1-a", "OnDemand")
	expected := &resources.ComputeInstanceState{
		Before: nil,
		After:  after,
		Action: "create",
	}

	var actual *resources.ComputeInstanceState
	actual, err = toInstanceState(classDetails, plan.ResourceChanges[0].Change)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", spew.Sdump(expected), spew.Sdump(actual))
	}

}

func TestGetResources(t *testing.T) {
	classDetails, err := cd.NewResourceDetail()
	if err != nil {
		t.Fatal(err.Error())
	}

	f, err := os.Open("../testdata/modified-compute-instance/tfplan.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var plan *tfjson.Plan
	plan, err = ExtractPlanStruct(f)
	if err != nil || plan == nil {
		t.Fatal(err)
	}

	before, _ := resources.NewComputeInstance(classDetails, "5889159656940809264", "test", "n1-standard-1", "us-central1-a", "OnDemand")
	after, _ := resources.NewComputeInstance(classDetails, "5889159656940809264", "test", "n1-standard-2", "us-central1-a", "OnDemand")
	expected := []resources.ResourceState{
		&resources.ComputeInstanceState{
			Before: before,
			After:  after,
			Action: "update",
		},
	}

	actual := GetResources(classDetails, plan)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", spew.Sdump(expected), spew.Sdump(actual))
	}
}
