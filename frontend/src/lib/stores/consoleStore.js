import { writable } from 'svelte/store';
import { ListCommands, RunSessionCommand } from '../api/console.js';
import { errorMessage } from '../utils/errors.js';

// Per-session console state, keyed by sessionID. Because each session has its
// own entry here (and every command is an RPC scoped to that sessionID), state
// never bleeds between sessions — switching tabs just shows a different entry.
//
//   { lines: [{type:'input'|'output'|'error', text}], history: string[],
//     histIdx: number, busy: boolean }
const sessions = new Map();

// A simple version counter components subscribe to so they re-render when any
// session's buffer changes (mutating Map entries in place isn't reactive).
export const tick = writable(0);
function notify() {
  tick.update((n) => n + 1);
}

// The real implant command list, fetched once from the backend (which derives
// it from Sliver's actual command tree). Used for command-name tab completion.
let commandsPromise = null;
export function getCommands() {
  if (!commandsPromise) {
    commandsPromise = ListCommands().catch(() => []);
  }
  return commandsPromise;
}

export function getSession(id) {
  const key = id || '_';
  if (!sessions.has(key)) {
    sessions.set(key, { lines: [], history: [], histIdx: 0, busy: false });
  }
  return sessions.get(key);
}

export function pushLine(id, line) {
  getSession(id).lines.push(line);
  notify();
}

function clearSession(id) {
  getSession(id).lines = [];
  notify();
}

// dispatchCommand runs a command against a specific session and records the
// input + output into that session's buffer. Safe to call from anywhere (the
// typed console, context menus, dynamic modals).
export async function dispatchCommand(id, raw) {
  const cmd = (raw || '').trim();
  if (!cmd) return;

  const s = getSession(id);
  if (s.busy) return;

  // History bookkeeping.
  s.history.push(cmd);
  s.histIdx = s.history.length;

  // Client-only conveniences.
  if (cmd === 'clear' || cmd === 'cls') {
    clearSession(id);
    return;
  }

  pushLine(id, { type: 'input', text: cmd });
  s.busy = true;
  notify();

  try {
    const out = await RunSessionCommand(id, cmd);
    if (out !== '') {
      pushLine(id, { type: 'output', text: out });
    }
  } catch (e) {
    pushLine(id, { type: 'error', text: errorMessage(e) });
  } finally {
    s.busy = false;
    notify();
  }
}
