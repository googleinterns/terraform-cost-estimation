locals {
  machine_types = split("\n", file("machine-types.txt"))
}

resource "google_compute_instance" "default" {
  for_each = toset(local.machine_types)

  name         = "test-${each.value}"
  machine_type = each.value
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-9"
    }
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral IP
    }
  }
}
