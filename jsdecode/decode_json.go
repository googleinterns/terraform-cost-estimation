package jsdecode

import (
	"encoding/json"
	"errors"
	"fmt"
	resources "github.com/googleinterns/terraform-cost-estimation/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"io/ioutil"
	"os"
)

// ComputeIstanceType is the type of ResourceChange and Resource supported
// by this package, we consider ComputeInstances only for now.
const ComputeIstanceType = "google_compute_instance"

// PlanFormatVersion is the supported version of the JSON plan format.
const PlanFormatVersion = "0.1"

// ResourceInfo conatains the information about Resource, this struct is
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

	var plan *tfjson.Plan
	if plan, err := UnmarshalJSON(byteFile); err != nil {
		return nil, err
	}
	return plan, nil
}

// TODO!
// complete Unmarshalling function with ResearchInfo in plan.ResourceChanges
// for every ResourceChange in Change.before/after  instead of interface{}
func UnmarshalJSON(b []byte) (*tfjson.Plan, error) {
	var p tfjson.Plan

	type rawPlan tfjson.Plan
	var plan rawPlan
	if err := json.Unmarshal(b, &plan); err != nil {
		return nil, err
	}

	// case p != nill ...
	*p = *(*tfjson.Plan)(&plan)
	return &p, Validate(&p)
}

// Validate checks to ensure that the plan is present, and the
// version matches PlanFormatVersion.
func Validate(plan *tfjson.Plan) error {
	if plan == nil {
		return errors.New("plan is nil")
	}

	if plan.FormatVersion == "" {
		return errors.New("unexpected plan input, it has to contain a format version field")
	}

	if PlanFormatVersion != plan.FormatVersion {
		return fmt.Errorf("plan format version not supported: expected %q, got %q", PlanFormatVersion, plan.FormatVersion)
	}

	return nil
}

// ExtractResource extructs ComputeInstance from the struct that
// contains information about resource.
func ExtractResource(resource *ResourceInfo) *resources.ComputeInstance {
	if resource == nil {
		return nil
	}
	return &resources.ComputeInstance{
		ID:          resource.ID,
		Name:        resource.Name,
		MachineType: resource.MachineType,
		//TODO consider case where zone empty if it is possible in valid json
		Region:      resource.Zone[:len(resource.Zone)-2],
	}
}

// GetChange returns the pointer to the struct with states of the
// certain resource of ComputeInstance type.
func GetChange(change *tfjson.Change) *ResourceStates {
	resource := &ResourceStates{
		Before: ExtractResource(change.Before),
		After:  ExtractResource(change.After),
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
