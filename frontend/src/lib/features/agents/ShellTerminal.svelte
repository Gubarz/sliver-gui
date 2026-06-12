<script>
  import { afterUpdate, onDestroy, onMount } from 'svelte';
  import { AnsiUp } from 'ansi_up';

  import { GetShellOutput, InterruptShell, ResizeShell, WriteShell } from '../../api/console.js';
  import { onShellOutput } from '../../api/runtime.js';
  import { errorMessage } from '../../utils/errors.js';
  import { stripTerminalMetadata } from '../../utils/text.js';

  export let shell = null;

  const ansiUp = new AnsiUp();
  ansiUp.use_classes = false;
  const MAX_RENDERED_OUTPUT = 512 * 1024;
  const OUTPUT_FLUSH_MS = 100;

  let output = '';
  let pendingOutput = [];
  let pendingOutputLength = 0;
  let command = '';
  let history = [];
  let historyIndex = 0;
  let historyDraft = '';
  let closed = false;
  let outputTruncated = false;
  let scrollEl;
  let inputEl;
  let stopOutput = () => {};
  let resizeObserver;
  let flushTimer;
  let lastRows = 0;
  let lastCols = 0;

  onMount(async () => {
    stopOutput = onShellOutput((event) => {
      if (event.id !== shell.id) return;
      if (event.data) queueOutput(event.data);
      if (event.error) queueOutput(`\r\n[!] ${event.error}\r\n`);
      if (event.closed) {
        flushOutput();
        closed = true;
      }
    });
    try {
      const bufferedOutput = await GetShellOutput(shell.id);
      const liveOutput = output + pendingOutput.join('');
      output = '';
      pendingOutput = [];
      pendingOutputLength = 0;
      setOutput(bufferedOutput.endsWith(liveOutput)
        ? bufferedOutput
        : bufferedOutput + liveOutput);
    } catch (error) {
      queueOutput(`[!] ${errorMessage(error)}\r\n`);
    }
    if (shell.pty && scrollEl && typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(([entry]) => {
        // Inactive tabs remain mounted with display:none. Ignore their 0x0
        // transition so switching tabs does not resize the remote PTY twice.
        if (entry.contentRect.width <= 0 || entry.contentRect.height <= 0) return;
        const cols = Math.max(20, Math.floor(entry.contentRect.width / 8));
        const rows = Math.max(5, Math.floor(entry.contentRect.height / 18));
        if (rows === lastRows && cols === lastCols) return;
        lastRows = rows;
        lastCols = cols;
        ResizeShell(shell.id, rows, cols).catch(() => {});
      });
      resizeObserver.observe(scrollEl);
    }
    inputEl?.focus();
  });

  onDestroy(() => {
    stopOutput();
    resizeObserver?.disconnect();
    if (flushTimer) clearTimeout(flushTimer);
  });

  afterUpdate(() => {
    if (scrollEl) scrollEl.scrollTop = scrollEl.scrollHeight;
  });

  async function sendLine() {
    if (closed || !command) return;
    const line = command;
    if (history[history.length - 1] !== line) history = [...history, line];
    historyIndex = history.length;
    historyDraft = '';
    command = '';
    try {
      const newline = shell.pty ? '\n' : '\r\n';
      const payload = line.replace(/\r?\n/g, newline);
      await WriteShell(shell.id, `${payload}${newline}`);
    } catch (error) {
      queueOutput(`\r\n[!] ${errorMessage(error)}\r\n`);
      closed = true;
    }
  }

  async function sendControlC() {
    if (closed) return;
    if (command) {
      command = '';
      inputEl?.focus();
      return;
    }
    try {
      const terminated = await InterruptShell(shell.id);
      if (terminated) {
        queueOutput('\r\n[!] Windows piped shell interrupted and closed. Open a new shell to continue.\r\n');
        flushOutput();
        closed = true;
      }
    } catch (error) {
      queueOutput(`\r\n[!] ${errorMessage(error)}\r\n`);
    }
  }

  function handleInputKeydown(event) {
    const hasSelection = inputEl?.selectionStart !== inputEl?.selectionEnd;
    if (event.ctrlKey && event.key.toLowerCase() === 'c' && !hasSelection) {
      event.preventDefault();
      sendControlC();
      return;
    }
    if (event.key === 'ArrowUp' && atHistoryBoundary('up')) {
      event.preventDefault();
      navigateHistory(-1);
      return;
    }
    if (event.key === 'ArrowDown' && atHistoryBoundary('down')) {
      event.preventDefault();
      navigateHistory(1);
      return;
    }
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      sendLine();
    }
  }

  function atHistoryBoundary(direction) {
    if (inputEl?.selectionStart !== inputEl?.selectionEnd) return false;
    const cursor = inputEl?.selectionStart ?? 0;
    if (direction === 'up') {
      const firstNewline = command.indexOf('\n');
      return firstNewline < 0 || cursor <= firstNewline;
    }
    const lastNewline = command.lastIndexOf('\n');
    return lastNewline < 0 || cursor > lastNewline;
  }

  function navigateHistory(direction) {
    if (history.length === 0) return;
    if (historyIndex === history.length) historyDraft = command;
    historyIndex = Math.max(0, Math.min(history.length, historyIndex + direction));
    command = historyIndex === history.length ? historyDraft : history[historyIndex];
    requestAnimationFrame(() => {
      const end = command.length;
      inputEl?.setSelectionRange(end, end);
    });
  }

  function queueOutput(value) {
    pendingOutput.push(value);
    pendingOutputLength += value.length;
    while (pendingOutputLength > MAX_RENDERED_OUTPUT * 2 && pendingOutput.length > 1) {
      pendingOutputLength -= pendingOutput.shift().length;
      outputTruncated = true;
    }
    if (!flushTimer) {
      flushTimer = setTimeout(flushOutput, OUTPUT_FLUSH_MS);
    }
  }

  function flushOutput() {
    if (flushTimer) clearTimeout(flushTimer);
    flushTimer = null;
    if (pendingOutput.length === 0) return;
    setOutput(output + pendingOutput.join(''));
    pendingOutput = [];
    pendingOutputLength = 0;
  }

  function setOutput(value) {
    if (value.length > MAX_RENDERED_OUTPUT) {
      let start = value.length - MAX_RENDERED_OUTPUT;
      const newline = value.indexOf('\n', start);
      if (newline >= 0) start = newline + 1;
      value = value.slice(start);
      outputTruncated = true;
    }
    output = value;
  }
</script>

<div class="shell-terminal">
  <div class="shell-header">
    <span><i class="fas fa-terminal"></i> {shell.path || 'Shell'} · PID {shell.pid}</span>
    <button
      type="button"
      class="control-button"
      title={shell.pty ? 'Send Ctrl+C' : 'Interrupt and close this Windows shell'}
      on:click={sendControlC}
      disabled={closed}
    >Ctrl+C</button>
  </div>

  <div class="shell-output" bind:this={scrollEl}>
    {#if outputTruncated}<div class="truncated">Older output was discarded to keep the UI responsive.</div>{/if}
    <!-- ansi_up escapes input before producing its own span markup. -->
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    <pre>{@html ansiUp.ansi_to_html(stripTerminalMetadata(output))}</pre>
    {#if closed}<div class="closed">Shell closed</div>{/if}
  </div>

  <div class="shell-input">
    <span>$</span>
    <textarea
      bind:this={inputEl}
      bind:value={command}
      rows="1"
      disabled={closed}
      autocomplete="off"
      spellcheck="false"
      placeholder="Enter executes · Shift+Enter adds a line"
      on:keydown={handleInputKeydown}
    ></textarea>
  </div>
</div>

<style>
  .shell-terminal {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: #0b0d10;
  }

  .shell-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 7px 10px;
    border-bottom: 1px solid var(--border-color);
    background: var(--header-bg);
    color: var(--text-muted);
    font-size: 0.85em;
  }

  .control-button {
    padding: 3px 7px;
    border: 1px solid var(--border-color);
    border-radius: 3px;
    background: var(--toolbar-bg);
    color: var(--text-color);
    cursor: pointer;
  }

  .control-button:disabled {
    opacity: 0.45;
  }

  .shell-output {
    flex: 1;
    overflow: auto;
    padding: 10px;
    font-family: var(--font-mono);
    font-size: 0.88em;
  }

  pre {
    margin: 0;
    color: #c7d0d9;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .closed {
    margin-top: 8px;
    color: var(--text-muted);
    font-style: italic;
  }

  .truncated {
    margin-bottom: 8px;
    color: #d6a23e;
    font-size: 0.85em;
  }

  .shell-input {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 8px 10px;
    border-top: 1px solid var(--border-color);
    background: var(--toolbar-bg);
    color: var(--success-color);
    font-family: var(--font-mono);
  }

  .shell-input textarea {
    flex: 1;
    min-height: 20px;
    max-height: 140px;
    resize: vertical;
    border: 0;
    background: transparent;
    color: var(--text-color);
    font: inherit;
    line-height: 1.4;
    outline: none;
  }
</style>
