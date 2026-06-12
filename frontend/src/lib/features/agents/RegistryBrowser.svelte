<script>
  import {
    CreateRegistryKey,
    DeleteRegistryEntry,
    ListRegistrySubKeys,
    ListRegistryValues,
    ReadRegistryValue,
    WriteRegistryValue,
  } from '../../api/agents.js';
  import { dialog } from '../../stores/dialogStore.js';
  import { errorMessage } from '../../utils/errors.js';
  import SplitPane from '../../components/layout/SplitPane.svelte';
  import DataTable from '../../components/ui/DataTable.svelte';
  import ContextMenu from './ContextMenu.svelte';

  export let sessionID = "";

  const hives = ["HKLM", "HKCU", "HKU", "HKCR", "HKCC"];

  let currentHive = "HKLM";
  let currentPath = "";
  let subKeys = [];
  let values = [];
  let loading = false;
  let error = "";
  let showContextMenu = false;
  let contextMenuX = 0;
  let contextMenuY = 0;
  let contextTarget = null;

  let tableColumns = [
    { key: "iconStr", label: "Name", width: 300 },
    { key: "typeStr", label: "Type", width: 150 },
    { key: "DataStr", label: "Data", width: 400 }
  ];

  $: normalizedData = [
    ...subKeys.map(k => ({ iconStr: `📁 ${k}`, typeStr: 'Key', DataStr: '', rawName: k, isKey: true })),
    ...values.map(value => ({
      iconStr: `📄 ${value.name}`,
      typeStr: value.type,
      DataStr: value.value,
      rawName: value.name,
      isKey: false,
      rawValue: value,
    }))
  ];

  $: coreMenuActions = contextTarget
    ? contextTarget.isKey
      ? [{ id: 'open', icon: 'fas fa-folder-open', label: 'Open' }]
      : [{ id: 'editValue', icon: 'fas fa-pen', label: 'Edit Value' }]
    : [
        { id: 'newKey', icon: 'fas fa-folder-plus', label: 'New Key' },
        { id: 'newValue', icon: 'fas fa-plus', label: 'New Value' },
      ];
  $: dangerMenuActions = contextTarget
    ? [{
        id: 'deleteEntry',
        icon: 'fas fa-trash',
        label: contextTarget.isKey ? 'Delete Key' : 'Delete Value',
      }]
    : [];

  $: if (sessionID) {
    loadRegistry(currentHive, currentPath);
  }

  async function loadRegistry(hive, path) {
    loading = true;
    error = "";
    try {
      const keysResponse = await ListRegistrySubKeys(sessionID, hive, path);
      subKeys = keysResponse.Subkeys || keysResponse.SubKeys || [];

      const valuesResponse = await ListRegistryValues(sessionID, hive, path);
      const valueNames = valuesResponse.ValueNames || valuesResponse.Values || [];
      values = await Promise.all(valueNames.map(async (name) => {
        try {
          const value = await ReadRegistryValue(sessionID, hive, path, name);
          return {
            name,
            type: value.type || value.Type || 'Value',
            value: value.value ?? value.Value ?? '',
          };
        } catch {
          return { name, type: 'Value', value: '<unavailable>' };
        }
      }));
    } catch (err) {
      error = errorMessage(err);
      subKeys = [];
      values = [];
    } finally {
      loading = false;
    }
  }

  function handleHiveClick(hive) {
    currentHive = hive;
    currentPath = "";
    loadRegistry(currentHive, currentPath);
  }

  function handleKeyDoubleClick(keyName) {
    let sep = "\\";
    if (currentPath === "") {
      currentPath = keyName;
    } else {
      currentPath = currentPath + sep + keyName;
    }
    loadRegistry(currentHive, currentPath);
  }

  function goUp() {
    let sep = "\\";
    if (currentPath === "") return;
    
    let parts = currentPath.split(sep);
    parts.pop();
    currentPath = parts.join(sep);
    loadRegistry(currentHive, currentPath);
  }

  async function createKey() {
    const name = await dialog.prompt('Key name:', 'New Registry Key');
    if (!name) return;
    try {
      await CreateRegistryKey(sessionID, currentHive, currentPath, name);
      await loadRegistry(currentHive, currentPath);
    } catch (err) {
      await dialog.alert(errorMessage(err, 'Create key failed: '), 'Registry Error');
    }
  }

  async function writeValue(existing = null) {
    const name = existing?.rawName || await dialog.prompt('Value name:', 'Registry Value');
    if (!name) return;
    const type = await dialog.prompt(
      'Value type (string, dword, qword, binary):',
      'Registry Value Type',
      'string',
    );
    if (!type) return;
    const value = await dialog.prompt(
      type.toLowerCase() === 'binary' ? 'Hexadecimal value:' : 'Value:',
      existing ? 'Edit Registry Value' : 'New Registry Value',
      existing?.rawValue?.value || '',
    );
    if (value === null) return;
    try {
      await WriteRegistryValue(sessionID, currentHive, currentPath, name, type, value);
      await loadRegistry(currentHive, currentPath);
    } catch (err) {
      await dialog.alert(errorMessage(err, 'Write value failed: '), 'Registry Error');
    }
  }

  async function deleteEntry(row) {
    const kind = row.isKey ? 'key' : 'value';
    if (!(await dialog.confirm(`Delete registry ${kind} "${row.rawName}"?`, 'Confirm Delete'))) return;
    try {
      await DeleteRegistryEntry(sessionID, currentHive, currentPath, row.rawName);
      await loadRegistry(currentHive, currentPath);
    } catch (err) {
      await dialog.alert(errorMessage(err, `Delete ${kind} failed: `), 'Registry Error');
    }
  }

  function showMenu(event, row = null) {
    contextTarget = row;
    contextMenuX = event.clientX;
    contextMenuY = event.clientY;
    showContextMenu = true;
  }

  function handleMenuSelect(event) {
    showContextMenu = false;
    switch (event.detail.action) {
      case 'open': handleKeyDoubleClick(contextTarget.rawName); break;
      case 'newKey': createKey(); break;
      case 'newValue': writeValue(); break;
      case 'editValue': writeValue(contextTarget); break;
      case 'deleteEntry': deleteEntry(contextTarget); break;
    }
  }
</script>

<svelte:window on:click={() => showContextMenu = false} />

<SplitPane initialLeftWidth={25}>
  <svelte:fragment slot="left">
    <div class="tab-wrapper left-pane-bg">
      <div class="tab-header">
        <span><strong>Registry Hives</strong></span>
      </div>
      <div class="tab-content">
        <ul class="hive-list">
          {#each hives as hive}
            <!-- svelte-ignore a11y-click-events-have-key-events -->
            <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
            <li 
              class:active={hive === currentHive}
              on:click={() => handleHiveClick(hive)}
            >
              <i class="fas fa-database" style="margin-right: 8px;"></i> {hive}
            </li>
          {/each}
        </ul>
      </div>
    </div>
  </svelte:fragment>

  <svelte:fragment slot="right">
    <div class="tab-wrapper">
      <div class="tab-header">
        <button class="btn" style="margin-right: 10px;" on:click={goUp} disabled={currentPath === ""}>&uarr; Up</button>
        <div class="path-display mono">
          {currentHive}\{currentPath}
        </div>
        <button class="btn" style="margin-left: 10px;" on:click={() => loadRegistry(currentHive, currentPath)}>Refresh</button>
        <button class="btn" style="margin-left: 10px;" on:click={createKey}>New Key</button>
        <button class="btn btn-primary" style="margin-left: 10px;" on:click={() => writeValue()}>New Value</button>
      </div>
      
      <div class="tab-content" role="region" on:contextmenu|preventDefault={(event) => showMenu(event)}>
        {#if loading}
          <div style="padding: 20px; text-align: center;">Loading registry...</div>
        {:else if error}
          <div style="padding: 20px; color: var(--danger-color); text-align: center;">{error}</div>
        {:else}
          <DataTable data={normalizedData} columns={tableColumns} defaultSortKey="iconStr" let:rows let:columns>
            {#each rows as row}
              {#if row.isKey}
                <tr
                  on:dblclick={() => handleKeyDoubleClick(row.rawName)}
                  on:contextmenu|preventDefault|stopPropagation={(event) => showMenu(event, row)}
                >
                  <td style="width: {columns[0].width}px; max-width: {columns[0].width}px;"><span style="color: var(--accent-color)">{row.iconStr}</span></td>
                  <td style="width: {columns[1].width}px; max-width: {columns[1].width}px;">{row.typeStr}</td>
                  <td style="width: {columns[2].width}px; max-width: {columns[2].width}px;" class="mono">{row.DataStr}</td>
                </tr>
              {:else}
                <tr
                  on:dblclick={() => writeValue(row)}
                  on:contextmenu|preventDefault|stopPropagation={(event) => showMenu(event, row)}
                >
                  <td style="width: {columns[0].width}px; max-width: {columns[0].width}px;"><span>{row.iconStr}</span></td>
                  <td style="width: {columns[1].width}px; max-width: {columns[1].width}px;">{row.typeStr}</td>
                  <td style="width: {columns[2].width}px; max-width: {columns[2].width}px;" class="mono">{row.DataStr}</td>
                </tr>
              {/if}
            {/each}
            {#if rows.length === 0}
              <tr>
                <td colspan="3" style="text-align: center; color: var(--text-muted); padding: 20px;">No keys or values found.</td>
              </tr>
            {/if}
          </DataTable>
        {/if}
      </div>
    </div>
  </svelte:fragment>
</SplitPane>

{#if showContextMenu}
  <ContextMenu
    x={contextMenuX}
    y={contextMenuY}
    coreActions={coreMenuActions}
    footerActions={[]}
    dangerActions={dangerMenuActions}
    categories={[]}
    onSelect={handleMenuSelect}
  />
{/if}

<style>
  .tab-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-color);
  }
  .tab-header {
    display: flex;
    align-items: center;
    background-color: var(--header-bg);
    padding: 10px 15px;
    border-bottom: 1px solid var(--border-color);
  }
  .path-display {
    flex: 1;
    background: var(--bg-color);
    border: 1px solid var(--border-color);
    color: var(--text-color);
    padding: 5px 10px;
    border-radius: 3px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .tab-content {
    flex: 1;
    overflow: auto;
  }
  tr:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }
  .left-pane-bg {
    background-color: var(--panel-bg);
    border-right: 1px solid var(--border-color);
  }
  .hive-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .hive-list li {
    padding: 12px 20px;
    cursor: pointer;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    transition: background 0.2s;
  }
  .hive-list li:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  .hive-list li.active {
    background: var(--accent-color);
    color: var(--bg-color);
    border-left: 3px solid var(--bg-color);
  }
</style>
