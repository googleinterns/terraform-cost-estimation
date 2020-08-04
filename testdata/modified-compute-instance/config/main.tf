resource "google_compute_instance" "default" {
  project      = var.project_id
  name         = "test"
  machine_type = var.previous_state ? "n1-standard-1" : "n1-standard-2"
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
