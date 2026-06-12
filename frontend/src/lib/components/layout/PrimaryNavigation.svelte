<script>
  export let active = 'agents';
  export let onSelect = () => {};

  const tabs = [
    { id: 'agents', label: 'Agents', icon: 'fa-bullseye' },
    { id: 'automation', label: 'Automation', icon: 'fa-bolt' },
    { id: 'server', label: 'Server', icon: 'fa-server' },
    { id: 'settings', label: 'Settings', icon: 'fa-cog' },
  ];
</script>

<nav class="primary-tabs" aria-label="Primary">
  {#each tabs as tab}
    <button
      type="button"
      class="ptab"
      class:active={active === tab.id}
      on:click={() => onSelect(tab.id)}
    >
      <i class="fas {tab.icon}"></i> {tab.label}
    </button>
  {/each}
</nav>
