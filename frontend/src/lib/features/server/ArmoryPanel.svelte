<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ListCommands, RunSessionCommand } from '../../api/console.js';
  import DataTable from '../../components/ui/DataTable.svelte';
  import PanelShell from '../../components/ui/PanelShell.svelte';
  import { dialog } from '../../stores/dialogStore.js';
  import { errorMessage } from '../../utils/errors.js';
  import { stripTerminalFormatting } from '../../utils/text.js';

  const dispatch = createEventDispatcher();
  function close() { dispatch('close'); }

  let packages = [];
  let loading = true;
  let error = "";
  let warning = "";
  let installing = ""; // Track which package is currently installing

  const tableColumns = [
    { key: "armory", label: "Armory", width: 130 },
    { key: "command", label: "Command", width: 150 },
    { key: "version", label: "Version", width: 100 },
    { key: "type", label: "Type", width: 100 },
    { key: "installedStr", label: "Status", width: 120 },
    { key: "action", label: "Action", width: 150 }
  ];

  onMount(() => {
    loadArmory();
  });

  async function loadArmory() {
    loading = true;
    error = "";
    warning = "";
    try {
      const output = await RunSessionCommand("", "armory");
      const installedCommands = await ListCommands();
      packages = parseArmoryOutput(output, new Set(installedCommands || []));
      warning = parseArmoryWarnings(output);
      if (packages.length === 0) {
        const cleaned = stripTerminalFormatting(output).trim();
        error = cleaned || 'Armory returned no packages.';
      }
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  }

  function parseArmoryOutput(output, installedCommands) {
    const cleanedOutput = stripTerminalFormatting(output);
    const lines = cleanedOutput.split('\n');
    const pkgs = [];
    let columnsByName = null;

    for (const line of lines) {
      if (!line.includes('│') && !line.includes('|')) {
        if (/bundles/i.test(line)) columnsByName = null;
        continue;
      }

      const delimiter = line.includes('│') ? '│' : '|';
      const cells = line
        .split(delimiter)
        .slice(1, -1)
        .map((cell) => cell.trim());
      if (cells.length < 2) continue;

      const normalized = cells.map((cell) => cell.toLowerCase());
      if (normalized.includes('command name')) {
        columnsByName = Object.fromEntries(normalized.map((name, index) => [name, index]));
        continue;
      }
      if (!columnsByName) continue;

      const command = cells[columnsByName['command name']] || '';
      if (!command || /^[-\s]+$/.test(command)) continue;
      const installed = installedCommands.has(command);
      pkgs.push({
        armory: cells[columnsByName.armory] || '',
        command,
        version: cells[columnsByName.version] || '',
        type: cells[columnsByName.type] || '',
        installed,
        installedStr: installed ? 'Installed' : 'Not Installed',
      });
    }

    // An extension manifest can expose the same command more than once in
    // wrapped table output. Keep one stable row per armory/command pair.
    const parsed = pkgs.length > 0
      ? pkgs
      : parseBorderlessArmoryTable(cleanedOutput, installedCommands);
    return parsed.filter((pkg, index) => {
      return parsed.findIndex((candidate) =>
        candidate.armory === pkg.armory && candidate.command === pkg.command
      ) === index;
    });
  }

  function parseBorderlessArmoryTable(output, installedCommands) {
    const packagesStart = output.search(/\bPackages\b[\s\S]*?\bArmory\s+Command Name\s+Version\s+Type\s+Help\b/i);
    if (packagesStart < 0) return [];

    const packageSection = output
      .slice(packagesStart)
      .split(/\n\s*Bundles\b/i)[0]
      .replace(/\bPackages\b[\s\S]*?\bArmory\s+Command Name\s+Version\s+Type\s+Help\b/i, ' ')
      .replace(/[=─━-]{3,}/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();

    const rowPattern = /(\S+)\s+(\S+)\s+(\S+)\s+(Alias|Extension)\s+([\s\S]*?)(?=\s+\S+\s+\S+\s+\S+\s+(?:Alias|Extension)\s+|$)/gi;
    return Array.from(packageSection.matchAll(rowPattern), (match) => {
      const command = match[2];
      const installed = installedCommands.has(command);
      return {
        armory: match[1],
        command,
        version: match[3],
        type: match[4],
        help: match[5].trim(),
        installed,
        installedStr: installed ? 'Installed' : 'Not Installed',
      };
    });
  }

  function parseArmoryWarnings(output) {
    const lines = stripTerminalFormatting(output)
      .split('\n')
      .map((line) => line.trim())
      .filter((line) => line.includes('[!]') || /failed to (?:parse|download)/i.test(line));
    return [...new Set(lines)].join('\n');
  }

  async function installPackage(pkg, force = false) {
    installing = pkg.command;
    try {
      const armory = pkg.armory ? ` --armory ${quoteArgument(pkg.armory)}` : '';
      const forceFlag = force ? ' --force' : '';
      const output = await RunSessionCommand(
        "",
        `armory install${forceFlag}${armory} ${quoteArgument(pkg.command)}`,
      );
      const cleaned = stripTerminalFormatting(output).trim();
      if (/\b(?:could not install|failed to install|no package|package not found)\b/i.test(cleaned)) {
        await dialog.alert(cleaned, 'Armory Install Error');
      }
      await loadArmory();
    } catch (err) {
      await dialog.alert(
        errorMessage(err, 'Failed to install package: '),
        'Armory Install Error',
      );
    } finally {
      installing = "";
    }
  }

  async function installAll() {
    const pending = packages.filter((pkg) => !pkg.installed);
    if (pending.length === 0) {
      await dialog.alert('All available packages are already installed.', 'Armory');
      return;
    }
    if (!(await dialog.confirm(
      `Install all ${pending.length} available packages?`,
      'Install All Packages',
    ))) return;

    installing = "__all__";
    const failures = [];
    try {
      for (const pkg of pending) {
        const armory = pkg.armory ? ` --armory ${quoteArgument(pkg.armory)}` : '';
        try {
          const output = await RunSessionCommand(
            "",
            `armory install${armory} ${quoteArgument(pkg.command)}`,
          );
          const cleaned = stripTerminalFormatting(output).trim();
          if (/\b(?:could not install|failed to install|no package|package not found)\b/i.test(cleaned)) {
            failures.push(`${pkg.command}: ${cleaned}`);
          }
        } catch (err) {
          failures.push(`${pkg.command}: ${errorMessage(err)}`);
        }
      }
      await loadArmory();
      if (failures.length > 0) {
        await dialog.alert(
          `${pending.length - failures.length} installed, ${failures.length} failed.\n\n${failures.join('\n\n')}`,
          'Armory Install Results',
        );
      }
    } finally {
      installing = "";
    }
  }

  function quoteArgument(value) {
    if (!/[\s"'\\]/.test(value)) return value;
    return `"${value.replaceAll('\\', '\\\\').replaceAll('"', '\\"')}"`;
  }
</script>

<PanelShell
  title="Armory / Script Manager"
  icon="fa-shield-alt"
  on:close={close}
  width="960px"
  height="78vh"
>
  <svelte:fragment slot="actions">
    <button
      class="btn btn-primary"
      on:click={installAll}
      disabled={loading || installing !== "" || packages.every((pkg) => pkg.installed)}
    >
      <i class="fas fa-download"></i>
      {installing === "__all__" ? 'Installing All...' : 'Install All'}
    </button>
    <button class="btn" on:click={loadArmory} disabled={loading || installing !== ""}>
      <i class="fas fa-sync"></i> Refresh
    </button>
  </svelte:fragment>

  <div class="panel-content">
    {#if loading}
      <div class="message">Fetching packages from Armory... This may take a moment.</div>
    {:else if error}
      <div class="message error">{error}</div>
    {:else}
      {#if warning}
        <div class="warning">
          <i class="fas fa-triangle-exclamation"></i>
          <pre>{warning}</pre>
        </div>
      {/if}
      <DataTable data={packages} columns={tableColumns} defaultSortKey="command" let:rows let:columns>
        {#each rows as pkg}
          <tr>
            <td style="width: {columns[0].width}px; max-width: {columns[0].width}px;"><strong>{pkg.armory}</strong></td>
            <td style="width: {columns[1].width}px; max-width: {columns[1].width}px;" class="mono">{pkg.command}</td>
            <td style="width: {columns[2].width}px; max-width: {columns[2].width}px;" class="mono">{pkg.version}</td>
            <td style="width: {columns[3].width}px; max-width: {columns[3].width}px;">
              {pkg.type}
            </td>
            <td style="width: {columns[4].width}px; max-width: {columns[4].width}px;">
              {#if pkg.installed}
                <span style="color: var(--success-color);"><i class="fas fa-check-circle"></i> Installed</span>
              {:else}
                <span style="color: var(--text-muted);">Not Installed</span>
              {/if}
            </td>
            <td style="width: {columns[5].width}px; max-width: {columns[5].width}px;">
              {#if pkg.installed}
                <button
                  class="btn btn-secondary"
                  on:click={() => installPackage(pkg, true)}
                  disabled={installing !== ""}
                >
                  {installing === pkg.command ? 'Reinstalling...' : 'Reinstall'}
                </button>
              {:else if installing === pkg.command}
                <button class="btn btn-primary" disabled><i class="fas fa-spinner fa-spin"></i> Installing...</button>
              {:else}
                <button
                  class="btn btn-primary"
                  on:click={() => installPackage(pkg)}
                  disabled={installing !== ""}
                >Install</button>
              {/if}
            </td>
          </tr>
        {/each}
        {#if rows.length === 0}
          <tr>
            <td colspan="6" style="text-align: center; padding: 20px; color: var(--text-muted);">
              No Armory packages were returned.
            </td>
          </tr>
        {/if}
      </DataTable>
    {/if}
  </div>
</PanelShell>

<style>
  .panel-content {
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow-y: auto;
    padding: 20px;
  }
  .message {
    padding: 40px;
    text-align: center;
    color: var(--text-muted);
    font-size: 1.1em;
  }
  .message.error {
    color: var(--danger-color);
  }
  .warning {
    display: flex;
    gap: 10px;
    margin-bottom: 12px;
    padding: 10px 12px;
    border: 1px solid #8a6826;
    border-radius: 4px;
    background: rgba(214, 162, 62, 0.1);
    color: #d6a23e;
  }
  .warning pre {
    margin: 0;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    font: inherit;
  }
</style>
