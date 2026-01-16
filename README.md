# TailSwan

Bridge strongSwan/swanctl IPsec VPN and Tailscale networks in a Docker container.

## Overview

TailSwan allows you to connect existing IPsec VPN networks to your Tailscale network (tailnet). It's perfect for:

- Devices that don't support Tailscale but have IPsec VPN capability
- Bridging legacy IPsec infrastructure with modern Tailscale networks
- Exposing IPsec-connected subnets to your tailnet
- Centralized IPsec gateway management with Tailscale's security and access controls

## Features

- **Dual VPN Bridge**: Runs both strongSwan (IPsec) and Tailscale in a single container
- **Automatic Subnet Advertisement**: IPsec subnets are automatically advertised to your Tailscale network
- **Tailscale SSH Access**: SSH into the container via Tailscale to manage IPsec connections
- **Multiple Authentication Methods**: Supports PSK, certificates, and other IPsec auth methods
- **Site-to-Site & Remote Access**: Configure both site-to-site and road warrior VPN scenarios
- **Health Monitoring**: Built-in health checks for both services
- **Easy Configuration**: Environment variables and standard swanctl.conf syntax

## Architecture

```
[IPsec Client/Site] <--IPsec--> [TailSwan Container] <--Tailscale--> [Your Tailnet]
                                 ├── strongSwan (charon)
                                 ├── Tailscale client
                                 └── SSH daemon (via Tailscale)
```

## Quick Start

### Prerequisites

- Docker installed
- Tailscale account
- IPsec peer configuration details

### Pre-built Images

TailSwan images are automatically built and published to GitHub Container Registry:

```bash
# Pull the latest image from GHCR
docker pull ghcr.io/klowdo/tailswan:latest

# Or use a specific version
docker pull ghcr.io/klowdo/tailswan:v1.0.0
```

Supported platforms: `linux/amd64`, `linux/arm64`

### Basic Usage

1. **Create a swanctl configuration file** (`swanctl.conf`):

```conf
connections {
    mysite {
        version = 2
        local_addrs = 0.0.0.0
        remote_addrs = 203.0.113.10

        local {
            auth = psk
            id = my-gateway
        }

        remote {
            auth = psk
            id = remote-gateway
        }

        children {
            net-net {
                local_ts = 10.1.0.0/24
                remote_ts = 10.2.0.0/24
                start_action = trap
                esp_proposals = aes256-sha256-modp2048
            }
        }

        proposals = aes256-sha256-modp2048
    }
}

secrets {
    ike-mysite {
        id = my-gateway
        secret = "your-strong-preshared-key"
    }
}
```

2. **Run the container**:

```bash
docker run -d \
  --name tailswan \
  --cap-add=NET_ADMIN \
  --cap-add=NET_RAW \
  --device=/dev/net/tun \
  -v $(pwd)/swanctl.conf:/etc/swanctl/swanctl.conf:ro \
  -e TS_AUTHKEY=tskey-auth-xxxxxxxxxxxx \
  -e TS_ROUTES=10.1.0.0/24,10.2.0.0/24 \
  -e TS_HOSTNAME=tailswan-gateway \
  -e TS_SSH=true \
  ghcr.io/klowdo/tailswan:latest
```

3. **Approve routes in Tailscale admin console**:
   - Go to https://login.tailscale.com/admin/machines
   - Find your TailSwan device
   - Click "Edit route settings"
   - Approve the advertised routes

4. **SSH into the container via Tailscale**:

```bash
ssh root@tailswan-gateway
```

5. **Initiate the IPsec connection** (if not using auto-start):

```bash
swanctl --initiate --child net-net
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TS_AUTHKEY` | - | Tailscale auth key (get from https://login.tailscale.com/admin/settings/keys) |
| `TS_HOSTNAME` | `tailswan` | Hostname for the Tailscale node |
| `TS_ROUTES` | - | Comma-separated list of subnets to advertise (e.g., `10.1.0.0/24,10.2.0.0/24`) |
| `TS_SSH` | `true` | Enable Tailscale SSH server |
| `TS_EXTRA_ARGS` | - | Additional arguments for `tailscale up` |
| `TS_STATE_DIR` | `/var/lib/tailscale` | Tailscale state directory |
| `SWAN_CONFIG` | `/etc/swanctl/swanctl.conf` | Path to swanctl configuration file |
| `SWAN_AUTO_START` | `false` | Auto-start IPsec connections on container start |
| `SWAN_CONNECTIONS` | - | Comma-separated list of connections to auto-start (requires `SWAN_AUTO_START=true`) |

## Configuration Examples

### Example 1: Site-to-Site VPN with Auto-Start

```bash
docker run -d \
  --name tailswan \
  --cap-add=NET_ADMIN \
  --cap-add=NET_RAW \
  --device=/dev/net/tun \
  -v $(pwd)/swanctl.conf:/etc/swanctl/swanctl.conf:ro \
  -e TS_AUTHKEY=tskey-auth-xxxxxxxxxxxx \
  -e TS_ROUTES=10.1.0.0/24,10.2.0.0/24 \
  -e TS_HOSTNAME=office-gateway \
  -e SWAN_AUTO_START=true \
  -e SWAN_CONNECTIONS=net-net \
  tailswan:latest
```

### Example 2: Using Docker Compose

See `docker-compose.yml` for a complete example.

### Example 3: Certificate-based Authentication

Mount certificate directories:

```bash
docker run -d \
  --name tailswan \
  --cap-add=NET_ADMIN \
  --cap-add=NET_RAW \
  --device=/dev/net/tun \
  -v $(pwd)/swanctl.conf:/etc/swanctl/swanctl.conf:ro \
  -v $(pwd)/certs/x509:/etc/swanctl/x509:ro \
  -v $(pwd)/certs/x509ca:/etc/swanctl/x509ca:ro \
  -v $(pwd)/certs/private:/etc/swanctl/private:ro \
  -e TS_AUTHKEY=tskey-auth-xxxxxxxxxxxx \
  -e TS_ROUTES=10.0.0.0/8 \
  tailswan:latest
```

## Management

### SSH Access via Tailscale

Once the container is running and connected to your tailnet:

```bash
# SSH using the hostname
ssh root@tailswan-gateway

# Or using the Tailscale IP
ssh root@100.x.y.z
```

### Helper Script

Inside the container, use the `swan-status.sh` script:

```bash
# Show complete status
/tailswan/swan-status.sh status

# List connections
/tailswan/swan-status.sh connections

# Start a connection
/tailswan/swan-status.sh start mysite

# Stop a connection
/tailswan/swan-status.sh stop mysite

# Reload configuration
/tailswan/swan-status.sh reload
```

### Manual swanctl Commands

```bash
# Load configuration
swanctl --load-all

# List configured connections
swanctl --list-conns

# Initiate a connection
swanctl --initiate --child net-net

# List active Security Associations
swanctl --list-sas

# Terminate a connection
swanctl --terminate --ike mysite
```

## Building from Source

```bash
git clone https://github.com/yourusername/tailswan.git
cd tailswan
docker build -t tailswan:latest .
```

## Troubleshooting

### Container fails to start

- Ensure you have the required capabilities: `NET_ADMIN`, `NET_RAW`
- Check that `/dev/net/tun` device is available
- Verify your swanctl.conf syntax

### IPsec tunnel not establishing

```bash
# Check logs
docker logs tailswan

# SSH into container and check status
ssh root@tailswan-gateway
swanctl --list-sas
journalctl -u strongswan -f  # if available
```

### Routes not appearing in Tailscale

- Verify routes are approved in Tailscale admin console
- Check that `TS_ROUTES` matches your IPsec subnet configuration
- Ensure IP forwarding is enabled (automatically set by entrypoint)

### Cannot SSH via Tailscale

- Verify `TS_SSH=true` is set
- Check Tailscale ACLs allow SSH access
- Ensure the container is showing as "Connected" in Tailscale admin

## Security Considerations

1. **Authentication Keys**: Use environment variables or Docker secrets for sensitive data
2. **Tailscale ACLs**: Configure proper access controls in your Tailscale admin console
3. **IPsec Credentials**: Prefer certificate-based authentication over PSK for production
4. **SSH Access**: Leverage Tailscale's built-in SSH with ACL-based access control
5. **Subnet Routing**: Only advertise routes you intend to expose to your tailnet

## Advanced Configuration

### Route-based VPN with VTI

For more complex routing scenarios, you can use Virtual Tunnel Interfaces (VTI). See the `swanctl.conf.example` file for configuration details.

### Multiple IPsec Connections

Configure multiple connections in your swanctl.conf and use `SWAN_CONNECTIONS` to auto-start them:

```bash
-e SWAN_CONNECTIONS=site1,site2,site3
```

### Custom DNS

```bash
-e TG_NAMESERVERS=8.8.8.8,1.1.1.1
```

## Similar Projects

- [TailGuard](https://github.com/juhovh/tailguard) - Bridges WireGuard and Tailscale (inspiration for this project)

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See LICENSE file for details.

## Resources

- [strongSwan Documentation](https://docs.strongswan.org/)
- [swanctl.conf Reference](https://docs.strongswan.org/docs/latest/swanctl/swanctlConf.html)
- [Tailscale Documentation](https://tailscale.com/kb/)
- [Tailscale Subnet Routers](https://tailscale.com/kb/1019/subnets)
