package jsdecode

import (
	"encoding/json"
	resources "github.com/googleinterns/terraform-cost-estimation/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"io"
	"io/ioutil"
	"log"
)

// ComputeInstanceType is the type of ResourceChange and Resource supported
// by this package, we consider ComputeInstances only for now.
const ComputeInstanceType = "google_compute_instance"

// Possible Actions in ComputeInstanceState.
const (
	ActionNoop    string = "no-op"
	ActionCreate  string = "create"
	ActionDelete  string = "delete"
	ActionUpdate  string = "update"
	ActionReplace string = "replace"
)

// ResourceInfo contains the information about Resource, this struct is
// used to cast interfaces with before/after states of the certain Resource.
type ResourceInfo struct {
	ID          string      `json:"id,omitempty"`
	InstanceID  string      `json:"instance_id"`
	Name        string      `json:"name,omitempty"`
	MachineType string      `json:"machine_type,omitempty"`
	Zone        string      `json:"zone,omitempty"`
	Scheduling  []UsageType `json:"scheduling,omitempty"`
}

// UsageType contains the information whether the certain istance
// is preemptible or not.
type UsageType struct {
	IsPreemptible bool `json:"preemptible,omitempty"`
}

// ExtractPlanStruct extracts tfjson.Plan struct from file in provided path
// and returns the pointer on it if it is possible, otherwise returns error.
func ExtractPlanStruct(reader io.Reader) (*tfjson.Plan, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var plan tfjson.Plan
	err = plan.UnmarshalJSON(bytes)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// toComputeInstance extracts ComputeInstance from the interface that
// contains information about the resource.
func toComputeInstance(resource interface{}) (*resources.ComputeInstance, error) {
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

	usageType := "OnDemand"
	if len(r.Scheduling) >= 1 && r.Scheduling[0].IsPreemptible {
		usageType = "Preemptible"
	}
	instance, err := resources.NewComputeInstance(r.InstanceID, r.Name, r.MachineType, r.Zone, usageType)
	if err != nil {
		return instance, err
	}
	instance.Name = r.Name
	return instance, nil
	// return resources.NewComputeInstance(r.InstanceID, r.Name, r.MachineType, r.Zone, usageType)
}

// TODO pass the function f, for ComputeInstance it will be f = toComputeInstance():
// func GetChange(change *tfjson.Change, f func(interface{})(*resources.Resource, error)) (*resources.ResourceState, error) {

// GetChange returns the pointer to the struct with states of the
// certain resource of ComputeInstance type.
func GetChange(change *tfjson.Change) (*resources.ComputeInstanceState, error) {
	before, err := toComputeInstance(change.Before)
	if err != nil {
		return nil, err
	}
	after, err := toComputeInstance(change.After)
	if err != nil {
		return nil, err
	}

	actions := change.Actions
	var action string
	switch true {
	case actions.NoOp():
		action = ActionNoop
	case actions.Create():
		action = ActionCreate
	case actions.Delete():
		action = ActionDelete
	case actions.Update():
		action = ActionUpdate
	case actions.Replace():
		action = ActionReplace
	default:
		log.Println("Wrong action provided.")
	}

	resource := &resources.ComputeInstanceState{
		Before: before,
		After:  after,
		Action: action,
	}
	if resource.Before == nil && resource.After == nil {
		return nil, nil
	}
	return resource, nil
}

// GetResources extracts all resources of ComputeInstance type and their
// before and after states from plan file.
func GetResources(plan *tfjson.Plan) []*resources.ComputeInstanceState {
	var resources []*resources.ComputeInstanceState
	for _, resourceChange := range plan.ResourceChanges {
		if ComputeInstanceType == resourceChange.Type {
			// if resource, err := GetChange(resourceChange.Change, toComputeInstance()); resource != nil && err == nil {
			if resource, err := GetChange(resourceChange.Change); resource != nil && err == nil {
				resources = append(resources, resource)
			}
		}
	}
	return resources
}
