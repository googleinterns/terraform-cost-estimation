package jsdecode

import (
	"encoding/json"
	resources "github.com/googleinterns/terraform-cost-estimation/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"io"
	"io/ioutil"
)

// ComputeInstanceType is the type of ResourceChange and Resource supported
// by this package, we consider ComputeInstances only for now.
const ComputeInstanceType = "google_compute_instance"

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
func ExtractPlanStruct(reader io.Reader) (*tfjson.Plan, error) {
	bytes, _ := ioutil.ReadAll(reader)

	var plan tfjson.Plan
	if err := plan.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}
	return &plan, nil
}

// ToComputeInstance extracts ComputeInstance from the interface that
// contains information about resource.
func ToComputeInstance(resource interface{}) (*resources.ComputeInstance, error) {
	if resource == nil {
		return nil, nil
	}

	jsonString, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	var r *ResourceInfo
	if err := json.Unmarshal(jsonString, &r); err != nil || r == nil {
		return nil, err
	}

	// Region of ComputeInstance is determined as <region>-<zone>, in case if r.Zone
	// is empty we leave region empty and validate it in another function.
	var region string
	if len(r.Zone) >= 2 && r.Zone[len(r.Zone)-1] == '-' {
		if r.Zone[len(r.Zone)-2] >= 'a' && r.Zone[len(r.Zone)-2] <= 'f' {
			region = r.Zone[:len(r.Zone)-2]
		}
	} else {
		region = r.Zone
	}

	return &resources.ComputeInstance{
		// TODO add core/memory for some types of ComputeInstances
		ID:          r.ID,
		Name:        r.Name,
		MachineType: r.MachineType,
		Region:      region,
	}, nil
}

// GetChange returns the pointer to the struct with states of the
// certain resource of ComputeInstance type.
func GetChange(change *tfjson.Change) (*ResourceStates, error) {
	before, err := ToComputeInstance(change.Before)
	if err != nil {
		return nil, err
	}
	after, err := ToComputeInstance(change.After)
	if err != nil {
		return nil, err
	}
	resource := &ResourceStates{
		Before: before,
		After:  after,
	}
	if resource.Before == nil && resource.After == nil {
		return nil, nil
	}
	return resource, nil
}

// GetResources extracts all resources of ComputeInstance type and their
// before and after states from plan file.
func GetResources(plan *tfjson.Plan) []*ResourceStates {
	var resources []*ResourceStates
	for _, resourceChange := range plan.ResourceChanges {
		if ComputeInstanceType == resourceChange.Type {
			if resource, err := GetChange(resourceChange.Change); resource != nil && err == nil {
				resources = append(resources, resource)
			}
		}
	}
	return resources
}
