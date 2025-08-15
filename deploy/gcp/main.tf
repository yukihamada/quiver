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

  metadata_startup_script = file("${path.module}/scripts/bootstrap-node-fixed.sh")

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

# Multiple gateway nodes for load balancing
resource "google_compute_instance" "gateway_nodes" {
  count        = 3
  name         = "quiver-gateway-${count.index + 1}"
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

  metadata_startup_script = file("${path.module}/scripts/gateway-node-fixed.sh")

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

# Instance group for gateway nodes
resource "google_compute_instance_group" "gateway_group" {
  name        = "quiver-gateway-group"
  description = "Gateway instance group"
  zone        = var.zone

  instances = google_compute_instance.gateway_nodes[*].self_link

  named_port {
    name = "http"
    port = 8080
  }
}

# Health check
resource "google_compute_health_check" "gateway_health" {
  name                = "quiver-gateway-health"
  check_interval_sec  = 5
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 2

  http_health_check {
    request_path = "/health"
    port         = 8080
  }
}

# Backend service
resource "google_compute_backend_service" "gateway_backend" {
  name                  = "quiver-gateway-backend"
  protocol              = "HTTP"
  port_name             = "http"
  timeout_sec           = 30
  health_checks         = [google_compute_health_check.gateway_health.id]
  load_balancing_scheme = "EXTERNAL"

  backend {
    group           = google_compute_instance_group.gateway_group.id
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
}

# URL map
resource "google_compute_url_map" "gateway_lb" {
  name            = "quiver-gateway-lb"
  default_service = google_compute_backend_service.gateway_backend.id
}

# HTTP proxy
resource "google_compute_target_http_proxy" "gateway_proxy" {
  name    = "quiver-gateway-proxy"
  url_map = google_compute_url_map.gateway_lb.id
}

# Forwarding rule
resource "google_compute_global_forwarding_rule" "gateway_forwarding" {
  name       = "quiver-gateway-forwarding"
  target     = google_compute_target_http_proxy.gateway_proxy.id
  port_range = "80"
}

# Outputs
output "bootstrap_ip" {
  value = google_compute_instance.bootstrap_node.network_interface[0].access_config[0].nat_ip
}

output "provider_ips" {
  value = google_compute_instance.provider_nodes[*].network_interface[0].access_config[0].nat_ip
}

output "gateway_ips" {
  value = google_compute_instance.gateway_nodes[*].network_interface[0].access_config[0].nat_ip
}

output "stats_ip" {
  value = google_compute_instance.stats_node.network_interface[0].access_config[0].nat_ip
}

output "stats_websocket_url" {
  value = "ws://${google_compute_instance.stats_node.network_interface[0].access_config[0].nat_ip}:8087/ws"
}

output "global_gateway_url" {
  value = "http://${google_compute_global_forwarding_rule.gateway_forwarding.ip_address}"
  description = "Global load-balanced gateway URL"
}