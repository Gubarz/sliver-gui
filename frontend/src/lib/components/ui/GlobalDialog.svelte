<script>
  import { tick } from 'svelte';
  import { dialogStore } from '../../stores/dialogStore.js';
  import Modal from './Modal.svelte';

  let s;
  let promptInput;
  dialogStore.subscribe(value => {
    s = value;
    if (value.isOpen && value.type === 'prompt') {
      tick().then(() => promptInput?.focus());
    }
  });

  function close(result) {
    if (s.resolve) s.resolve(result);
    dialogStore.update(state => ({ ...state, isOpen: false }));
  }

  function handleCancel() {
    close(s.type === 'prompt' ? null : false);
  }

  function handleConfirm() {
    close(s.type === 'prompt' ? s.inputValue : true);
  }

  function handleKeydown(e) {
    if (s.isOpen) {
      if (e.key === 'Enter') handleConfirm();
      if (e.key === 'Escape') handleCancel();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown}/>

{#if s && s.isOpen}
  <Modal title={s.title} on:close={handleCancel}>
    <div style="margin-bottom: 20px; color: var(--text-color); font-size: 1.1em; white-space: pre-wrap; word-wrap: break-word;">
      {s.message}
    </div>

    {#if s.type === 'prompt'}
      <input type="text" class="prompt-input" bind:this={promptInput} bind:value={s.inputValue} />
    {/if}

    <div class="actions">
      {#if s.type === 'confirm' || s.type === 'prompt'}
        <button class="btn btn-secondary" on:click={handleCancel}>Cancel</button>
      {/if}
      <button class="btn btn-primary" on:click={handleConfirm}>OK</button>
    </div>
  </Modal>
{/if}

<style>
  .prompt-input {
    width: 100%;
    padding: 8px 12px;
    background-color: var(--toolbar-bg);
    border: 1px solid var(--panel-border);
    border-radius: 4px;
    color: var(--text-color);
    box-sizing: border-box;
    margin-bottom: 10px;
    font-family: inherit;
  }
  .prompt-input:focus { outline: none; border-color: var(--accent-color); }
  .actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
  .btn { padding: 8px 16px; border-radius: 4px; cursor: pointer; font-weight: bold; border: none; }
  .btn-primary { background: var(--accent-color); color: var(--bg-color); }
  .btn-secondary { background: var(--toolbar-bg); color: var(--text-color); border: 1px solid var(--panel-border); }
  .btn:hover:not(:disabled) { filter: brightness(1.1); }
</style>
