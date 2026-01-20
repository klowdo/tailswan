const API_BASE = '/api';

let autoRefreshInterval = null;

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

async function loadConnections() {
    const container = document.getElementById('connections-list');
    container.innerHTML = '<div class="loading">Loading connections...</div>';

    try {
        const response = await fetch(`${API_BASE}/vici/connections/list`);
        const data = await response.json();

        if (!data.success) {
            throw new Error('Failed to load connections');
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
    } catch (error) {
        container.innerHTML = `<div class="empty-state">Error loading connections: ${escapeHtml(error.message)}</div>`;
    }
}

async function loadSAs() {
    const container = document.getElementById('sas-list');
    container.innerHTML = '<div class="loading">Loading security associations...</div>';

    try {
        const response = await fetch(`${API_BASE}/vici/sas/list`);
        const data = await response.json();

        if (!data.success) {
            throw new Error('Failed to load security associations');
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
    } catch (error) {
        container.innerHTML = `<div class="empty-state">Error loading SAs: ${escapeHtml(error.message)}</div>`;
    }
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

function refreshAll() {
    checkServerStatus();
    loadConnections();
    loadSAs();
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

    autoRefreshInterval = setInterval(() => {
        loadSAs();
    }, 10000);
});
