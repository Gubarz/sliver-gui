<script>
  import { createEventDispatcher, onMount } from 'svelte';
  
  export let options = [];
  export let value;
  export let disabled = false;
  export let id = undefined;
  
  const dispatch = createEventDispatcher();
  
  let isOpen = false;
  let container;
  
  function toggle() {
    if (disabled) return;
    isOpen = !isOpen;
  }
  
  function selectOption(opt) {
    value = opt;
    isOpen = false;
    dispatch('change', value);
  }
  
  function handleOutsideClick(event) {
    if (container && !container.contains(event.target)) {
      isOpen = false;
    }
  }
  
  onMount(() => {
    document.addEventListener('click', handleOutsideClick);
    return () => {
      document.removeEventListener('click', handleOutsideClick);
    };
  });
</script>

<div class="custom-select" bind:this={container} class:disabled>
  <button
    type="button"
    {id}
    class="selected-value"
    aria-haspopup="listbox"
    aria-expanded={isOpen}
    {disabled}
    on:click={toggle}
  >
    {value || ''}
    <span class="arrow" class:up={isOpen}>▼</span>
  </button>
  
  {#if isOpen}
    <div class="dropdown" role="listbox" aria-label="Options">
      {#each options as opt}
        <button
          type="button"
          class="option" 
          class:selected={value === opt}
          role="option"
          aria-selected={value === opt}
          on:click={() => selectOption(opt)}
        >
          {opt}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .custom-select {
    position: relative;
    width: 100%;
    user-select: none;
  }
  .custom-select.disabled {
    opacity: 0.6;
    pointer-events: none;
  }
  .selected-value {
    width: 100%;
    padding: 8px 12px;
    background-color: var(--toolbar-bg);
    border: 1px solid var(--panel-border);
    border-radius: 4px;
    color: var(--text-color);
    cursor: pointer;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font: inherit;
    text-align: left;
  }
  .selected-value:hover {
    border-color: var(--accent-color);
  }
  .arrow {
    font-size: 0.8em;
    transition: transform 0.2s;
  }
  .arrow.up {
    transform: rotate(180deg);
  }
  .dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: var(--panel-bg);
    border: 1px solid var(--panel-border);
    border-radius: 4px;
    margin-top: 4px;
    max-height: 200px;
    overflow-y: auto;
    z-index: 100;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5);
  }
  .dropdown::-webkit-scrollbar { width: 6px; }
  .dropdown::-webkit-scrollbar-track { background: transparent; }
  .dropdown::-webkit-scrollbar-thumb { background: #444; border-radius: 3px; }
  .option {
    display: block;
    width: 100%;
    border: 0;
    background: transparent;
    padding: 8px 12px;
    cursor: pointer;
    color: var(--text-color);
    font: inherit;
    text-align: left;
  }
  .option:hover {
    background-color: var(--row-hover);
  }
  .option.selected {
    background-color: var(--accent-color);
    color: var(--bg-color);
  }
</style>
