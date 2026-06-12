<script>
  import { createEventDispatcher } from 'svelte';

  export let title;
  export let icon = '';
  export let embedded = false;
  export let showHeader = true;
  export let width = '760px';
  export let height = 'auto';
  export let bodyPadding = '0 20px 16px';
  export let bodyBackground = 'transparent';

  const dispatch = createEventDispatcher();

  function close() {
    if (!embedded) dispatch('close');
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div
  class="panel-layer"
  class:embedded
  style="--panel-width: {width}; --panel-height: {height}; --panel-body-padding: {bodyPadding}; --panel-body-background: {bodyBackground};"
>
  {#if !embedded}
    <button
      type="button"
      class="backdrop-dismiss"
      aria-label="Close {title}"
      on:click={close}
    ></button>
  {/if}

  <section
    class="panel"
    role={embedded ? undefined : 'dialog'}
    aria-modal={embedded ? undefined : 'true'}
    aria-label={title}
  >
    {#if showHeader}
      <header>
        <h3>
          {#if icon && !embedded}<i class="fas {icon}"></i>{/if}
          {title}
        </h3>
        <div class="header-actions">
          <slot name="actions"></slot>
          {#if !embedded}
            <button type="button" class="close-btn" aria-label="Close {title}" on:click={close}>×</button>
          {/if}
        </div>
      </header>
    {/if}

    <div class="body">
      <slot></slot>
    </div>
  </section>
</div>

<style>
  .panel-layer {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px;
  }

  .backdrop-dismiss {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    padding: 0;
    border: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    cursor: default;
  }

  .panel {
    position: relative;
    display: flex;
    box-sizing: border-box;
    width: var(--panel-width);
    height: var(--panel-height);
    min-width: 0;
    max-width: 95%;
    max-height: 85vh;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    background: var(--panel-bg);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 20px;
    border-bottom: 1px solid var(--border-color);
    background: var(--header-bg);
  }

  h3 {
    margin: 0;
    color: var(--text-color);
    font-size: 1.1em;
  }

  h3 i {
    margin-right: 6px;
    color: var(--accent-color);
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .close-btn {
    padding: 0;
    border: 0;
    background: none;
    color: var(--text-muted);
    font-size: 1.5em;
    line-height: 1;
    cursor: pointer;
  }

  .close-btn:hover {
    color: var(--text-color);
  }

  .body {
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: auto;
    padding: var(--panel-body-padding);
    background: var(--panel-body-background);
  }

  .panel-layer.embedded {
    position: static;
    z-index: auto;
    display: block;
    height: 100%;
    padding: 0;
  }

  .embedded .panel {
    width: 100%;
    height: 100%;
    max-width: none;
    max-height: none;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .embedded header {
    padding: 16px 20px;
    border-bottom: 0;
    background: transparent;
  }

  .embedded h3 {
    color: var(--text-color);
    font-size: 1.2em;
  }
</style>
