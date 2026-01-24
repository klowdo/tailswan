# TailSwan Control Server

HTTP API server with web UI for controlling strongSwan/swanctl IPsec connections via the VICI interface.

## Overview

The TailSwan Control Server provides both a web-based UI and a RESTful HTTP API to manage IPsec connections using the govici library to communicate with strongSwan's VICI (Versatile IKE Configuration Interface). The server runs as part of the TailSwan container and is accessible as a Tailscale app.

## Features

- **Web UI**: Modern, dark-themed interface for managing connections
- **REST API**: JSON-based API for programmatic control
- **Real-time Updates**: Automatic refresh of security associations
- **Embedded Static Files**: No external dependencies, all files embedded in binary

## Endpoints

### Health Check
**GET** `/health`

Check if the control server is running.

**Response:**
```json
{
  "success": true,
  "message": "TailSwan control server is healthy"
}
```

### Bring Connection Up
**POST** `/connections/up`

Initiate an IPsec connection.

**Request Body:**
```json
{
  "name": "connection-name"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Connection 'connection-name' initiated successfully"
}
```

**Response (Error):**
```json
{
  "success": false,
  "message": "Failed to initiate connection 'connection-name'",
  "error": "detailed error message"
}
```

### Bring Connection Down
**POST** `/connections/down`

Terminate an IPsec connection.

**Request Body:**
```json
{
  "name": "connection-name"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Connection 'connection-name' terminated successfully"
}
```

**Response (Error):**
```json
{
  "success": false,
  "message": "Failed to terminate connection 'connection-name'",
  "error": "detailed error message"
}
```

### List Connections
**GET** `/connections/list`

List all configured IPsec connections.

**Response:**
```json
{
  "success": true,
  "connections": [...]
}
```

### List Security Associations
**GET** `/sas/list`

List active security associations.

**Response:**
```json
{
  "success": true,
  "sas": [...]
}
```

## Configuration

The control server can be configured using environment variables:

- `CONTROL_PORT` - Port to listen on (default: `8080`)

## Usage Examples

### Using the Web UI

Access the web interface at:
```
http://localhost:8080/
```

Or via Tailscale:
```
http://tailswan:8080/
```

The web UI provides:
- **Server Status**: Real-time health check indicator
- **Configured Connections**: List all configured IPsec connections with quick action buttons
- **Active Security Associations**: View all active SAs with details
- **Manual Control**: Bring connections up or down by name
- **Auto-refresh**: Security associations automatically refresh every 10 seconds

### Using curl

Bring a connection up:
```bash
curl -X POST http://localhost:8080/connections/up \
  -H "Content-Type: application/json" \
  -d '{"name":"my-vpn"}'
```

Bring a connection down:
```bash
curl -X POST http://localhost:8080/connections/down \
  -H "Content-Type: application/json" \
  -d '{"name":"my-vpn"}'
```

List all connections:
```bash
curl http://localhost:8080/connections/list
```

List security associations:
```bash
curl http://localhost:8080/sas/list
```

### Tailscale App Access

When the TailSwan container is running and connected to Tailscale, you can access the control server from any device on your Tailnet:

```bash
curl -X POST http://tailswan:8080/connections/up \
  -H "Content-Type: application/json" \
  -d '{"name":"my-vpn"}'
```

## Dependencies

- [govici](https://pkg.go.dev/github.com/strongswan/govici/vici) - Go bindings for strongSwan's VICI interface

## Architecture

The control server:
1. Connects to strongSwan's VICI socket on startup
2. Exposes HTTP endpoints for connection management
3. Translates HTTP requests to VICI commands
4. Returns JSON responses with operation results

## Error Handling

All endpoints return appropriate HTTP status codes:
- `200 OK` - Successful operation
- `400 Bad Request` - Invalid request body or missing parameters
- `405 Method Not Allowed` - Wrong HTTP method
- `500 Internal Server Error` - VICI communication or command execution failed
