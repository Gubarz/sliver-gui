<script>
  const themes = [
    { id: 'dark', name: 'Dark', bg: '#1a1a1a', accent: '#4CAF50' },
    { id: 'light', name: 'Light', bg: '#f5f5f5', accent: '#2e7d32' },
    { id: 'hacker', name: 'Hacker', bg: '#000000', accent: '#00ff00' },
    { id: 'cyberpunk', name: 'Cyberpunk', bg: '#0d0221', accent: '#00f0ff' },
    { id: 'dracula', name: 'Dracula', bg: '#282a36', accent: '#bd93f9' },
  ];

  let current = document.documentElement.getAttribute('data-theme') || 'dark';

  function setTheme(id) {
    current = id;
    document.documentElement.setAttribute('data-theme', id);
    try { localStorage.setItem('sliver-theme', id); } catch { /* ignore */ }
  }
</script>

<div class="settings">
  <div class="card">
    <h3><i class="fas fa-palette"></i> Theme</h3>
    <p class="hint">Click a theme to apply it instantly. Your choice is remembered.</p>
    <div class="theme-grid">
      {#each themes as t}
        <button type="button" class="theme-card" class:active={current === t.id} on:click={() => setTheme(t.id)} style="background:{t.bg}">
          <div class="theme-dot" style="background:{t.accent}"></div>
          <span style="color:{t.id === 'light' ? '#333' : '#eee'}">{t.name}</span>
          {#if current === t.id}<i class="fas fa-check chk" style="color:{t.accent}"></i>{/if}
        </button>
      {/each}
    </div>
  </div>

  <div class="card">
    <h3><i class="fas fa-circle-info"></i> About</h3>
    <p class="hint">A Sliver C2 graphical client. Built with Wails + Svelte.</p>
  </div>
</div>

<style>
  .settings { height: 100%; overflow: auto; padding: 24px; display: flex; flex-direction: column; gap: 18px; max-width: 720px; }
  .card { background: var(--panel-bg); border: 1px solid var(--panel-border); border-radius: 8px; padding: 18px 20px; }
  .card h3 { margin: 0 0 4px; color: var(--text-color); font-size: 1.05em; }
  .card h3 i { color: var(--accent-color); margin-right: 6px; }
  .hint { color: var(--text-muted); font-size: 0.85em; margin: 4px 0 14px; }
  .theme-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); gap: 12px; }
  .theme-card {
    position: relative; height: 64px; border-radius: 6px; border: 2px solid transparent;
    display: flex; align-items: center; gap: 10px; padding: 0 14px; cursor: pointer;
    color: inherit; font: inherit; text-align: left;
    transition: transform 0.1s ease;
  }
  .theme-card:hover { transform: translateY(-2px); }
  .theme-card.active { border-color: #fff; }
  .theme-dot { width: 16px; height: 16px; border-radius: 50%; }
  .theme-card .chk { position: absolute; top: 6px; right: 8px; }
</style>
