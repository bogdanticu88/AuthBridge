document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

const API_KEY = new URLSearchParams(window.location.search).get('api_key') || "";

async function secureFetch(url, options = {}) {
    if (API_KEY) {
        if (!options.headers) options.headers = {};
        options.headers['X-AuthBridge-Key'] = API_KEY;
    }
    return fetch(url, options);
}

function initApp() {
    // Navigation handling
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', () => {
            const view = item.getAttribute('data-view');
            switchView(view);
        });
    });

    // Form submission
    document.getElementById('add-cred-form').addEventListener('submit', handleAddCredential);

    // Initial load
    loadDashboardData();
}

function switchView(viewId) {
    document.querySelectorAll('.nav-item').forEach(i => i.classList.remove('active'));
    document.querySelector(`[data-view="${viewId}"]`).classList.add('active');

    document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
    document.getElementById(`${viewId}-view`).classList.add('active');

    if (viewId === 'credentials') loadCredentials();
    if (viewId === 'audit') loadAuditLogs();
}

async function loadDashboardData() {
    try {
        const [creds, audit] = await Promise.all([
            secureFetch('/api/v1/credentials').then(r => r.json()),
            secureFetch('/api/v1/audit?limit=5').then(r => r.json())
        ]);

        document.getElementById('stat-total-creds').textContent = creds.credentials.length;
        
        let totalUsage = 0;
        creds.credentials.forEach(c => totalUsage += c.usage_count);
        document.getElementById('stat-total-reqs').textContent = totalUsage;

        // Mini audit log on dashboard
        const miniLog = document.getElementById('mini-audit-log');
        miniLog.innerHTML = (audit.logs || []).map(log => `
            <div style="padding: 0.5rem 0; border-bottom: 1px solid rgba(255,255,255,0.05);">
                <span style="color: var(--accent-blue)">${log.action}</span> 
                <span style="color: var(--text-primary)">${log.credential_name}</span>
                <div style="font-size: 0.7rem; opacity: 0.6;">${log.timestamp}</div>
            </div>
        `).join('');

    } catch (err) {
        console.error('Dashboard load failed', err);
    }
}

async function loadCredentials() {
    const container = document.getElementById('credentials-container');
    container.innerHTML = '<div style="color: var(--text-secondary)">Scanning vault...</div>';

    try {
        const resp = await secureFetch('/api/v1/credentials');
        const data = await resp.json();
        
        container.innerHTML = (data.credentials || []).map(c => `
            <div class="cred-card">
                <span class="cred-type-badge">${c.type}</span>
                <div class="cred-title">
                    <i class="fas fa-file-shield" style="color: var(--accent-blue)"></i>
                    ${c.name}
                </div>
                <div class="cred-meta">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 0.5rem;">
                        <span>Invocations</span>
                        <span class="code-pill">${c.usage_count}</span>
                    </div>
                    <div style="display: flex; justify-content: space-between;">
                        <span>Security Level</span>
                        <span style="color: var(--accent-green)">AES-256</span>
                    </div>
                </div>
                <div style="margin-top: 1.5rem; display: flex; gap: 0.5rem;">
                    <button class="btn btn-secondary" style="flex: 1" onclick="deleteCredential('${c.name}')">Revoke</button>
                    <button class="btn btn-secondary" style="width: 40px"><i class="fas fa-ellipsis-v"></i></button>
                </div>
            </div>
        `).join('');
    } catch (err) {
        container.innerHTML = '<div style="color: var(--accent-red)">Failed to connect to vault.</div>';
    }
}

async function loadAuditLogs() {
    const table = document.getElementById('audit-full-table');
    table.innerHTML = '<tr><td colspan="5" style="text-align: center; padding: 2rem; color: var(--text-secondary);">Querying logs...</td></tr>';

    try {
        const resp = await secureFetch('/api/v1/audit?limit=100');
        const data = await resp.json();
        
        table.innerHTML = (data.logs || []).map(log => `
            <tr>
                <td style="font-family: var(--font-mono); font-size: 0.8rem; color: var(--text-secondary)">${log.timestamp}</td>
                <td style="font-weight: 600;">${log.action}</td>
                <td><span class="code-pill">${log.credential_name}</span></td>
                <td>${log.source_ip}</td>
                <td>
                    <span style="color: ${log.status === 'success' ? 'var(--accent-green)' : 'var(--accent-red)'}">
                        ${log.status.toUpperCase()}
                    </span>
                </td>
            </tr>
        `).join('');
    } catch (err) {
        table.innerHTML = '<tr><td colspan="5" style="color: var(--accent-red)">Audit stream interrupted.</td></tr>';
    }
}

async function handleAddCredential(e) {
    e.preventDefault();
    const btn = e.target.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Storing...';

    const data = {
        name: document.getElementById('cred-name').value,
        type: document.getElementById('cred-type').value,
        token: document.getElementById('cred-token').value
    };

    try {
        const resp = await secureFetch('/api/v1/credentials', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (resp.ok) {
            closeModal('add-modal');
            loadDashboardData();
            if (document.getElementById('credentials-view').classList.contains('active')) loadCredentials();
            e.target.reset();
        } else {
            const err = await resp.json();
            alert('Vault error: ' + err.error);
        }
    } catch (err) {
        alert('Connection failure');
    } finally {
        btn.disabled = false;
        btn.textContent = 'Create Object';
    }
}

async function deleteCredential(name) {
    if (!confirm(`Permanently revoke access for ${name}?`)) return;
    try {
        await secureFetch(`/api/v1/credentials/${name}`, { method: 'DELETE' });
        loadCredentials();
    } catch (err) {
        alert('Revocation failed');
    }
}

function openModal(id) {
    document.getElementById(id).style.display = 'flex';
}

function closeModal(id) {
    document.getElementById(id).style.display = 'none';
}
