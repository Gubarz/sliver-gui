<script>
  export let initialLeftWidth = 30; // percentage
  export let minLeftWidth = 10;
  export let maxLeftWidth = 90;
  
  let leftWidth = initialLeftWidth;
  let isResizing = false;
  
  function startResize() {
    isResizing = true;
    window.addEventListener('mousemove', handleResize);
    window.addEventListener('mouseup', stopResize);
  }
  
  function handleResize(e) {
    if (!isResizing) return;
    const container = document.getElementById('split-pane-container');
    if (!container) return;
    
    const rect = container.getBoundingClientRect();
    let newWidth = ((e.clientX - rect.left) / rect.width) * 100;
    
    newWidth = Math.max(minLeftWidth, Math.min(newWidth, maxLeftWidth));
    leftWidth = newWidth;
  }
  
  function stopResize() {
    isResizing = false;
    window.removeEventListener('mousemove', handleResize);
    window.removeEventListener('mouseup', stopResize);
  }
</script>

<div id="split-pane-container" class="split-pane-container" class:resizing={isResizing}>
  <div class="pane left-pane" style="width: {leftWidth}%">
    <slot name="left"></slot>
  </div>
  
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="divider" on:mousedown={startResize}></div>
  
  <div class="pane right-pane" style="width: {100 - leftWidth}%">
    <slot name="right"></slot>
  </div>
</div>

<style>
  .split-pane-container {
    display: flex;
    width: 100%;
    height: 100%;
    overflow: hidden;
  }
  
  .split-pane-container.resizing {
    user-select: none;
  }
  
  .pane {
    height: 100%;
    overflow: auto;
  }
  
  .divider {
    width: 6px;
    height: 100%;
    background-color: var(--panel-border, #333);
    cursor: col-resize;
    position: relative;
    z-index: 10;
    transition: background-color 0.2s;
  }
  
  .divider:hover, .split-pane-container.resizing .divider {
    background-color: var(--accent-color, #007bff);
  }
</style>
