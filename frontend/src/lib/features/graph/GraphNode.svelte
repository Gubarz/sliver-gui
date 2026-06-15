<script>
  import { Handle, Position } from '@xyflow/svelte';
  let { data, selected = false } = $props();
</script>

{#if data.variant === 'agent'}
  <div class="node agent {data.kind}" class:dead={data.dead} class:selected>
    <Handle type="target" position={Position.Top} class="handle" />
    <div class="head">
      <i class={data.icon}></i>
      <span class="name" title={data.agentID}>{data.agentID}</span>
      <span class="dot" class:alive={!data.dead}></span>
    </div>
    <div class="implant" title={data.implantName}>{data.implantName}</div>
    <div class="sub">{data.user}@{data.host}</div>
    <div class="foot">
      <span class="badge {data.kind}">{data.kind}</span>
      <span class="addr">{data.addr || ''}</span>
    </div>
    <Handle type="source" position={Position.Bottom} class="handle" />
  </div>
{:else if data.variant === 'listener'}
  <div class="node listener">
    <Handle type="target" position={Position.Top} class="handle" />
    <i class="fas fa-satellite-dish"></i>
    <span title={data.label}>{data.label}</span>
    <Handle type="source" position={Position.Bottom} class="handle" />
  </div>
{:else if data.variant === 'device'}
  <div class="node device" class:selected>
    <Handle type="target" position={Position.Top} class="handle" />
    <div class="head">
      <i class="fas fa-laptop"></i>
      <span class="name" title={data.ip}>{data.ip}</span>
      <span class="dot discovered"></span>
    </div>
    <div class="implant" title={data.hostname || 'Unknown hostname'}>
      {data.hostname || 'Unknown hostname'}
    </div>
    <div class="sub" title={`Vendor: ${data.vendor || 'unknown'} / OS: ${data.osHint || 'unknown'}`}>
      {data.vendor ? `Vendor: ${data.vendor}` : 'Vendor unknown'} · {data.osHint || 'OS unknown'}
    </div>
    <div class="foot">
      <span class="badge discovered">{data.method || 'discovered'}</span>
      <span class="addr" title={data.mac || 'MAC unavailable'}>{data.mac || 'MAC unavailable'}</span>
    </div>
  </div>
{:else}
  <div class="node server">
    <i class="fas fa-server"></i>
    <span>{data.label}</span>
    <Handle type="source" position={Position.Bottom} class="handle" />
  </div>
{/if}

<style>
  .node {
    box-sizing: border-box;
    font-family: var(--font-mono);
    border-radius: 6px;
    border: 1px solid var(--border-color);
    background: var(--panel-bg);
    color: var(--text-color);
    box-shadow: 0 2px 8px rgba(0,0,0,0.35);
    cursor: pointer;
  }
  .node :global(.handle) { opacity: 0; }

  /* Server */
  .server {
    width: 180px;
    height: 44px;
    display: flex; align-items: center; gap: 8px;
    padding: 10px 14px; font-weight: bold;
    border-color: var(--accent-color);
  }
  .server i { color: var(--accent-color); }

  /* Listener */
  .listener {
    width: 220px;
    height: 40px;
    display: flex; align-items: center; gap: 8px;
    padding: 7px 12px; font-size: 0.85em;
    border-style: dashed;
  }
  .listener i { color: var(--text-muted); }
  .listener span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

  /* Discovered device */
  .device {
    width: 210px;
    height: 88px;
    padding: 8px 10px;
    border-left: 3px solid #58a6ff;
    cursor: default;
  }
  .device .head i { color: #58a6ff; }
  .device.selected {
    border-color: #9ecbff;
    box-shadow: 0 0 0 2px rgba(88,166,255,0.45), 0 2px 8px rgba(0,0,0,0.35);
    background: rgba(88,166,255,0.12);
  }
  /* Agent card */
  .agent { width: 210px; height: 88px; padding: 8px 10px; }
  .agent.session { border-left: 3px solid var(--success-color); }
  .agent.beacon  { border-left: 3px solid #d6a23e; }
  .agent.selected {
    border-color: var(--accent-color);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-color) 48%, transparent), 0 2px 8px rgba(0,0,0,0.35);
    background: color-mix(in srgb, var(--accent-color) 12%, var(--panel-bg));
  }
  .agent.dead { opacity: 0.72; filter: grayscale(0.35); }
  .head { display: flex; align-items: center; gap: 7px; }
  .head i { font-size: 1.05em; }
  .head .name { font-weight: bold; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .implant { font-size: 0.78em; margin-top: 3px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--danger-color); flex-shrink: 0; }
  .dot.alive { background: var(--success-color); box-shadow: 0 0 6px var(--success-color); }
  .dot.discovered { background: #58a6ff; box-shadow: 0 0 6px #58a6ff; }
  .sub { font-size: 0.78em; color: var(--text-muted); margin: 3px 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .foot { display: flex; align-items: center; justify-content: space-between; gap: 6px; }
  .badge { font-size: 0.7em; padding: 1px 6px; border-radius: 8px; text-transform: uppercase; }
  .badge.session { background: rgba(76,175,80,0.18); color: var(--success-color); }
  .badge.beacon { background: rgba(214,162,62,0.18); color: #d6a23e; }
  .badge.discovered { background: rgba(88,166,255,0.18); color: #58a6ff; }
  .addr { font-size: 0.72em; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
