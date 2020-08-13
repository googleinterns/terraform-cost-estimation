variable "region" {}

variable "zone" {}

resource "google_compute_instance" "default" {
  count = 3

  name         = "test-${var.region}-${var.zone}-${count.index}"
  machine_type = "n1-standard-1"
  zone         = "${var.region}-${var.zone}"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-9"
    }
  }

  network_interface {
  }
}
