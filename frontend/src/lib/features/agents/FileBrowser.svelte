<script>
  import { onMount } from 'svelte';
  import {
    DownloadFile,
    listFiles,
    MakeDir,
    RemovePath,
    RenamePath,
    UploadFile,
    UploadFiles,
  } from '../../api/agents.js';
  import { onFileDrop } from '../../api/runtime.js';
  import { dialog } from '../../stores/dialogStore.js';
  import { errorMessage } from '../../utils/errors.js';
  import ContextMenu from './ContextMenu.svelte';
  import DataTable from '../../components/ui/DataTable.svelte';

  export let sessionID = "";

  // Remember each session's current directory so switching tabs (and back)
  // restores that session's location instead of bleeding state between sessions.
  const pathBySession = new Map();

  let currentPath = "";
  let files = [];
  let loading = true;
  let error = "";
  let dropZone;
  let uploading = false;

  // Context menu state
  let showContextMenu = false;
  let contextMenuX = 0;
  let contextMenuY = 0;
  let contextMenuTargetFile = null;

  let tableColumns = [
    { key: "iconStr", label: "Name", width: 300 },
    { key: "sizeStr", label: "Size", width: 100 },
    { key: "typeStr", label: "Type", width: 100 },
    { key: "modTimeStr", label: "Last Modified", width: 250 }
  ];

  $: normalizedFiles = files.map(f => ({
    ...f,
    rawFile: f,
    isDir: f.IsDir || f.isDir,
    iconStr: (f.IsDir || f.isDir) ? `📁 ${f.Name || f.name}` : `📄 ${f.Name || f.name}`,
    sizeStr: (f.IsDir || f.isDir) ? '-' : formatSize(f.Size || f.size),
    typeStr: (f.IsDir || f.isDir) ? 'Directory' : 'File',
    modTimeStr: new Date((f.ModTime || f.modTime || 0) * 1000).toLocaleString()
  }));

  $: fileMenuActions = [
    ...(contextMenuTargetFile && !(contextMenuTargetFile.IsDir || contextMenuTargetFile.isDir) ? 
      [{ id: 'download', icon: 'fas fa-download', label: 'Download' }] : []),
  ];
  
  $: footerMenuActions = [
    { id: 'rename', icon: 'fas fa-pen', label: 'Rename' }
  ];

  $: dangerMenuActions = [
    { id: 'delete', icon: 'fas fa-trash', label: 'Delete' }
  ];

  // Reload when the active session changes (the component instance is reused, so
  // onMount alone would never refresh). Restores the remembered path if any.
  let lastSession = null;
  $: if (sessionID !== lastSession) {
    lastSession = sessionID;
    loadFiles(pathBySession.get(sessionID) || "");
  }

  onMount(() => onFileDrop(async (x, y, paths) => {
    const target = document.elementFromPoint(x, y);
    if (!dropZone?.contains(target) || !paths?.length || uploading) return;

    uploading = true;
    try {
      await UploadFiles(sessionID, currentPath, paths);
      await loadFiles(currentPath);
    } catch (err) {
      await dialog.alert(errorMessage(err, "Upload failed: "), 'Upload Error');
    } finally {
      uploading = false;
    }
  }));

  async function loadFiles(path) {
    loading = true;
    error = "";
    try {
      const result = await listFiles(sessionID, path);
      files = result.files;
      currentPath = result.path;
      pathBySession.set(sessionID, currentPath);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  }

  function formatSize(bytes) {
    if (bytes === 0 || !bytes) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function handleDoubleClick(file) {
    const isDir = file.IsDir || file.isDir;
    const name = file.Name || file.name;
    if (isDir) {
      let sep = currentPath.includes('\\') ? '\\' : '/';
      let nextPath = currentPath.endsWith(sep) ? currentPath + name : currentPath + sep + name;
      loadFiles(nextPath);
    }
  }

  // A drive root ("C:\") or filesystem root ("/") — going up from here is invalid
  // and the implant returns an RPC error, so we no-op.
  function isAtRoot(path) {
    return path === "" || path === "/" || /^[A-Za-z]:\\?$/.test(path);
  }

  function goUp() {
    if (isAtRoot(currentPath)) return;

    const isWindows = currentPath.includes('\\') || /^[A-Za-z]:/.test(currentPath);
    const sep = isWindows ? '\\' : '/';

    const parts = currentPath.split(sep);
    if (parts[parts.length - 1] === '') parts.pop(); // trailing separator
    parts.pop();
    let nextPath = parts.join(sep);

    if (isWindows) {
      // Keep the drive root as "C:\" rather than "C:".
      if (/^[A-Za-z]:$/.test(nextPath)) nextPath += '\\';
      else if (nextPath === '') return;
    } else if (nextPath === '') {
      nextPath = '/';
    }

    loadFiles(nextPath);
  }

  async function downloadFile(f) {
    let sep = currentPath.includes('\\') ? '\\' : '/';
    let name = f.Name || f.name;
    let path = currentPath.endsWith(sep) ? currentPath + name : currentPath + sep + name;
    try {
      await DownloadFile(sessionID, path);
    } catch(err) {
      await dialog.alert(errorMessage(err, "Download failed: "), 'Download Error');
    }
  }

  async function uploadFile() {
    try {
      await UploadFile(sessionID, currentPath);
      loadFiles(currentPath); // Refresh
    } catch(err) {
      await dialog.alert(errorMessage(err, "Upload failed: "), 'Upload Error');
    }
  }

  function sep() { return currentPath.includes('\\') ? '\\' : '/'; }
  function joinPath(name) {
    const s = sep();
    return currentPath.endsWith(s) ? currentPath + name : currentPath + s + name;
  }

  async function newFolder() {
    const name = await dialog.prompt("New folder name:", "Create Folder");
    if (!name) return;
    try {
      await MakeDir(sessionID, joinPath(name));
      loadFiles(currentPath);
    } catch (err) { await dialog.alert(errorMessage(err, "Create failed: "), 'Error'); }
  }

  async function deleteFile(f) {
    const name = f.Name || f.name;
    const isDir = f.IsDir || f.isDir;
    if (!(await dialog.confirm(`Delete ${isDir ? 'folder' : 'file'} "${name}"?`, 'Confirm Delete'))) return;
    try {
      await RemovePath(sessionID, joinPath(name), !!isDir);
      loadFiles(currentPath);
    } catch (err) { await dialog.alert(errorMessage(err, "Delete failed: "), 'Error'); }
  }

  async function renameFile(f) {
    const name = f.Name || f.name;
    const newName = await dialog.prompt(`Rename "${name}" to:`, "Rename Item", name);
    if (!newName || newName === name) return;
    try {
      await RenamePath(sessionID, joinPath(name), joinPath(newName));
      loadFiles(currentPath);
    } catch (err) { await dialog.alert(errorMessage(err, "Rename failed: "), 'Error'); }
  }

  function handleRightClick(event, file) {
    showContextMenu = true;
    contextMenuX = event.clientX;
    contextMenuY = event.clientY;
    contextMenuTargetFile = file;
  }

  function handleMenuSelect(event) {
    showContextMenu = false;
    const action = event.detail.action;
    if (!contextMenuTargetFile) return;

    if (action === 'download') {
      downloadFile(contextMenuTargetFile);
    } else if (action === 'rename') {
      renameFile(contextMenuTargetFile);
    } else if (action === 'delete') {
      deleteFile(contextMenuTargetFile);
    }
  }

  function hideContextMenu() {
    showContextMenu = false;
  }

</script>

<svelte:window on:click={hideContextMenu} />

<div 
  bind:this={dropZone}
  class="tab-wrapper" 
>
  <div class="drag-overlay">
    <i class="fas fa-upload fa-3x"></i>
    <h2>Drop files to upload to:</h2>
    <span class="mono">{currentPath}</span>
  </div>
  <div class="tab-header">
    <button class="btn" style="margin-right: 10px;" on:click={goUp} disabled={isAtRoot(currentPath)}>&uarr; Up</button>
    <input type="text" class="path-input" bind:value={currentPath} on:keydown={(e) => {if(e.key==='Enter') loadFiles(currentPath)}} />
    <button class="btn" style="margin-left: 10px;" on:click={() => loadFiles(currentPath)}>Go</button>
    <button class="btn" style="margin-left: 10px;" on:click={newFolder}>New Folder</button>
    <button class="btn btn-primary" style="margin-left: 10px;" on:click={uploadFile} disabled={uploading}>
      {uploading ? 'Uploading...' : 'Upload Here'}
    </button>
  </div>
  
  <div class="tab-content">
    {#if loading}
      <div style="padding: 20px; text-align: center;">Loading directory...</div>
    {:else if error}
      <div style="padding: 20px; color: var(--danger-color); text-align: center;">{error}</div>
    {:else}
      <DataTable data={normalizedFiles} columns={tableColumns} defaultSortKey="iconStr" let:rows let:columns>
        {#each rows as f}
          <tr 
            on:dblclick={() => handleDoubleClick(f.rawFile)}
            on:contextmenu|preventDefault={(e) => handleRightClick(e, f.rawFile)}
          >
            <td style="width: {columns[0].width}px; max-width: {columns[0].width}px;">
              {#if f.isDir}
                <span style="color: var(--accent-color)">{f.iconStr}</span>
              {:else}
                <span>{f.iconStr}</span>
              {/if}
            </td>
            <td style="width: {columns[1].width}px; max-width: {columns[1].width}px;">{f.sizeStr}</td>
            <td style="width: {columns[2].width}px; max-width: {columns[2].width}px;">{f.typeStr}</td>
            <td class="mono" style="width: {columns[3].width}px; max-width: {columns[3].width}px;">{f.modTimeStr}</td>
          </tr>
        {/each}
        {#if rows.length === 0}
          <tr>
            <td colspan={columns.length} style="text-align: center; color: var(--text-muted); padding: 20px;">No files found.</td>
          </tr>
        {/if}
      </DataTable>
    {/if}
  </div>
</div>

{#if showContextMenu}
  <ContextMenu
    x={contextMenuX}
    y={contextMenuY}
    coreActions={fileMenuActions}
    footerActions={footerMenuActions}
    dangerActions={dangerMenuActions}
    categories={[]}
    onSelect={handleMenuSelect}
  />
{/if}

<style>
  .tab-wrapper {
    --wails-drop-target: drop;
    position: relative;
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-color);
  }
  .tab-header {
    display: flex;
    background-color: var(--header-bg);
    padding: 10px 15px;
    border-bottom: 1px solid var(--border-color);
  }
  .path-input {
    flex: 1;
    background: var(--bg-color);
    border: 1px solid var(--border-color);
    color: var(--text-color);
    padding: 5px 10px;
    font-family: var(--font-mono);
    border-radius: 3px;
    outline: none;
  }
  .tab-content {
    flex: 1;
    overflow: auto;
  }
  tr:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }
  .tab-wrapper:global(.wails-drop-target-active) {
    border: 2px dashed var(--accent-color);
  }
  .drag-overlay {
    position: absolute;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: none;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    z-index: 50;
    color: var(--accent-color);
  }
  .tab-wrapper:global(.wails-drop-target-active) .drag-overlay {
    display: flex;
  }
  .drag-overlay h2 {
    margin-top: 20px;
    color: #fff;
  }
</style>
