package jsdecode

import (
	"encoding/json"
	resources "github.com/googleinterns/terraform-cost-estimation/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"io/ioutil"
	"os"
)

// ComputeIstanceType is the type of ResourceChange and Resource supported
// by this package, we consider ComputeInstances only for now.
const ComputeIstanceType = "google_compute_instance"

// TODO maybe add mem/core info
// ResourceInfo contains the information about Resource, this struct is
// used to cast interfaces with before/after states of the certain Resource.
type ResourceInfo struct {
	ID          string `json:"id,omitempty"`
	InstanceID  string `json:"instance_id"`
	Name        string `json:"name,omitempty"`
	MachineType string `json:"machine_type,omitempty"`
	Zone        string `json:"zone,omitempty"`
}

// ResourceStates contains information about the states of resource
// of type ComputeInstance in plan configuration.
type ResourceStates struct {
	Before *resources.ComputeInstance
	After  *resources.ComputeInstance
}

// ExtractPlanStruct extracts tfjson.Plan struct from file in provided path
// and returns the pointer on it if it is possible, otherwise returns error.
func ExtractPlanStruct(filePath string) (*tfjson.Plan, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byteFile, _ := ioutil.ReadAll(f)

	var plan tfjson.Plan
	if err = plan.UnmarshalJSON(byteFile); err != nil {
		return nil, err
	}
	return &plan, nil
}

// toComputeInstance extracts ComputeInstance from the interface that
// contains information about resource.
func toComputeInstance(resource interface{}) *resources.ComputeInstance {
	if resource == nil {
		return nil
	}

	jsonString, _ := json.Marshal(resource)
	var r ResourceInfo
	json.Unmarshal(jsonString, &r)
	return &resources.ComputeInstance{
        // TODO check if unmarshaled resource is valid and zone>2 chars
        // add core/memory for some types of ComputeInstances
		ID:          r.ID,
		Name:        r.Name,
		MachineType: r.MachineType,
		Region:      r.Zone[:len(r.Zone)-2],
	}
}

// GetChange returns the pointer to the struct with states of the
// certain resource of ComputeInstance type.
func GetChange(change *tfjson.Change) *ResourceStates {
	resource := &ResourceStates{
		Before: toComputeInstance(change.Before),
		After:  toComputeInstance(change.After),
	}
	if resource.Before == nil && resource.After == nil {
		return nil
	}
	return resource
}

// GetConfiguration extracts all resources of ComputeInstance type and their
// before and after states from plan file.
func GetConfiguration(plan *tfjson.Plan) []*ResourceStates {
	var resources []*ResourceStates
	for _, resourceChange := range plan.ResourceChanges {
		if ComputeIstanceType == resourceChange.Type {
			if resource := GetChange(resourceChange.Change); resource != nil {
				resources = append(resources, resource)
			}
		}
	}
	return resources
}
