# Multi-region deployment for global P2P network

# US Region
module "us_deployment" {
  source = "./modules/regional"
  
  project_id = var.project_id
  region     = "us-central1"
  zone       = "us-central1-a"
  prefix     = "quiver-us"
  
  bootstrap_peer = google_compute_instance.bootstrap_node.network_interface[0].access_config[0].nat_ip
}

# Europe Region
module "eu_deployment" {
  source = "./modules/regional"
  
  project_id = var.project_id
  region     = "europe-west1"
  zone       = "europe-west1-b"
  prefix     = "quiver-eu"
  
  bootstrap_peer = google_compute_instance.bootstrap_node.network_interface[0].access_config[0].nat_ip
}

# Asia Region (already handled in main.tf)

# Global HTTP(S) Load Balancer
resource "google_compute_global_address" "quiver_global_ip" {
  name = "quiver-global-ip"
}

# Health check for global LB
resource "google_compute_health_check" "global_gateway_health" {
  name                = "quiver-global-gateway-health"
  check_interval_sec  = 10
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 3

  http_health_check {
    request_path = "/health"
    port         = 8080
  }
}

# Backend service for each region
resource "google_compute_backend_service" "global_gateway_backend" {
  name                  = "quiver-global-gateway-backend"
  protocol              = "HTTP"
  port_name             = "http"
  timeout_sec           = 30
  health_checks         = [google_compute_health_check.global_gateway_health.id]
  load_balancing_scheme = "EXTERNAL"
  
  # Asia backend
  backend {
    group           = google_compute_instance_group.gateway_group.id
    balancing_mode  = "RATE"
    max_rate        = 1000
    capacity_scaler = 1.0
  }
  
  # US backend
  dynamic "backend" {
    for_each = module.us_deployment.gateway_instance_group != null ? [1] : []
    content {
      group           = module.us_deployment.gateway_instance_group
      balancing_mode  = "RATE"
      max_rate        = 1000
      capacity_scaler = 1.0
    }
  }
  
  # EU backend
  dynamic "backend" {
    for_each = module.eu_deployment.gateway_instance_group != null ? [1] : []
    content {
      group           = module.eu_deployment.gateway_instance_group
      balancing_mode  = "RATE"
      max_rate        = 1000
      capacity_scaler = 1.0
    }
  }
  
  # Enable CDN for caching
  enable_cdn = true
  
  cdn_policy {
    cache_mode = "CACHE_ALL_STATIC"
    default_ttl = 300
    max_ttl = 3600
    signed_url_cache_max_age_sec = 3600
    
    negative_caching = true
    negative_caching_policy {
      code = 404
      ttl = 120
    }
  }
}

# URL map for global LB
resource "google_compute_url_map" "global_gateway_lb" {
  name            = "quiver-global-gateway-lb"
  default_service = google_compute_backend_service.global_gateway_backend.id
  
  # Path matcher for different regions based on URL
  host_rule {
    hosts        = ["quiver-asia.quiver.network"]
    path_matcher = "asia"
  }
  
  host_rule {
    hosts        = ["quiver-us.quiver.network"]
    path_matcher = "us"
  }
  
  host_rule {
    hosts        = ["quiver-eu.quiver.network"]
    path_matcher = "eu"
  }
  
  path_matcher {
    name            = "asia"
    default_service = google_compute_backend_service.gateway_backend.id
  }
  
  path_matcher {
    name            = "us"
    default_service = module.us_deployment.gateway_backend_service
  }
  
  path_matcher {
    name            = "eu"
    default_service = module.eu_deployment.gateway_backend_service
  }
}

# HTTPS proxy
resource "google_compute_target_https_proxy" "global_gateway_proxy" {
  name             = "quiver-global-gateway-proxy"
  url_map          = google_compute_url_map.global_gateway_lb.id
  ssl_certificates = [google_compute_managed_ssl_certificate.quiver_cert.id]
}

# SSL certificate
resource "google_compute_managed_ssl_certificate" "quiver_cert" {
  name = "quiver-cert"

  managed {
    domains = [
      "quiver-global-lb.quiver.network",
      "quiver-asia.quiver.network",
      "quiver-us.quiver.network",
      "quiver-eu.quiver.network",
      "api.quiver.network"
    ]
  }
}

# Global forwarding rule
resource "google_compute_global_forwarding_rule" "global_gateway_forwarding" {
  name       = "quiver-global-gateway-forwarding"
  target     = google_compute_target_https_proxy.global_gateway_proxy.id
  port_range = "443"
  ip_address = google_compute_global_address.quiver_global_ip.address
}

# Outputs
output "global_gateway_ip" {
  value       = google_compute_global_address.quiver_global_ip.address
  description = "Global load balancer IP address"
}

output "global_gateway_https_url" {
  value       = "https://quiver-global-lb.quiver.network"
  description = "Global load-balanced HTTPS gateway URL"
}

output "regional_urls" {
  value = {
    asia = "https://quiver-asia.quiver.network"
    us   = "https://quiver-us.quiver.network"
    eu   = "https://quiver-eu.quiver.network"
  }
  description = "Regional gateway URLs"
}