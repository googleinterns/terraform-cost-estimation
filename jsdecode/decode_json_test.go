package jsdecode

/*
import (
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	resources "github.com/googleinterns/terraform-cost-estimation/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"os"
	"reflect"
	"testing"
)

var (
	str1 = `
		{
					"allow_stopping_for_update": null,
					"attached_disk": [],
					"boot_disk": [
						{
							"auto_delete": true,
							"disk_encryption_key_raw": null,
							"initialize_params": [
								{
									"image": "debian-cloud/debian-9"
								}
							],
							"mode": "READ_WRITE"
						}
					],
					"can_ip_forward": false,
					"deletion_protection": false,
					"description": null,
					"disk": [],
					"enable_display": null,
					"hostname": null,
					"labels": null,
					"machine_type": "n1-standard-1",
					"metadata": null,
					"metadata_startup_script": null,
					"min_cpu_platform": null,
					"name": "test",
					"network_interface": [
						{
							"access_config": [
								{
									"public_ptr_domain_name": null
								}
							],
							"alias_ip_range": [],
							"network": "default"
						}
					],
					"scratch_disk": [],
					"service_account": [],
					"shielded_instance_config": [],
					"tags": null,
					"timeouts": null,
					"zone": "us-central1-a"
		}
		`

	str2 = `
				{
						"allow_stopping_for_update": null,
						"attached_disk": [],
						"boot_disk": [
								{
								"auto_delete": true,
								"device_name": "persistent-disk-0",
								"disk_encryption_key_raw": "",
								"disk_encryption_key_sha256": "",
								"initialize_params": [
										{
										"image": "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-9-stretch-v20200714",
										"labels": {},
										"size": 10,
										"type": "pd-standard"
										}
								],
								"kms_key_self_link": "",
								"mode": "READ_WRITE",
								"source": "https://www.googleapis.com/compute/v1/projects/google.com:stschmidt/zones/us-central1-a/disks/test"
								}
						],
						"can_ip_forward": false,
						"cpu_platform": "Intel Haswell",
						"deletion_protection": false,
						"description": "",
						"disk": [],
						"enable_display": false,
						"guest_accelerator": [],
						"hostname": "",
						"id": "test",
						"instance_id": "5889159656940809264",
						"label_fingerprint": "42WmSpB8rSM=",
						"labels": {},
						"machine_type": "n1-standard-1",
						"metadata": {},
						"metadata_fingerprint": "s1ovITMUN_Y=",
						"metadata_startup_script": "",
						"min_cpu_platform": "",
						"name": "test",
						"network_interface": [
								{
								"access_config": [
										{
										"assigned_nat_ip": "",
										"nat_-ip": "34.72.220.173",
										"network_tier": "PREMIUM",
										"public_ptr_domain_name": ""
										}
								],
								"address": "",
								"alias_ip_range": [],
								"name": "nic0",
								"network": "https://www.googleapis.com/compute/v1/projects/google.com:stschmidt/global/networks/default",
								"network_ip": "10.128.0.18",
								"subnetwork": "https://www.googleapis.com/compute/v1/projects/google.com:stschmidt/regions/us-central1/subnetworks/default",
								"subnetwork_project": "google.com:stschmidt"
								}
						],
						"project": "google.com:stschmidt",
						"scheduling": [
								{
								"automatic_restart": true,
								"node_affinities": [],
								"on_host_maintenance": "MIGRATE",
								"preemptible": true
								}
						],
						"scratch_disk": [],
						"self_link": "https://www.googleapis.com/compute/v1/projects/google.com:stschmidt/zones/us-central1-a/instances/test",
						"service_account": [],
						"shielded_instance_config": [],
						"tags": [],
						"tags_fingerprint": "42WmSpB8rSM=",
						"timeouts": null,
						"zone": "us-central1-a"
				}
		`
	str3 = `
				{
					"allow_stopping_for_update": null,
					"attached_disk": [],
					"boot_disk": [
						{
							"auto_delete": true,
							"disk_encryption_key_raw": null,
							"initialize_params": [
								{
									"image": "debian-cloud/debian-9"
								}
							],
							"mode": "READ_WRITE"
						}
					],
					"can_ip_forward": false,
					"deletion_protection": false,
					"description": null,
					"disk": [],
					"enable_display": null,
					"hostname": null,
					"labels": null,
					"machine_type": "n1-standard-1",
					"metadata": null,
					"metadata_startup_script": null,
					"min_cpu_platform": null,
					"name": "test-us-east1-a-1",
					"network_interface": [
						{
							"access_config": [],
							"alias_ip_range": []
						}
					],
					"scratch_disk": [],
					"service_account": [],
					"shielded_instance_config": [],
					"tags": null,
					"timeouts": null,
					"zone": "us-east1-a"
				}
		`
	str4 = `
				{
					"allow_stopping_for_update": null,
					"attached_disk": [],
					"boot_disk": [
						{
							"auto_delete": true,
							"disk_encryption_key_raw": null,
							"initialize_params": [
								{
									"image": "debian-cloud/debian-9"
								}
							],
							"mode": "READ_WRITE"
						}
					],
					"can_ip_forward": false,
					"deletion_protection": false,
					"description": null,
					"disk": [],
					"enable_display": null,
					"hostname": null,
					"labels": null,
					"machine_type": "c2-standard-8",
					"metadata": null,
					"metadata_startup_script": null,
					"min_cpu_platform": null,
					"name": "test-c2-standard-8",
					"network_interface": [
						{
							"access_config": [
								{
									"public_ptr_domain_name": null
								}
							],
							"alias_ip_range": [],
							"network": "default"
						}
					],
					"scratch_disk": [],
					"service_account": [],
					"shielded_instance_config": [],
					"tags": null,
					"timeouts": null,
					"zone": "us-central1-a"
				}
		`
)

func TesttoComputeInstance(t *testing.T) {
	var res1, res2, res3, res4 interface{}
	json.Unmarshal([]byte(str1), &res1)
	json.Unmarshal([]byte(str2), &res2)
	json.Unmarshal([]byte(str3), &res3)
	json.Unmarshal([]byte(str4), &res4)

	out1, _ := resources.NewComputeInstance("", "test", "n1-standard-1", "us-central1-a", "OnDemand")
	out2, _ := resources.NewComputeInstance("5889159656940809264", "test", "n1-standard-1", "us-central1-a", "Preemptible")
	out3, _ := resources.NewComputeInstance("", "test-us-east1-a-1", "n1-standard-1", "us-east1-a", "OnDemand")
	out4, _ := resources.NewComputeInstance("", "test-c2-standard-8", "c2-standard-8", "us-central1-a", "OnDemand")

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
		actual, err := toComputeInstance(test.in)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", spew.Sdump(test.expected), spew.Sdump(actual))
		}
	}
}

func TestGetChange(t *testing.T) {
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

	after, _ := resources.NewComputeInstance("", "test", "n1-standard-1", "us-central1-a", "OnDemand")
	expected := &resources.ComputeInstanceState{
		Before: nil,
		After:  after,
		Action: "create",
	}

	var actual *resources.ComputeInstanceState
	actual, err = GetChange(plan.ResourceChanges[0].Change)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", spew.Sdump(expected), spew.Sdump(actual))
	}

}

// TODO add test without "google_compute_instance".
func TestGetResources(t *testing.T) {
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

	before, _ := resources.NewComputeInstance("5889159656940809264", "test", "n1-standard-1", "us-central1-a", "OnDemand")
	after, _ := resources.NewComputeInstance("5889159656940809264", "test", "n1-standard-2", "us-central1-a", "OnDemand")
	expected := []*resources.ComputeInstanceState{
		&resources.ComputeInstanceState{
			Before: before,
			After:  after,
			Action: "update",
		},
	}

	actual := GetResources(plan)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected:\n\n%s\n\ngot:\n\n%s", spew.Sdump(expected), spew.Sdump(actual))
	}
}
*/
