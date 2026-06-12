<script>
  import { createEventDispatcher, tick } from 'svelte';
  import { commandTooltip } from '../../utils/text.js';
  export let x = 0;
  export let y = 0;
  export let categories = [];
  
  let menuElement;
  let menuX = x;
  let menuY = y;
  
  $: {
    if (menuElement) {
      const rect = menuElement.getBoundingClientRect();
      const winW = window.innerWidth;
      const winH = window.innerHeight;
      
      menuX = x + rect.width > winW ? winW - rect.width - 10 : x;
      menuY = y + rect.height > winH ? winH - rect.height - 10 : y;
      
      if (menuX < 0) menuX = 10;
      if (menuY < 0) menuY = 10;
    }
  }
  
  export let coreActions = [
    { id: 'console', icon: 'fas fa-terminal', label: 'Console' },
    { id: 'fileBrowser', icon: 'fas fa-folder', label: 'File Browser' },
    { id: 'processExplorer', icon: 'fas fa-microchip', label: 'Process Explorer' },
    { id: 'screenshot', icon: 'fas fa-desktop', label: 'Take Screenshot' }
  ];

  export let footerActions = [
    { id: 'rename', icon: 'fas fa-pen', label: 'Rename' }
  ];

  export let dangerActions = [
    { id: 'kill', icon: 'fas fa-skull', label: 'Kill Agent' }
  ];

  export let onSelect = null;

  const dispatch = createEventDispatcher();
  
  let activeSubmenu = null;
  let submenuStyle = '';

  function handleAction(action) {
    if (onSelect) onSelect({ detail: { type: 'core', action } });
    dispatch('select', { type: 'core', action });
  }

  function handleCommand(command) {
    if (onSelect) onSelect({ detail: { type: 'command', command } });
    dispatch('select', { type: 'command', command });
  }

  async function openSubmenu(categoryName, event) {
    activeSubmenu = categoryName;
    // Wait for the submenu DOM to render, then position it
    await tick();
    const trigger = event.currentTarget;
    const triggerRect = trigger.getBoundingClientRect();
    const submenuEl = trigger.parentElement.querySelector('.submenu');
    if (!submenuEl) return;

    const subRect = submenuEl.getBoundingClientRect();
    const winW = window.innerWidth;
    const winH = window.innerHeight;

    // Horizontal: prefer right of trigger, flip left if clipped
    let sx = triggerRect.right + 2;
    if (sx + subRect.width > winW) {
      sx = triggerRect.left - subRect.width - 2;
    }
    if (sx < 0) sx = 4;

    // Vertical: align top with trigger, shift up if clipped at bottom
    let sy = triggerRect.top;
    if (sy + subRect.height > winH) {
      sy = winH - subRect.height - 4;
    }
    if (sy < 0) sy = 4;

    submenuStyle = `top: ${sy}px; left: ${sx}px;`;
  }
</script>

<div bind:this={menuElement} class="context-menu" role="menu" tabindex="-1" style="top: {menuY}px; left: {menuX}px;" on:mouseleave={() => activeSubmenu = null}>
  
  <!-- Dynamic Core Actions -->
  {#each coreActions as action}
    <button type="button" class="menu-item" role="menuitem" on:click={() => handleAction(action.id)} on:mouseenter={() => activeSubmenu = null}>
      <i class="{action.icon} icon"></i> {action.label}
    </button>
  {/each}

  {#if categories.length > 0 || footerActions.length > 0 || dangerActions.length > 0}
    <div class="menu-separator"></div>
  {/if}

  <!-- Dynamic Categories / Submenus -->
  {#each categories as category}
    <div class="submenu-group">
      <button
        type="button"
        class="menu-item has-submenu"
        role="menuitem"
        aria-haspopup="menu"
        aria-expanded={activeSubmenu === category.category}
        on:mouseenter={(e) => openSubmenu(category.category, e)}
        on:focus={(e) => openSubmenu(category.category, e)}
      >
      {category.category}
      <i class="fas fa-chevron-right submenu-arrow"></i>
      </button>
      {#if activeSubmenu === category.category}
        <div class="submenu" role="menu" tabindex="-1" style={submenuStyle}>
          {#each category.commands as command}
            <button
              type="button"
              class="menu-item"
              class:unavailable={!command.supported}
              role="menuitem"
              on:click|stopPropagation={() => handleCommand(command)}
              title={commandTooltip(command)}
            >
              {command.name}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/each}

  {#if footerActions.length > 0 || dangerActions.length > 0}
    <div class="menu-separator"></div>
  {/if}

  <!-- Dynamic Footer Actions -->
  {#each footerActions as action}
    <button type="button" class="menu-item" role="menuitem" on:click={() => handleAction(action.id)} on:mouseenter={() => activeSubmenu = null}>
      <i class="{action.icon} icon"></i> {action.label}
    </button>
  {/each}

  <!-- Dynamic Danger Actions -->
  {#each dangerActions as action}
    <button type="button" class="menu-item danger" role="menuitem" on:click={() => handleAction(action.id)} on:mouseenter={() => activeSubmenu = null}>
      <i class="{action.icon} icon"></i> {action.label}
    </button>
  {/each}

</div>

<style>
  .context-menu,
  .submenu {
    position: fixed;
    background-color: var(--panel-bg);
    border: 1px solid var(--panel-border);
    box-shadow: 0 4px 12px rgba(0,0,0,0.5);
    border-radius: 4px;
    padding: 5px 0;
    overflow-y: auto;
  }

  .context-menu {
    min-width: 220px;
    max-height: 90vh;
    z-index: 10000;
  }
  
  .context-menu::-webkit-scrollbar, .submenu::-webkit-scrollbar { width: 6px; }
  .context-menu::-webkit-scrollbar-track, .submenu::-webkit-scrollbar-track { background: transparent; }
  .context-menu::-webkit-scrollbar-thumb, .submenu::-webkit-scrollbar-thumb { background: #444; border-radius: 3px; }

  .menu-item {
    width: 100%;
    border: 0;
    background: transparent;
    position: relative;
    padding: 8px 15px;
    font-size: 0.9em;
    cursor: pointer;
    color: var(--text-color);
    display: flex;
    align-items: center;
    font: inherit;
    text-align: left;
  }
  .menu-item .icon {
    width: 20px;
    opacity: 0.7;
    margin-right: 8px;
    text-align: center;
  }
  .menu-item:hover {
    background-color: var(--accent-color);
    color: #fff;
  }

  .has-submenu {
    justify-content: space-between;
  }
  .submenu-arrow {
    font-size: 0.7em;
    opacity: 0.5;
    margin-left: auto;
    padding-left: 10px;
  }

  .submenu-group {
    position: relative;
  }

  /* Submenu uses position:fixed so it escapes the parent's overflow clipping */
  .submenu {
    min-width: 200px;
    z-index: 10001;
    max-height: 80vh;
  }
  .submenu .menu-item {
    padding: 6px 15px;
  }

  .menu-separator {
    height: 1px;
    background-color: var(--panel-border);
    margin: 4px 0;
  }
  .danger {
    color: #ff4a4a;
  }
  .menu-item.unavailable {
    color: var(--text-muted);
    font-style: italic;
  }
  .danger:hover {
    background-color: #ff4a4a;
    color: #fff;
  }
</style>
