# WebSocket Signalling Server for WebRTC

# Signalling server instance
resource "google_compute_instance" "signalling_server" {
  name         = "quiver-signalling"
  machine_type = "e2-medium"
  zone         = var.zone
  tags         = ["quiver-node", "signalling", "https-server", "websocket"]

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

  metadata_startup_script = <<-EOF
    #!/bin/bash
    set -e
    
    # Install dependencies
    apt-get update
    apt-get install -y curl wget git golang-go certbot nginx
    
    # Install Go 1.21
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    
    # Clone QUIVer repository
    cd /opt
    git clone https://github.com/yukihamada/quiver.git
    cd quiver/gateway
    
    # Build signalling server
    /usr/local/go/bin/go build -o /usr/local/bin/quiver-signalling ./cmd/signalling
    
    # Create systemd service
    cat > /etc/systemd/system/quiver-signalling.service <<EOL
[Unit]
Description=QUIVer Signalling Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/quiver-signalling
Restart=always
Environment="PORT=8444"

[Install]
WantedBy=multi-user.target
EOL
    
    # Configure nginx for WebSocket proxy
    cat > /etc/nginx/sites-available/signalling <<EOL
server {
    listen 80;
    server_name signal.quiver.network signal-asia.quiver.network;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name signal.quiver.network signal-asia.quiver.network;
    
    ssl_certificate /etc/letsencrypt/live/signal.quiver.network/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/signal.quiver.network/privkey.pem;
    
    location /signal {
        proxy_pass http://localhost:8444;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOL
    
    # Enable nginx site
    ln -s /etc/nginx/sites-available/signalling /etc/nginx/sites-enabled/
    
    # Start services
    systemctl enable quiver-signalling
    systemctl start quiver-signalling
    systemctl restart nginx
    
    echo "Signalling server deployed!"
  EOF

  service_account {
    scopes = ["cloud-platform"]
  }
}

# Public bootstrap nodes with enhanced configuration
resource "google_compute_instance" "public_bootstrap" {
  count        = 3
  name         = "quiver-bootstrap-public-${count.index + 1}"
  machine_type = "e2-standard-2"
  zone         = var.zone
  tags         = ["quiver-node", "bootstrap", "public"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 30
    }
  }

  network_interface {
    network    = google_compute_network.quiver_network.name
    subnetwork = google_compute_subnetwork.quiver_subnet.name

    access_config {
      // Ephemeral public IP
    }
  }

  metadata_startup_script = <<-EOF
    #!/bin/bash
    set -e
    
    # Install dependencies
    apt-get update
    apt-get install -y curl wget git build-essential
    
    # Install Go
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    
    # Clone and build QUIVer
    cd /opt
    git clone https://github.com/yukihamada/quiver.git
    cd quiver
    
    # Build bootstrap with WebSocket support
    cd bootstrap
    /usr/local/go/bin/go build -o /usr/local/bin/quiver-bootstrap .
    
    # Generate peer ID
    PEER_ID=$(/usr/local/bin/quiver-bootstrap --generate-id)
    echo $PEER_ID > /etc/quiver-peer-id
    
    # Create bootstrap configuration
    cat > /etc/quiver-bootstrap.json <<EOL
{
  "peer_id": "$PEER_ID",
  "listen_addresses": [
    "/ip4/0.0.0.0/tcp/4001",
    "/ip4/0.0.0.0/udp/4001/quic-v1",
    "/ip4/0.0.0.0/tcp/4003/ws"
  ],
  "announce_addresses": [
    "/dns4/bootstrap${count.index + 1}.quiver.network/tcp/4001/p2p/$PEER_ID",
    "/dns4/bootstrap${count.index + 1}.quiver.network/udp/4001/quic-v1/p2p/$PEER_ID",
    "/dns4/bootstrap${count.index + 1}.quiver.network/tcp/443/wss/p2p/$PEER_ID"
  ],
  "enable_relay": true,
  "enable_nat_service": true
}
EOL
    
    # Create systemd service
    cat > /etc/systemd/system/quiver-bootstrap.service <<EOL
[Unit]
Description=QUIVer Bootstrap Node (Public)
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/quiver-bootstrap --config /etc/quiver-bootstrap.json --public
Restart=always
Environment="QUIVER_ENABLE_WEBSOCKET=true"
Environment="QUIVER_PUBLIC_NODE=true"

[Install]
WantedBy=multi-user.target
EOL
    
    # Configure nginx for WebSocket
    apt-get install -y nginx certbot
    
    cat > /etc/nginx/sites-available/bootstrap <<EOL
server {
    listen 80;
    server_name bootstrap${count.index + 1}.quiver.network;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name bootstrap${count.index + 1}.quiver.network;
    
    ssl_certificate /etc/letsencrypt/live/bootstrap${count.index + 1}.quiver.network/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/bootstrap${count.index + 1}.quiver.network/privkey.pem;
    
    location /p2p {
        proxy_pass http://localhost:4003;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOL
    
    ln -s /etc/nginx/sites-available/bootstrap /etc/nginx/sites-enabled/
    
    # Start services
    systemctl enable quiver-bootstrap
    systemctl start quiver-bootstrap
    systemctl restart nginx
    
    # Register with stats server
    curl -X POST http://${google_compute_instance.stats_node.network_interface[0].access_config[0].nat_ip}:8087/api/register \
      -H "Content-Type: application/json" \
      -d "{\"peer_id\": \"$PEER_ID\", \"type\": \"bootstrap\", \"public\": true}"
    
    echo "Public bootstrap node ${count.index + 1} deployed!"
  EOF

  service_account {
    scopes = ["cloud-platform"]
  }
}

# Firewall rules for WebSocket
resource "google_compute_firewall" "websocket" {
  name    = "quiver-websocket"
  network = google_compute_network.quiver_network.name

  allow {
    protocol = "tcp"
    ports    = ["443", "4003", "8444"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["websocket", "public"]
}

# Cloud DNS configuration
resource "google_dns_managed_zone" "quiver_zone" {
  name        = "quiver-network"
  dns_name    = "quiver.network."
  description = "QUIVer P2P Network DNS Zone"
}

# DNS records for bootstrap nodes
resource "google_dns_record_set" "bootstrap_records" {
  count = 3
  name  = "bootstrap${count.index + 1}.${google_dns_managed_zone.quiver_zone.dns_name}"
  type  = "A"
  ttl   = 300

  managed_zone = google_dns_managed_zone.quiver_zone.name
  rrdatas      = [google_compute_instance.public_bootstrap[count.index].network_interface[0].access_config[0].nat_ip]
}

# DNS record for signalling server
resource "google_dns_record_set" "signal_record" {
  name = "signal.${google_dns_managed_zone.quiver_zone.dns_name}"
  type = "A"
  ttl  = 300

  managed_zone = google_dns_managed_zone.quiver_zone.name
  rrdatas      = [google_compute_instance.signalling_server.network_interface[0].access_config[0].nat_ip]
}

# Regional aliases
resource "google_dns_record_set" "signal_asia" {
  name = "signal-asia.${google_dns_managed_zone.quiver_zone.dns_name}"
  type = "CNAME"
  ttl  = 300

  managed_zone = google_dns_managed_zone.quiver_zone.name
  rrdatas      = ["signal.${google_dns_managed_zone.quiver_zone.dns_name}"]
}

# Outputs
output "signalling_server_ip" {
  value = google_compute_instance.signalling_server.network_interface[0].access_config[0].nat_ip
}

output "bootstrap_ips" {
  value = google_compute_instance.public_bootstrap[*].network_interface[0].access_config[0].nat_ip
}

output "bootstrap_peer_ids" {
  value = [for i in range(3) : "Check /etc/quiver-peer-id on bootstrap${i + 1}"]
}

output "websocket_endpoints" {
  value = {
    signalling = "wss://signal.quiver.network/signal"
    bootstrap = [for i in range(3) : "wss://bootstrap${i + 1}.quiver.network/p2p"]
  }
}