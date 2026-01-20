const API_BASE = '/api';

let eventSource = null;
let reconnectDelay = 1000;
const maxReconnectDelay = 30000;
let currentTab = 'ipsec';

function switchTab(tab) {
    currentTab = tab;

    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
    });
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });

    event.target.classList.add('active');
    document.getElementById(`${tab}-tab`).classList.add('active');

    if (tab === 'tailscale') {
        loadTailscaleStatus();
        loadTailscalePeers();
        loadTailscaleServe();
    }
}

async function checkServerStatus() {
    try {
        const response = await fetch(`${API_BASE}/health`);
        const data = await response.json();

        const statusEl = document.getElementById('server-status');
        if (data.success) {
            statusEl.textContent = 'Online';
            statusEl.className = 'status-value online';
        } else {
            statusEl.textContent = 'Error';
            statusEl.className = 'status-value offline';
        }
    } catch (error) {
        const statusEl = document.getElementById('server-status');
        statusEl.textContent = 'Offline';
        statusEl.className = 'status-value offline';
    }
}

async function checkTailscaleStatus() {
    try {
        const response = await fetch(`${API_BASE}/tailscale/status`);
        const data = await response.json();

        const statusEl = document.getElementById('tailscale-status');
        if (data.success && data.status && data.status.BackendState === 'Running') {
            statusEl.textContent = 'Running';
            statusEl.className = 'status-value online';
        } else {
            statusEl.textContent = 'Not Running';
            statusEl.className = 'status-value offline';
        }
    } catch (error) {
        const statusEl = document.getElementById('tailscale-status');
        statusEl.textContent = 'Error';
        statusEl.className = 'status-value offline';
    }
}

async function loadConnections() {
    const container = document.getElementById('connections-list');
    container.innerHTML = '<div class="loading">Loading connections...</div>';

    try {
        const response = await fetch(`${API_BASE}/vici/connections/list`);
        const data = await response.json();
        updateConnectionsList(data);
    } catch (error) {
        const container = document.getElementById('connections-list');
        container.innerHTML = `<div class="empty-state">Error loading connections: ${escapeHtml(error.message)}</div>`;
    }
}

function updateConnectionsList(data) {
    const container = document.getElementById('connections-list');

    if (!data.success) {
        container.innerHTML = '<div class="empty-state">Failed to load connections</div>';
        return;
    }

    if (!data.connections || data.connections.length === 0) {
        container.innerHTML = '<div class="empty-state">No connections configured</div>';
        return;
    }

    container.innerHTML = data.connections.map((conn, index) => {
        const connName = Object.keys(conn)[0] || `connection-${index}`;
        return `
            <div class="connection-item">
                <div class="connection-info">
                    <div class="connection-name">${escapeHtml(connName)}</div>
                    <div class="connection-details">
                        ${getConnectionDetails(conn[connName])}
                    </div>
                </div>
                <div class="connection-actions">
                    <button onclick="bringConnectionUpByName('${escapeHtml(connName)}')" class="btn btn-success btn-sm">▲ Up</button>
                    <button onclick="bringConnectionDownByName('${escapeHtml(connName)}')" class="btn btn-danger btn-sm">▼ Down</button>
                </div>
            </div>
        `;
    }).join('');
}

async function loadSAs() {
    const container = document.getElementById('sas-list');
    container.innerHTML = '<div class="loading">Loading security associations...</div>';

    try {
        const response = await fetch(`${API_BASE}/vici/sas/list`);
        const data = await response.json();
        updateSAsList(data);
    } catch (error) {
        const container = document.getElementById('sas-list');
        container.innerHTML = `<div class="empty-state">Error loading SAs: ${escapeHtml(error.message)}</div>`;
    }
}

function updateSAsList(data) {
    const container = document.getElementById('sas-list');

    if (!data.success) {
        container.innerHTML = '<div class="empty-state">Failed to load security associations</div>';
        return;
    }

    if (!data.sas || data.sas.length === 0) {
        container.innerHTML = '<div class="empty-state">No active security associations</div>';
        return;
    }

    container.innerHTML = data.sas.map((sa, index) => {
        const saName = Object.keys(sa)[0] || `sa-${index}`;
        return `
            <div class="sa-item">
                <div class="sa-info">
                    <div class="sa-name">${escapeHtml(saName)}</div>
                    <div class="sa-details">
                        ${getSADetails(sa[saName])}
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

function getConnectionDetails(conn) {
    if (!conn) return 'No details available';

    const details = [];

    if (conn.local_addrs) {
        details.push(`Local: ${Array.isArray(conn.local_addrs) ? conn.local_addrs.join(', ') : conn.local_addrs}`);
    }
    if (conn.remote_addrs) {
        details.push(`Remote: ${Array.isArray(conn.remote_addrs) ? conn.remote_addrs.join(', ') : conn.remote_addrs}`);
    }
    if (conn.version) {
        details.push(`IKE v${conn.version}`);
    }

    return details.length > 0 ? details.join(' • ') : 'Connection configured';
}

function getSADetails(sa) {
    if (!sa) return 'No details available';

    const details = [];

    if (sa.state) {
        details.push(`State: ${sa.state}`);
    }
    if (sa['local-host']) {
        details.push(`Local: ${sa['local-host']}`);
    }
    if (sa['remote-host']) {
        details.push(`Remote: ${sa['remote-host']}`);
    }

    return details.length > 0 ? details.join(' • ') : 'Active';
}

async function bringConnectionUp() {
    const nameInput = document.getElementById('connection-name');
    const name = nameInput.value.trim();

    if (!name) {
        showNotification('Please enter a connection name', 'warning');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/vici/connections/up`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name }),
        });

        const data = await response.json();

        if (data.success) {
            showNotification(data.message, 'success');
            nameInput.value = '';
            setTimeout(() => {
                loadSAs();
            }, 1000);
        } else {
            showNotification(data.error || data.message, 'error');
        }
    } catch (error) {
        showNotification(`Error: ${error.message}`, 'error');
    }
}

async function bringConnectionDown() {
    const nameInput = document.getElementById('connection-name');
    const name = nameInput.value.trim();

    if (!name) {
        showNotification('Please enter a connection name', 'warning');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/vici/connections/down`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name }),
        });

        const data = await response.json();

        if (data.success) {
            showNotification(data.message, 'success');
            nameInput.value = '';
            setTimeout(() => {
                loadSAs();
            }, 1000);
        } else {
            showNotification(data.error || data.message, 'error');
        }
    } catch (error) {
        showNotification(`Error: ${error.message}`, 'error');
    }
}

async function bringConnectionUpByName(name) {
    try {
        const response = await fetch(`${API_BASE}/vici/connections/up`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name }),
        });

        const data = await response.json();

        if (data.success) {
            showNotification(data.message, 'success');
            setTimeout(() => {
                loadSAs();
            }, 1000);
        } else {
            showNotification(data.error || data.message, 'error');
        }
    } catch (error) {
        showNotification(`Error: ${error.message}`, 'error');
    }
}

async function bringConnectionDownByName(name) {
    try {
        const response = await fetch(`${API_BASE}/vici/connections/down`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name }),
        });

        const data = await response.json();

        if (data.success) {
            showNotification(data.message, 'success');
            setTimeout(() => {
                loadSAs();
            }, 1000);
        } else {
            showNotification(data.error || data.message, 'error');
        }
    } catch (error) {
        showNotification(`Error: ${error.message}`, 'error');
    }
}

function showNotification(message, type = 'info') {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = `notification ${type}`;

    setTimeout(() => {
        notification.classList.add('show');
    }, 10);

    setTimeout(() => {
        notification.classList.remove('show');
    }, 3000);
}

async function loadTailscaleStatus() {
    const container = document.getElementById('tailscale-info');
    container.innerHTML = '<div class="loading">Loading Tailscale status...</div>';

    try {
        const response = await fetch(`${API_BASE}/tailscale/status`);
        const data = await response.json();
        updateTailscaleStatus(data);
    } catch (error) {
        const container = document.getElementById('tailscale-info');
        container.innerHTML = `<div class="empty-state">Error loading Tailscale status: ${escapeHtml(error.message)}</div>`;
    }
}

function updateTailscaleStatus(data) {
    const container = document.getElementById('tailscale-info');

    if (!data.success || !data.status) {
        container.innerHTML = '<div class="empty-state">Failed to load Tailscale status</div>';
        return;
    }

    const status = data.status;
    const self = status.Self;

    container.innerHTML = `
        <div class="info-item">
            <div class="info-label">State</div>
            <div class="info-value">${escapeHtml(status.BackendState || 'Unknown')}</div>
        </div>
        <div class="info-item">
            <div class="info-label">Hostname</div>
            <div class="info-value">${escapeHtml(self.HostName || 'N/A')}</div>
        </div>
        <div class="info-item">
            <div class="info-label">Tailscale IP</div>
            <div class="info-value">${escapeHtml(self.TailscaleIPs && self.TailscaleIPs[0] || 'N/A')}</div>
        </div>
        <div class="info-item">
            <div class="info-label">DNS Name</div>
            <div class="info-value">${escapeHtml(self.DNSName || 'N/A')}</div>
        </div>
        <div class="info-item">
            <div class="info-label">Online</div>
            <div class="info-value">${self.Online ? '✓ Yes' : '✗ No'}</div>
        </div>
        <div class="info-item">
            <div class="info-label">OS</div>
            <div class="info-value">${escapeHtml(self.OS || 'N/A')}</div>
        </div>
    `;
}

async function loadTailscalePeers() {
    const container = document.getElementById('tailscale-peers');
    container.innerHTML = '<div class="loading">Loading peers...</div>';

    try {
        const response = await fetch(`${API_BASE}/tailscale/peers`);
        const data = await response.json();
        updateTailscalePeers(data);
    } catch (error) {
        const container = document.getElementById('tailscale-peers');
        container.innerHTML = `<div class="empty-state">Error loading peers: ${escapeHtml(error.message)}</div>`;
    }
}

function updateTailscalePeers(data) {
    const container = document.getElementById('tailscale-peers');

    if (!data.success) {
        container.innerHTML = '<div class="empty-state">Failed to load peers</div>';
        return;
    }

    if (!data.peers || data.peers.length === 0) {
        container.innerHTML = '<div class="empty-state">No peers found</div>';
        return;
    }

    container.innerHTML = data.peers.map(peer => {
        const statusClass = peer.online ? 'online' : 'offline';
        const statusText = peer.online ? 'Online' : 'Offline';

        const details = [];
        if (peer.tailscale_ips && peer.tailscale_ips.length > 0) {
            details.push(`IP: ${peer.tailscale_ips[0]}`);
        }
        if (peer.os) {
            details.push(`OS: ${peer.os}`);
        }
        if (peer.dns_name) {
            details.push(`DNS: ${peer.dns_name}`);
        }
        if (peer.last_seen) {
            const lastSeen = new Date(peer.last_seen);
            details.push(`Last seen: ${lastSeen.toLocaleString()}`);
        }

        return `
            <div class="peer-item">
                <div class="peer-header">
                    <div class="peer-name">${escapeHtml(peer.hostname || 'Unknown')}</div>
                    <div class="peer-status ${statusClass}">${statusText}</div>
                </div>
                <div class="peer-details">${escapeHtml(details.join(' • '))}</div>
            </div>
        `;
    }).join('');
}

async function loadTailscaleServe() {
    const container = document.getElementById('tailscale-serve');
    container.innerHTML = '<div class="loading">Loading serve config...</div>';

    try {
        const response = await fetch(`${API_BASE}/tailscale/serve`);
        const data = await response.json();

        if (!data.success) {
            throw new Error('Failed to load serve configuration');
        }

        if (!data.config || Object.keys(data.config).length === 0) {
            container.innerHTML = '<div class="empty-state">No Tailscale Serve configuration active</div>';
            return;
        }

        container.innerHTML = `<pre class="serve-config">${escapeHtml(JSON.stringify(data.config, null, 2))}</pre>`;
    } catch (error) {
        container.innerHTML = `<div class="empty-state">Error loading serve config: ${escapeHtml(error.message)}</div>`;
    }
}

function refreshAll() {
    checkServerStatus();
    checkTailscaleStatus();

    if (currentTab === 'ipsec') {
        loadConnections();
        loadSAs();
    } else if (currentTab === 'tailscale') {
        loadTailscaleStatus();
        loadTailscalePeers();
        loadTailscaleServe();
    }
}

function connectSSE() {
    if (eventSource) {
        eventSource.close();
    }

    eventSource = new EventSource('/api/events');

    eventSource.addEventListener('sa-update', (e) => {
        const data = JSON.parse(e.data);
        updateSAsList(data);
    });

    eventSource.addEventListener('peer-update', (e) => {
        const data = JSON.parse(e.data);
        updateTailscalePeers(data);
    });

    eventSource.addEventListener('connection-update', (e) => {
        const data = JSON.parse(e.data);
        updateConnectionsList(data);
    });

    eventSource.addEventListener('node-update', (e) => {
        const data = JSON.parse(e.data);
        updateTailscaleStatus(data);
    });

    eventSource.onopen = () => {
        reconnectDelay = 1000;
        console.log('SSE connected');
    };

    eventSource.onerror = (e) => {
        console.error('SSE error:', e);
        eventSource.close();

        setTimeout(() => {
            connectSSE();
            reconnectDelay = Math.min(reconnectDelay * 2, maxReconnectDelay);
        }, reconnectDelay);
    };
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

document.getElementById('connection-name').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        bringConnectionUp();
    }
});

document.addEventListener('DOMContentLoaded', () => {
    refreshAll();
    connectSSE();
});
