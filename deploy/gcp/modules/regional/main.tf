variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
}

variable "zone" {
  description = "GCP Zone"
  type        = string
}

variable "prefix" {
  description = "Resource name prefix"
  type        = string
}

variable "bootstrap_peer" {
  description = "Bootstrap peer address"
  type        = string
}

# Gateway nodes
resource "google_compute_instance" "gateway_nodes" {
  count        = 2
  name         = "${var.prefix}-gateway-${count.index + 1}"
  machine_type = "e2-medium"
  zone         = var.zone
  tags         = ["quiver-node", "gateway", var.prefix]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 20
    }
  }

  network_interface {
    network = "quiver-network"
    
    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    bootstrap_addr = var.bootstrap_peer
    region        = var.region
  }

  metadata_startup_script = <<-EOF
    #!/bin/bash
    # Install dependencies
    apt-get update
    apt-get install -y curl wget git golang-go
    
    # Install QUIVer gateway
    wget -q https://github.com/yukihamada/quiver/releases/latest/download/gateway-linux-amd64 -O /usr/local/bin/gateway
    chmod +x /usr/local/bin/gateway
    
    # Configure and start gateway
    export QUIVER_BOOTSTRAP="/ip4/${var.bootstrap_peer}/tcp/4001"
    export QUIVER_REGION="${var.region}"
    
    # Start gateway with systemd
    cat > /etc/systemd/system/quiver-gateway.service <<EOL
[Unit]
Description=QUIVer Gateway
After=network.target

[Service]
Type=simple
User=root
Environment="QUIVER_BOOTSTRAP=/ip4/${var.bootstrap_peer}/tcp/4001"
Environment="QUIVER_REGION=${var.region}"
ExecStart=/usr/local/bin/gateway
Restart=always

[Install]
WantedBy=multi-user.target
EOL
    
    systemctl enable quiver-gateway
    systemctl start quiver-gateway
  EOF
}

# Instance group
resource "google_compute_instance_group" "gateway_group" {
  name        = "${var.prefix}-gateway-group"
  description = "Gateway instance group for ${var.region}"
  zone        = var.zone

  instances = google_compute_instance.gateway_nodes[*].self_link

  named_port {
    name = "http"
    port = 8080
  }
}

# Regional backend service
resource "google_compute_backend_service" "gateway_backend" {
  name                  = "${var.prefix}-gateway-backend"
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

# Health check
resource "google_compute_health_check" "gateway_health" {
  name                = "${var.prefix}-gateway-health"
  check_interval_sec  = 5
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 2

  http_health_check {
    request_path = "/health"
    port         = 8080
  }
}

# Outputs
output "gateway_instance_group" {
  value = google_compute_instance_group.gateway_group.id
}

output "gateway_backend_service" {
  value = google_compute_backend_service.gateway_backend.id
}

output "gateway_ips" {
  value = google_compute_instance.gateway_nodes[*].network_interface[0].access_config[0].nat_ip
}