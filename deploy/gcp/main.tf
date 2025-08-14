terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "asia-northeast1"
}

variable "zone" {
  description = "GCP Zone"
  type        = string
  default     = "asia-northeast1-a"
}

# VPC Network
resource "google_compute_network" "quiver_network" {
  name                    = "quiver-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "quiver_subnet" {
  name          = "quiver-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.quiver_network.id
}

# Firewall rules
resource "google_compute_firewall" "quiver_p2p" {
  name    = "quiver-p2p"
  network = google_compute_network.quiver_network.name

  allow {
    protocol = "tcp"
    ports    = ["4001-4010", "8080-8090"]
  }

  allow {
    protocol = "udp"
    ports    = ["4001-4010"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["quiver-node"]
}

resource "google_compute_firewall" "quiver_ssh" {
  name    = "quiver-ssh"
  network = google_compute_network.quiver_network.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["quiver-node"]
}

# Bootstrap node
resource "google_compute_instance" "bootstrap_node" {
  name         = "quiver-bootstrap"
  machine_type = "e2-medium"
  zone         = var.zone
  tags         = ["quiver-node", "bootstrap"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 20
    }
  }

  network_interface {
    network    = google_compute_network.quiver_network.name
    subnetwork = google_compute_subnetwork.quiver_subnet.name

    access_config {
      // Ephemeral public IP
    }
  }

  metadata_startup_script = file("${path.module}/scripts/bootstrap-node.sh")

  service_account {
    scopes = ["cloud-platform"]
  }
}

# Provider nodes
resource "google_compute_instance" "provider_nodes" {
  count        = 2
  name         = "quiver-provider-${count.index + 1}"
  machine_type = "n1-standard-2"
  zone         = var.zone
  tags         = ["quiver-node", "provider"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 50
    }
  }

  network_interface {
    network    = google_compute_network.quiver_network.name
    subnetwork = google_compute_subnetwork.quiver_subnet.name

    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    bootstrap_addr = google_compute_instance.bootstrap_node.network_interface[0].access_config[0].nat_ip
  }

  metadata_startup_script = file("${path.module}/scripts/provider-node.sh")

  service_account {
    scopes = ["cloud-platform"]
  }

  depends_on = [google_compute_instance.bootstrap_node]
}

# Gateway node
resource "google_compute_instance" "gateway_node" {
  name         = "quiver-gateway"
  machine_type = "e2-medium"
  zone         = var.zone
  tags         = ["quiver-node", "gateway"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 20
    }
  }

  network_interface {
    network    = google_compute_network.quiver_network.name
    subnetwork = google_compute_subnetwork.quiver_subnet.name

    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    bootstrap_addr = google_compute_instance.bootstrap_node.network_interface[0].access_config[0].nat_ip
  }

  metadata_startup_script = file("${path.module}/scripts/gateway-node.sh")

  service_account {
    scopes = ["cloud-platform"]
  }

  depends_on = [google_compute_instance.bootstrap_node]
}

# Realtime stats node
resource "google_compute_instance" "stats_node" {
  name         = "quiver-stats"
  machine_type = "e2-micro"
  zone         = var.zone
  tags         = ["quiver-node", "stats"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 20
    }
  }

  network_interface {
    network    = google_compute_network.quiver_network.name
    subnetwork = google_compute_subnetwork.quiver_subnet.name

    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    bootstrap_addr = google_compute_instance.bootstrap_node.network_interface[0].access_config[0].nat_ip
  }

  metadata_startup_script = file("${path.module}/scripts/stats-node.sh")

  service_account {
    scopes = ["cloud-platform"]
  }

  depends_on = [google_compute_instance.bootstrap_node]
}

# Outputs
output "bootstrap_ip" {
  value = google_compute_instance.bootstrap_node.network_interface[0].access_config[0].nat_ip
}

output "provider_ips" {
  value = google_compute_instance.provider_nodes[*].network_interface[0].access_config[0].nat_ip
}

output "gateway_ip" {
  value = google_compute_instance.gateway_node.network_interface[0].access_config[0].nat_ip
}

output "stats_ip" {
  value = google_compute_instance.stats_node.network_interface[0].access_config[0].nat_ip
}

output "stats_websocket_url" {
  value = "ws://${google_compute_instance.stats_node.network_interface[0].access_config[0].nat_ip}:8087/ws"
}