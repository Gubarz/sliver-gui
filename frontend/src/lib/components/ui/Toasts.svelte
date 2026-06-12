<script>
  import { onMount, onDestroy } from 'svelte';
  import { onSliverEvent } from '../../api/runtime.js';

  let toasts = [];
  let nextId = 1;

  // Map raw event types to a friendly message + severity.
  function describe(ev) {
    const host = ev.hostname ? ` ${ev.hostname}` : '';
    const who = ev.username ? ` (${ev.username})` : '';
    switch (ev.type) {
      case 'session-connected':
        return { kind: 'good', text: `Session opened:${host}${who}` };
      case 'session-disconnected':
        return { kind: 'warn', text: `Session closed:${host}` };
      case 'session-updated':
        return null; // too noisy
      case 'beacon-registered':
        return { kind: 'good', text: `Beacon registered:${host}${who}` };
      case 'job-started':
        return { kind: 'info', text: `Job started${ev.job ? ': ' + ev.job : ''}` };
      case 'job-stopped':
        return { kind: 'warn', text: `Job stopped${ev.job ? ': ' + ev.job : ''}` };
      default:
        return null; // ignore unmapped/noisy events
    }
  }

  function push(ev) {
    const d = describe(ev);
    if (!d) return;
    const id = nextId++;
    toasts = [...toasts, { id, ...d }];
    setTimeout(() => { toasts = toasts.filter((t) => t.id !== id); }, 6000);
  }

  let unlisten;
  onMount(() => {
    unlisten = onSliverEvent(push);
  });
  onDestroy(() => { if (unlisten) unlisten(); });
</script>

<div class="toast-stack">
  {#each toasts as t (t.id)}
    <div class="toast {t.kind}">
      <i class="fas {t.kind === 'good' ? 'fa-circle-check' : t.kind === 'warn' ? 'fa-triangle-exclamation' : 'fa-circle-info'}"></i>
      <span>{t.text}</span>
    </div>
  {/each}
</div>

<style>
  .toast-stack {
    position: fixed; bottom: 35px; right: 16px;
    display: flex; flex-direction: column; gap: 8px;
    z-index: 10000; pointer-events: none;
  }
  .toast {
    display: flex; align-items: center; gap: 10px;
    background: var(--panel-bg); border: 1px solid var(--border-color);
    border-left-width: 3px;
    padding: 10px 16px; border-radius: 4px;
    box-shadow: 0 4px 14px rgba(0,0,0,0.45);
    font-size: 0.88em; color: var(--text-color);
    min-width: 220px; animation: slideIn 0.2s ease-out;
  }
  .toast.good { border-left-color: var(--success-color); }
  .toast.good i { color: var(--success-color); }
  .toast.warn { border-left-color: #d6a23e; }
  .toast.warn i { color: #d6a23e; }
  .toast.info { border-left-color: var(--accent-color); }
  .toast.info i { color: var(--accent-color); }
  @keyframes slideIn { from { opacity: 0; transform: translateX(20px); } to { opacity: 1; transform: translateX(0); } }
</style>
