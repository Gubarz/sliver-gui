<script>
  import {
    minimizeWindow,
    quitApplication,
    toggleMaximizeWindow,
  } from '../../api/runtime.js';

  function minimize() {
    minimizeWindow();
  }
  
  function maximize() {
    toggleMaximizeWindow();
  }
  
  function closeApp() {
    quitApplication();
  }
</script>

<div class="titlebar" style="--wails-draggable:drag">
  <div class="drag-region">
    <div class="logo">
      <img src="/wails.png" alt="Sliver GUI" />
      <span>Sliver GUI</span>
    </div>
  </div>
  
  <div class="window-controls">
    <button type="button" class="control-btn minimize" aria-label="Minimize window" on:click={minimize}>
      <svg width="12" height="12" viewBox="0 0 12 12"><rect fill="currentColor" width="10" height="1" x="1" y="6"></rect></svg>
    </button>
    <button type="button" class="control-btn maximize" aria-label="Maximize window" on:click={maximize}>
      <svg width="12" height="12" viewBox="0 0 12 12"><rect width="9" height="9" x="1.5" y="1.5" fill="none" stroke="currentColor"></rect></svg>
    </button>
    <button type="button" class="control-btn close" aria-label="Close window" on:click={closeApp}>
      <svg width="12" height="12" viewBox="0 0 12 12"><polygon fill="currentColor" points="11 1.5 10.5 1 6 5.5 1.5 1 1 1.5 5.5 6 1 10.5 1.5 11 6 6.5 10.5 11 11 10.5 6.5 6"></polygon></svg>
    </button>
  </div>
</div>

<style>
  .titlebar {
    height: 32px;
    background-color: var(--header-bg);
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-color);
    user-select: none;
    flex-shrink: 0;
  }
  
  .drag-region {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: center;
    padding-left: 12px;
    /* This makes Wails treat it as the draggable region */
    --wails-draggable: drag;
  }
  
  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-color);
    pointer-events: none; /* Prevent interference with dragging */
  }
  
  .logo img {
    width: 16px;
    height: 16px;
  }
  
  .window-controls {
    display: flex;
    height: 100%;
    /* No dragging on the controls themselves */
    --wails-draggable: no-drag;
  }
  
  .control-btn {
    width: 46px;
    height: 100%;
    padding: 0;
    border: 0;
    background: transparent;
    display: flex;
    justify-content: center;
    align-items: center;
    color: var(--text-muted);
    cursor: default;
    transition: background-color 0.1s, color 0.1s;
  }
  
  .control-btn:hover {
    background-color: var(--row-hover);
    color: var(--text-color);
  }
  
  .control-btn.close:hover {
    background-color: #e81123;
    color: white;
  }
</style>
