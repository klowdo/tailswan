const API_BASE = '/api';

function tailswanApp() {
    return {
        currentTab: localStorage.getItem('tailswan_current_tab') || 'ipsec',
        serverOnline: false,
        tailscaleRunning: false,

        connections: [],
        sas: [],
        manualConnectionName: '',
        loadingConnections: {},

        nodeInfo: null,
        peers: [],
        serveConfig: null,

        notification: {
            show: false,
            message: '',
            type: 'info'
        },

        eventSource: null,
        reconnectDelay: 1000,
        maxReconnectDelay: 30000,

        init() {
            this.checkServerStatus();
            this.loadConnections();
            this.loadSAs();
            this.loadTailscaleStatus();
            this.loadTailscalePeers();
            this.loadTailscaleServe();
            this.connectSSE();
        },

        switchTab(tab) {
            this.currentTab = tab;
            localStorage.setItem('tailswan_current_tab', tab);
        },

        async checkServerStatus() {
            try {
                const response = await fetch(`${API_BASE}/health`);
                const data = await response.json();
                this.serverOnline = data.success;
            } catch (error) {
                this.serverOnline = false;
            }
        },

        async loadConnections() {
            try {
                const response = await fetch(`${API_BASE}/vici/connections/list`);
                const data = await response.json();
                this.updateConnections(data);
            } catch (error) {
                console.error('Error loading connections:', error);
            }
        },

        async loadSAs() {
            try {
                const response = await fetch(`${API_BASE}/vici/sas/list`);
                const data = await response.json();
                this.updateSAs(data);
            } catch (error) {
                console.error('Error loading SAs:', error);
            }
        },

        async loadTailscaleStatus() {
            try {
                const response = await fetch(`${API_BASE}/tailscale/status`);
                const data = await response.json();
                this.updateNodeInfo(data);
            } catch (error) {
                console.error('Error loading Tailscale status:', error);
            }
        },

        async loadTailscalePeers() {
            try {
                const response = await fetch(`${API_BASE}/tailscale/peers`);
                const data = await response.json();
                this.updatePeers(data);
            } catch (error) {
                console.error('Error loading peers:', error);
            }
        },

        async loadTailscaleServe() {
            try {
                const response = await fetch(`${API_BASE}/tailscale/serve`);
                const data = await response.json();
                if (data.success && data.config && Object.keys(data.config).length > 0) {
                    this.serveConfig = JSON.stringify(data.config, null, 2);
                } else {
                    this.serveConfig = null;
                }
            } catch (error) {
                console.error('Error loading serve config:', error);
            }
        },

        updateConnections(data) {
            if (!data.success || !data.connections) {
                this.connections = [];
                return;
            }

            this.connections = data.connections.map((conn, index) => {
                const connName = Object.keys(conn)[0] || `connection-${index}`;
                const connData = conn[connName];

                const details = [];
                if (connData?.local_addrs) {
                    details.push(`Local: ${Array.isArray(connData.local_addrs) ? connData.local_addrs.join(', ') : connData.local_addrs}`);
                }
                if (connData?.remote_addrs) {
                    details.push(`Remote: ${Array.isArray(connData.remote_addrs) ? connData.remote_addrs.join(', ') : connData.remote_addrs}`);
                }
                if (connData?.version) {
                    details.push(`IKE v${connData.version}`);
                }

                return {
                    name: connName,
                    details: details.length > 0 ? details.join(' • ') : 'Connection configured'
                };
            });
        },

        updateSAs(data) {
            if (!data.success || !data.sas) {
                this.sas = [];
                return;
            }

            this.sas = data.sas.map((sa, index) => {
                const saName = Object.keys(sa)[0] || `sa-${index}`;
                const saData = sa[saName];

                const details = [];
                if (saData?.state) {
                    details.push(`State: ${saData.state}`);
                }
                if (saData?.['local-host']) {
                    details.push(`Local: ${saData['local-host']}`);
                }
                if (saData?.['remote-host']) {
                    details.push(`Remote: ${saData['remote-host']}`);
                }

                return {
                    name: saName,
                    details: details.length > 0 ? details.join(' • ') : 'Active'
                };
            });
        },

        updateNodeInfo(data) {
            if (!data.success || !data.status) {
                this.nodeInfo = null;
                this.tailscaleRunning = false;
                return;
            }

            const status = data.status;
            const self = status.Self;

            this.tailscaleRunning = status.BackendState === 'Running';
            this.nodeInfo = {
                state: status.BackendState || 'Unknown',
                hostname: self?.HostName || 'N/A',
                ip: (self?.TailscaleIPs && self.TailscaleIPs[0]) || 'N/A',
                dns: self?.DNSName || 'N/A',
                online: self?.Online || false,
                os: self?.OS || 'N/A'
            };
        },

        updatePeers(data) {
            if (!data.success || !data.peers) {
                this.peers = [];
                return;
            }

            this.peers = data.peers.map(peer => {
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
                if (peer.last_seen && peer.last_seen !== '0001-01-01T00:00:00Z') {
                    const lastSeen = new Date(peer.last_seen);
                    details.push(`Last seen: ${lastSeen.toLocaleString()}`);
                }

                return {
                    id: peer.id,
                    hostname: peer.hostname || 'Unknown',
                    online: peer.online,
                    details: details.join(' • ')
                };
            });
        },

        async bringConnectionUp(name) {
            if (!name || name.trim() === '') {
                this.showNotification('Please enter a connection name', 'warning');
                return;
            }

            const connName = name.trim();
            this.loadingConnections[connName] = 'up';

            try {
                const response = await fetch(`${API_BASE}/vici/connections/up`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name: connName }),
                });

                const data = await response.json();

                if (data.success) {
                    this.showNotification(data.message, 'success');
                    this.manualConnectionName = '';
                    setTimeout(() => this.loadSAs(), 1000);
                } else {
                    this.showNotification(data.error || data.message, 'error');
                }
            } catch (error) {
                this.showNotification(`Error: ${error.message}`, 'error');
            } finally {
                this.loadingConnections[connName] = null;
            }
        },

        async bringConnectionDown(name) {
            if (!name || name.trim() === '') {
                this.showNotification('Please enter a connection name', 'warning');
                return;
            }

            const connName = name.trim();
            this.loadingConnections[connName] = 'down';

            try {
                const response = await fetch(`${API_BASE}/vici/connections/down`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name: connName }),
                });

                const data = await response.json();

                if (data.success) {
                    this.showNotification(data.message, 'success');
                    this.manualConnectionName = '';
                    setTimeout(() => this.loadSAs(), 1000);
                } else {
                    this.showNotification(data.error || data.message, 'error');
                }
            } catch (error) {
                this.showNotification(`Error: ${error.message}`, 'error');
            } finally {
                this.loadingConnections[connName] = null;
            }
        },

        refreshAll() {
            this.checkServerStatus();
            this.loadConnections();
            this.loadSAs();
            this.loadTailscaleStatus();
            this.loadTailscalePeers();
            this.loadTailscaleServe();
        },

        showNotification(message, type = 'info') {
            this.notification = {
                show: true,
                message: message,
                type: type
            };

            setTimeout(() => {
                this.notification.show = false;
            }, 10000);
        },

        connectSSE() {
            if (this.eventSource) {
                this.eventSource.close();
            }

            this.eventSource = new EventSource('/api/events');

            this.eventSource.addEventListener('sa-update', (e) => {
                const data = JSON.parse(e.data);
                this.updateSAs(data);
            });

            this.eventSource.addEventListener('peer-update', (e) => {
                const data = JSON.parse(e.data);
                this.updatePeers(data);
            });

            this.eventSource.addEventListener('connection-update', (e) => {
                const data = JSON.parse(e.data);
                this.updateConnections(data);
            });

            this.eventSource.addEventListener('node-update', (e) => {
                const data = JSON.parse(e.data);
                this.updateNodeInfo(data);
            });

            this.eventSource.onopen = () => {
                this.reconnectDelay = 1000;
            };

            this.eventSource.onerror = () => {
                this.eventSource.close();

                setTimeout(() => {
                    this.connectSSE();
                    this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
                }, this.reconnectDelay);
            };
        }
    };
}
