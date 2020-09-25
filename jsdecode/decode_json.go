package jsdecode

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	resources "github.com/googleinterns/terraform-cost-estimation/resources"
	cd "github.com/googleinterns/terraform-cost-estimation/resources/classdetail"
	tfjson "github.com/hashicorp/terraform-json"
)

// ComputeInstanceType and ComputeDiskType are the supported by this package types of ResourceChange and Resource.
const (
	ComputeDiskType     = "google_compute_disk"
	ComputeInstanceType = "google_compute_instance"
)

// Possible actions in resource changes.
const (
	ActionCreate  string = "create"
	ActionDelete  string = "delete"
	ActionNoop    string = "no-op"
	ActionReplace string = "replace"
	ActionUpdate  string = "update"
)

// ResourceInfo contains the information about Resource in json plan file, this struct is used to
// cast interface of before/after states to the certain Resource.
type ResourceInfo struct {
	Name        string      `json:"name,omitempty"`
	InstanceID  string      `json:"instance_id,omitempty"`
	Zone        string      `json:"zone,omitempty"`
	MachineType string      `json:"machine_type,omitempty"`
	DiskType    string      `json:"type,omitempty"`
	Image       string      `json:"image,omitempty"`
	Snapshot    string      `json:"snapshot,omitempty"`
	SizeGiB     int64       `json:"size,omitempty"`
	Scheduling  []UsageType `json:"scheduling,omitempty"`
}

// UsageType contains the information whether the certain istance is preemptible or not.
type UsageType struct {
	IsPreemptible bool `json:"preemptible,omitempty"`
}

// ExtractPlanStruct extracts tfjson.Plan struct from file in provided path if it is possible.
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

// toComputeInstance extracts ComputeInstance from the interface that contains information about the resource.
func toComputeInstance(details *cd.ResourceDetail, resource interface{}) (*resources.ComputeInstance, error) {
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
	return resources.NewComputeInstance(details, r.InstanceID, r.Name, r.MachineType, r.Zone, usageType)
}

// toComputeDisk extracts ComputeDisk from the interface that contains information about the resource.
func toComputeDisk(details *cd.ResourceDetail, resource interface{}) (*resources.ComputeDisk, error) {
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
	var zones []string
	zones = append(zones, r.Zone)
	return resources.NewComputeDisk(details, r.Name, r.InstanceID, r.DiskType, zones, r.Image, r.Snapshot, r.SizeGiB)
}

// toInstanceState returns the pointer to the struct with states of the certain resource of ComputeInstance type.
func toInstanceState(details *cd.ResourceDetail, change *tfjson.Change) (*resources.ComputeInstanceState, error) {
	before, err := toComputeInstance(details, change.Before)
	if err != nil {
		return nil, err
	}
	after, err := toComputeInstance(details, change.After)
	if err != nil {
		return nil, err
	}
	if before == nil && after == nil {
		return nil, nil
	}

	action, err := initAction(change.Actions)
	if err != nil {
		return nil, err
	}
	return &resources.ComputeInstanceState{
		Before: before,
		After:  after,
		Action: action,
	}, nil
}

// toDiskState returns the pointer to the struct with states of the certain
// resource of ComputeInstance type.
func toDiskState(details *cd.ResourceDetail, change *tfjson.Change) (*resources.ComputeDiskState, error) {
	before, err := toComputeDisk(details, change.Before)
	if err != nil {
		return nil, err
	}
	after, err := toComputeDisk(details, change.After)
	if err != nil {
		return nil, err
	}
	if before == nil && after == nil {
		return nil, nil
	}

	action, err := initAction(change.Actions)
	if err != nil {
		return nil, err
	}
	return &resources.ComputeDiskState{
		Before: before,
		After:  after,
		Action: action,
	}, nil
}

// initAction extracts an action in the change.
func initAction(actions tfjson.Actions) (string, error) {
	var action string
	switch {
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
		return action, fmt.Errorf("Wrong action provided.")
	}
	return action, nil
}

// GetResources extracts all resources of ComputeInstance and ComputeDisk type and their before and after states from plan file.
func GetResources(details *cd.ResourceDetail, plan *tfjson.Plan) []resources.ResourceState {
	var states []resources.ResourceState
	var r resources.ResourceState
	var err error
	for _, resourceChange := range plan.ResourceChanges {
		switch resourceChange.Type {
		case ComputeInstanceType:
			r, err = toInstanceState(details, resourceChange.Change)
		case ComputeDiskType:
			r, err = toDiskState(details, resourceChange.Change)
		default:
			log.Printf("Unsupported resource type: %v", resourceChange.Type)
		}
		if err != nil {
			log.Printf("Error: %v", err)
		} else if r != nil {
			states = append(states, r)
		}
		r, err = nil, nil
	}
	return states
}
