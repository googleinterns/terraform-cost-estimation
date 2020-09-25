package resources

import (
	"fmt"
	"strings"
)

// Description holds information about information of the SKU description, whith strings to be included/omitted.
type Description struct {
	Contains []string
	Omits    []string
}

func (d *Description) fillForComputeInstance(machineType, usageType string) error {
	anythingButN1 := []string{"N2", "N2D", "E2", "Compute", "Memory", "Sole Tenancy"}

	if usageType == "Preemptible" {
		d.Contains = append(d.Contains, "Preemptible")
	} else {
		d.Omits = append(d.Omits, "Preemptible")
	}

	// Commitment N1 machines don't have "Commitment" specified.
	if strings.HasPrefix(usageType, "Commit") {
		d.Contains = append(d.Contains, "Commitment")
		if strings.Contains(machineType, "n1") {
			d.Omits = append(d.Omits, "N1")
			d.Omits = append(d.Omits, anythingButN1...)
		}
	} else {
		d.Omits = append(d.Omits, "Commitment")
	}

	// Custom E2 machines don't have separate SKUs.
	if strings.Contains(machineType, "custom") {
		if !strings.HasPrefix(machineType, "e2") {
			d.Contains = append(d.Contains, "Custom")
		}
	} else {
		d.Omits = append(d.Omits, "Custom")
	}

	// Custom N1 machines don't have any type specified, so all types must be excluded.
	if strings.HasPrefix(machineType, "custom") {
		d.Omits = append(d.Omits, "N1")
		d.Omits = append(d.Omits, anythingButN1...)
	} else {

		switch {
		case strings.HasPrefix(machineType, "c2-"):
			d.Contains = append(d.Contains, "Compute")

		case strings.HasPrefix(machineType, "m1-") || strings.HasPrefix(machineType, "m2-"):
			d.Contains = append(d.Contains, "Memory")
			d.Omits = append(d.Omits, "Upgrade")

		case strings.HasPrefix(machineType, "n1-mega") || strings.HasPrefix(machineType, "n1-ultra"):
			d.Contains = append(d.Contains, "Memory")
			d.Omits = append(d.Omits, "Upgrade")

		case strings.HasPrefix(machineType, "n1-") || strings.HasPrefix(machineType, "f1-") || strings.HasPrefix(machineType, "g1-"):
			if !strings.HasPrefix(usageType, "Commit") {
				d.Contains = append(d.Contains, "N1")
			}

		default:
			// All other machines have their type specified.
			i := strings.Index(machineType, "-")
			if i < 0 {
				return fmt.Errorf("wrong machine type format")
			}

			d.Contains = append(d.Contains, strings.ToUpper(machineType[:i])+" ")
		}
	}

	return nil
}

func (d *Description) fillForComputeDisk(diskType string, regional bool) {
	switch diskType {
	case "pd-standard":
		d.Contains = []string{"Storage PD Capacity"}
	case "pd-ssd":
		d.Contains = []string{"SSD backed PD Capacity"}
	default:
	}

	if regional {
		d.Contains = append(d.Contains, "Regional")
	} else {
		d.Omits = append(d.Omits, "Regional")
	}
}
