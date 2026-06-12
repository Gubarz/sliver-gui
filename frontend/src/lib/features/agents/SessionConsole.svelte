<script>
  import { afterUpdate, createEventDispatcher } from "svelte";
  import { CompleteCommand, CompletePath } from "../../api/agents.js";
  import {
    tick,
    getSession,
    dispatchCommand,
    pushLine,
    getCommands,
  } from "../../stores/consoleStore.js";
  import { AnsiUp } from "ansi_up";
  const ansiUp = new AnsiUp();
  ansiUp.use_classes = false; // Use inline styles for simplicity

  export let sessionID = "";

  const dispatch = createEventDispatcher();
  let command = "";
  let scrollEl;
  let inputEl;

  // Re-resolve the visible session buffer whenever the prop changes or any
  // session buffer is updated (subscribing to `$tick` drives reactivity).
  let session;
  $: $tick, sessionID, (session = getSession(sessionID));
  $: shortID = sessionID ? sessionID.substring(0, 8) : "";

  function onKeydown(e) {
    if (e.key === "Enter") {
      const cmd = command;
      command = "";
      if (cmd.trim().toLowerCase() === 'shell') {
        dispatch('shell');
      } else {
        dispatchCommand(sessionID, cmd);
      }
    } else if (e.key === "Tab") {
      e.preventDefault();
      complete();
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      navHistory(-1);
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      navHistory(1);
    }
  }

  // Tab completion: the first token completes against the command list; any
  // later token completes against nested subcommands or the target's filesystem via the backend.
  async function complete() {
    const parts = command.split(" ");
    const token = parts[parts.length - 1];
    const isFirst = parts.length === 1;

    let candidates = [];
    if (isFirst) {
      const cmds = await getCommands();
      candidates = cmds.filter((c) => c.startsWith(token.toLowerCase()));
    } else {
      try {
        // Try completing subcommands first
        candidates = await CompleteCommand(
          sessionID,
          command,
        );
        // Fallback to path completion if no subcommands matched
        if (!candidates || candidates.length === 0) {
          candidates = await CompletePath(sessionID, token);
        }
      } catch {
        candidates = [];
      }
    }
    if (!candidates || candidates.length === 0) return;

    if (candidates.length === 1) {
      parts[parts.length - 1] = candidates[0];
      command = parts.join(" ");
      return;
    }

    // Multiple matches: extend to the longest common prefix, and if that adds
    // nothing, list the options (like a shell does).
    const cp = commonPrefix(candidates);
    if (cp.length > token.length) {
      parts[parts.length - 1] = cp;
      command = parts.join(" ");
    } else {
      pushLine(sessionID, { type: "completions", items: candidates });
    }
  }

  function commonPrefix(items) {
    if (items.length === 0) return "";
    let prefix = items[0];
    for (const s of items) {
      while (!s.startsWith(prefix)) {
        prefix = prefix.slice(0, -1);
        if (prefix === "") return "";
      }
    }
    return prefix;
  }

  function navHistory(dir) {
    const s = getSession(sessionID);
    if (!s.history.length) return;
    s.histIdx = Math.max(0, Math.min(s.history.length, s.histIdx + dir));
    command = s.history[s.histIdx] ?? "";
  }

  // Keep the view pinned to the latest output.
  afterUpdate(() => {
    if (scrollEl) scrollEl.scrollTop = scrollEl.scrollHeight;
  });

  // Refocus the input when switching into a session.
  $: if (sessionID && inputEl) inputEl.focus();
</script>

<div class="console-wrapper">
  <div class="console-header">
    Console
  </div>

  <div class="console-output" bind:this={scrollEl}>
    {#each session.lines as item}
      {#if item.type === "input"}
        <div class="cmd-line">
          <span class="prompt">sliver ({shortID}) &gt;</span>
          {item.text}
        </div>
      {:else if item.type === "output"}
        <!-- ansi_up escapes source HTML before adding its own color spans. -->
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        <pre class="cmd-output">{@html ansiUp.ansi_to_html(item.text)}</pre>
      {:else if item.type === "completions"}
        <div class="completions-grid">
          {#each item.items as comp}
            <div class="completion-item">{comp}</div>
          {/each}
        </div>
      {:else}
        <!-- ansi_up escapes source HTML before adding its own color spans. -->
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        <pre class="cmd-error">{@html ansiUp.ansi_to_html(item.text)}</pre>
      {/if}
    {/each}
    {#if session.busy}
      <div class="cmd-output dim">running…</div>
    {/if}
  </div>

  <div class="console-input">
    <span class="prompt">sliver ({shortID}) &gt;</span>
    <div class="console-input-wrapper">
      <!-- svelte-ignore a11y-autofocus -->
      <input
        type="text"
        bind:this={inputEl}
        bind:value={command}
        on:keydown={onKeydown}
        placeholder="Type a command..."
        disabled={session.busy}
        autofocus
      />
    </div>
  </div>
</div>

<style>
  .console-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--panel-bg);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    overflow: hidden;
  }
  .console-header {
    background-color: var(--header-bg);
    padding: 8px 15px;
    font-size: 0.9em;
    border-bottom: 1px solid var(--border-color);
  }
  .console-output {
    flex: 1;
    padding: 10px;
    overflow-y: auto;
    font-family: var(--font-mono);
    font-size: 0.85em;
    text-align: left;
  }
  .console-input {
    display: flex;
    align-items: center;
    padding: 10px;
    background-color: var(--toolbar-bg);
    border-top: 1px solid var(--border-color);
    font-family: var(--font-mono);
  }
  .console-input-wrapper {
    display: flex;
    flex: 1;
    align-items: center;
    background-color: var(--panel-bg);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    padding: 6px 12px;
    margin-left: 10px;
    transition:
      border-color 0.2s,
      box-shadow 0.2s;
  }
  .console-input-wrapper:focus-within {
    border-color: var(--accent-color);
  }
  .console-input input {
    flex: 1;
    background: transparent;
    border: none;
    color: var(--text-color);
    font-family: var(--font-mono);
    outline: none;
  }
  .console-input input:focus {
    box-shadow: none !important;
    border-color: transparent !important;
  }
  .prompt {
    color: var(--success-color);
    font-weight: bold;
    white-space: nowrap;
  }
  .cmd-line {
    margin-bottom: 5px;
  }
  .cmd-output {
    color: #a0aab5;
    margin: 0 0 12px 0;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .cmd-output.dim {
    color: #666;
  }
  .cmd-error {
    color: var(--danger-color);
    margin: 0 0 12px 0;
    white-space: pre-wrap;
  }
  .completions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 4px 10px;
    margin: 0 0 12px 0;
    color: #a0aab5;
  }
  .completion-item {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
