let lastUpdate = null;

async function refresh() {
  try {
    const resp = await fetch('/api/status');
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const data = await resp.json();
    render(data);
    lastUpdate = new Date();
  } catch (err) {
    console.error('refresh error:', err);
  }
}

function render(d) {
  // Header
  setText('project-name', d.project?.name || '—');

  // Repository
  setText('repo-name', d.project?.name || '—');
  const remote = d.project?.remote || '—';
  setText('repo-remote', remote);
  setText('repo-branch', d.git?.branch || '—');
  setText('repo-tree', d.git?.isClean ? '✓ clean' : '● dirty');
  el('repo-tree').style.color = d.git?.isClean ? 'var(--green)' : 'var(--yellow)';
  const ab = `+${d.git?.ahead ?? 0} / -${d.git?.behind ?? 0}`;
  setText('repo-ahead', ab);
  setText('repo-tag', d.git?.latestTag || '(none)');

  // PR
  const prDiv = el('pr-content');
  if (d.pr) {
    const ciColor = ciStatusColor(d.pr.ci);
    prDiv.innerHTML = `
      <div class="pr-title">#${d.pr.number} — ${esc(d.pr.title)}</div>
      <table class="kv">
        <tr><td>state</td><td>${badge(d.pr.state, 'blue')}</td></tr>
        <tr><td>CI</td><td>${badge(d.pr.ci, ciColor)}</td></tr>
        <tr><td>mergeable</td><td>${badge(d.pr.mergeable, mergeableColor(d.pr.mergeable))}</td></tr>
      </table>
      <div class="pr-url" style="margin-top:8px"><a href="${d.pr.url}" target="_blank">↗ open PR</a></div>`;
  } else {
    prDiv.innerHTML = '<p class="muted">No open PR for current branch</p>';
  }

  // Release
  const relDiv = el('release-content');
  if (d.release) {
    relDiv.innerHTML = `
      <table class="kv">
        <tr><td>current</td><td>${d.release.current || '(none)'}</td></tr>
        ${d.release.proposed ? `<tr><td>proposed</td><td style="color:var(--green)">${d.release.proposed}</td></tr>` : ''}
        ${d.release.prNumber ? `<tr><td>release PR</td><td><a href="${d.release.prUrl}" target="_blank" style="color:var(--accent)">#${d.release.prNumber}</a></td></tr>` : ''}
        ${d.release.ci ? `<tr><td>CI</td><td>${badge(d.release.ci, ciStatusColor(d.release.ci))}</td></tr>` : ''}
        <tr><td>status</td><td>${badge(d.release.status, 'green')}</td></tr>
      </table>`;
  } else {
    relDiv.innerHTML = '<p class="muted">No release data available</p>';
  }

  // Pipeline
  if (d.pipeline) {
    for (const stage of d.pipeline) {
      const stageId = `stage-${stage.name.replace(/\s+/g, '-').replace(/_/g, '-')}`;
      const stageEl = el(stageId);
      if (!stageEl) continue;
      stageEl.className = `stage stage-${stage.status}`;
      const statusEl = stageEl.querySelector('.stage-status');
      if (statusEl) statusEl.textContent = stage.status;
    }
  }

  // hcflow info
  setText('hcflow-schema', d.hcflow ? `v${d.hcflow.schema}` : '—');
  setText('hcflow-workflow', d.hcflow?.workflowVersion || '—');
  setText('hcflow-updated', d.updatedAt ? new Date(d.updatedAt).toLocaleTimeString() : '—');
}

function el(id) { return document.getElementById(id); }
function setText(id, text) { const e = el(id); if (e) e.textContent = text; }
function esc(s) { return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

function badge(text, color) {
  return `<span class="badge badge-${color}">${esc(text || '—')}</span>`;
}

function ciStatusColor(ci) {
  switch (ci) {
    case 'passed': return 'green';
    case 'running': return 'yellow';
    case 'failed': return 'red';
    default: return 'muted';
  }
}

function mergeableColor(m) {
  if (m === 'MERGEABLE') return 'green';
  if (m === 'CONFLICTING') return 'red';
  return 'muted';
}

// Auto-refresh every 30s
refresh();
setInterval(refresh, 30000);
