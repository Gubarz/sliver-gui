<script>
  import { createEventDispatcher } from 'svelte';
  import { stripTerminalFormatting } from '../../utils/text.js';
  import Modal from './Modal.svelte';

  export let sessionID = '';
  export let command = null;

  const dispatch = createEventDispatcher();

  let argumentValues = {};
  let flagValues = {};
  let advancedLine = '';
  let showAdvanced = false;
  let commandPreview = '';
  let formValid = false;

  $: if (command) initializeForm(command);
  $: commandPreview = command
    ? buildCommand(command, flagValues, argumentValues, advancedLine)
    : '';
  $: formValid = command
    ? isValid(command, flagValues, argumentValues)
    : false;

  function initializeForm(currentCommand) {
    const nextArguments = {};
    const nextFlags = {};

    for (const argument of currentCommand.arguments ?? []) {
      nextArguments[argument.name] = argumentValues[argument.name] ?? '';
    }
    for (const flag of currentCommand.flags ?? []) {
      nextFlags[flag.name] = flagValues[flag.name] ?? (flag.boolean ? false : '');
    }

    argumentValues = nextArguments;
    flagValues = nextFlags;
  }

  function quote(value) {
    const text = String(value);
    if (text !== '' && !/[\s"'\\]/.test(text)) return text;
    return `"${text.replaceAll('\\', '\\\\').replaceAll('"', '\\"')}"`;
  }

  function buildCommand(currentCommand, currentFlags, currentArguments, extraLine) {
    const parts = [currentCommand.path];

    for (const flag of currentCommand.flags ?? []) {
      const value = currentFlags[flag.name];
      if (flag.boolean) {
        if (value) parts.push(`--${flag.name}`);
      } else if (value !== '' && value !== null && value !== undefined) {
        parts.push(`--${flag.name}`, quote(value));
      }
    }

    for (const argument of currentCommand.arguments ?? []) {
      const value = currentArguments[argument.name];
      if (!value) continue;
      if (argument.variadic) {
        parts.push(value);
      } else {
        parts.push(quote(value));
      }
    }

    const extra = extraLine.trim();
    if (extra) parts.push(extra);
    return parts.join(' ');
  }

  function execute() {
    if (!formValid) return;
    dispatch('execute', { cmd: commandPreview });
  }

  function isValid(currentCommand, currentFlags, currentArguments) {
    if (!currentCommand.supported) return false;
    return (currentCommand.arguments ?? []).every(
      (argument) => !argument.required || Boolean(currentArguments[argument.name]?.trim()),
    ) && (currentCommand.flags ?? []).every(
      (flag) => !flag.required || (
        flag.boolean
          ? Boolean(currentFlags[flag.name])
          : Boolean(String(currentFlags[flag.name] ?? '').trim())
      ),
    );
  }

  function inputType(flag) {
    const name = flag.name.toLowerCase();
    if (name.includes('password') || name === 'pass') return 'password';
    if (/^(int|int32|int64|uint|uint32|uint64|float)/.test(flag.type)) return 'number';
    return 'text';
  }

  function label(text) {
    return text
      .replaceAll('-', ' ')
      .replace(/\b\w/g, (character) => character.toUpperCase());
  }

</script>

{#if command}
  <Modal title="{command.name}{sessionID ? ` - ${sessionID}` : ''}" width="min(760px, 94vw)" on:close>
    <div class="command-form">
      {#if command.description}
        <p class="description">{stripTerminalFormatting(command.description)}</p>
      {/if}

      <div class="usage">
        <span>Usage</span>
        <code>{command.usage}</code>
      </div>

      {#if !command.supported}
        <div class="unavailable">{command.unavailable}</div>
      {/if}

      {#if command.arguments?.length}
        <section>
          <h3>Arguments</h3>
          {#each command.arguments as argument}
            <label class="field">
              <span>
                {label(argument.name)}
                {#if argument.required}<strong>*</strong>{/if}
                {#if argument.variadic}<small>multiple values</small>{/if}
              </span>
              <input
                type="text"
                bind:value={argumentValues[argument.name]}
                placeholder={argument.name}
              />
            </label>
          {/each}
        </section>
      {/if}

      {#if command.flags?.length}
        <section>
          <h3>Options</h3>
          <div class="flag-grid">
            {#each command.flags as flag}
              {#if flag.boolean}
                <label class="boolean-field">
                  <input type="checkbox" bind:checked={flagValues[flag.name]} />
                  <span>
                    {label(flag.name)}
                    {#if flag.shorthand}<code>-{flag.shorthand}</code>{/if}
                    {#if flag.required}<strong>*</strong>{/if}
                    <small>{flag.usage}</small>
                  </span>
                </label>
              {:else}
                <label class="field">
                  <span>
                    {label(flag.name)}
                    {#if flag.shorthand}<code>-{flag.shorthand}</code>{/if}
                    {#if flag.required}<strong>*</strong>{/if}
                  </span>
                  <input
                    type={inputType(flag)}
                    bind:value={flagValues[flag.name]}
                    placeholder={flag.default && flag.default !== '0' ? flag.default : flag.usage}
                  />
                  <small>{flag.usage}</small>
                </label>
              {/if}
            {/each}
          </div>
        </section>
      {/if}

      <button
        type="button"
        class="advanced-toggle"
        aria-expanded={showAdvanced}
        on:click={() => showAdvanced = !showAdvanced}
      >
        <i class="fas fa-chevron-{showAdvanced ? 'down' : 'right'}"></i>
        Advanced command-line arguments
      </button>
      {#if showAdvanced}
        <label class="field advanced">
          <span>Additional arguments</span>
          <input
            type="text"
            bind:value={advancedLine}
            placeholder="Passed through exactly as entered"
            on:keydown={(event) => event.key === 'Enter' && execute()}
          />
        </label>
      {/if}

      <div class="preview">
        <span>Command preview</span>
        <code>{commandPreview}</code>
      </div>
    </div>

    <div class="actions">
      <button class="btn btn-secondary" on:click={() => dispatch('close')}>Cancel</button>
      <button class="btn btn-primary" on:click={execute} disabled={!formValid}>Execute</button>
    </div>
  </Modal>
{/if}

<style>
  .command-form {
    max-height: 65vh;
    width: 100%;
    box-sizing: border-box;
    overflow-y: auto;
    overflow-x: hidden;
    padding-right: 4px;
  }

  .description {
    margin: 0 0 16px;
    color: var(--text-muted);
    white-space: pre-line;
  }

  .usage,
  .preview {
    margin-bottom: 16px;
  }

  .usage span,
  .preview span,
  h3 {
    display: block;
    margin: 0 0 7px;
    color: var(--text-color);
    font-size: 0.9em;
    font-weight: 600;
  }

  .usage code,
  .preview code {
    display: block;
    padding: 9px 11px;
    border: 1px solid var(--panel-border);
    border-radius: 4px;
    background: var(--toolbar-bg);
    color: var(--text-color);
    overflow-wrap: anywhere;
  }

  .unavailable {
    margin-bottom: 16px;
    padding: 10px 12px;
    border: 1px solid #8f3737;
    border-radius: 4px;
    background: rgba(143, 55, 55, 0.15);
    color: #ff9b9b;
  }

  section {
    margin-bottom: 18px;
  }

  .flag-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 12px;
  }

  .field {
    display: block;
    margin-bottom: 12px;
  }

  .field > span,
  .boolean-field > span {
    display: block;
    margin-bottom: 5px;
    color: var(--text-color);
    font-size: 0.88em;
  }

  .field strong,
  .boolean-field strong {
    margin-left: 3px;
    color: #ff7070;
  }

  .field small,
  .boolean-field small {
    display: block;
    margin-top: 4px;
    color: var(--text-muted);
    font-weight: normal;
  }

  .field code,
  .boolean-field code {
    margin-left: 5px;
    color: var(--text-muted);
  }

  input[type='text'],
  input[type='password'],
  input[type='number'] {
    width: 100%;
    box-sizing: border-box;
    padding: 8px 10px;
    border: 1px solid var(--panel-border);
    border-radius: 4px;
    background: var(--toolbar-bg);
    color: var(--text-color);
  }

  input:focus {
    border-color: var(--accent-color);
    outline: none;
  }

  .boolean-field {
    display: flex;
    gap: 8px;
    align-items: flex-start;
    min-height: 42px;
    padding: 8px 10px;
    border: 1px solid var(--panel-border);
    border-radius: 4px;
    cursor: pointer;
  }

  .boolean-field input {
    margin-top: 3px;
  }

  .advanced-toggle {
    margin: 0 0 12px;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--text-muted);
    cursor: pointer;
  }

  .advanced-toggle i {
    width: 14px;
    margin-right: 3px;
  }

  .advanced {
    margin-bottom: 16px;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 20px;
  }

  .btn {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
  }

  .btn-primary {
    background: var(--accent-color);
    color: var(--bg-color);
  }

  .btn-primary:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .btn-secondary {
    border: 1px solid var(--panel-border);
    background: var(--toolbar-bg);
    color: var(--text-color);
  }

  @media (max-width: 800px) {
    .command-form {
      min-width: 0;
    }

    .flag-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
