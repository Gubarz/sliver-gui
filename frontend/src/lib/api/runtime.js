import {
  EventsOn,
  OnFileDrop,
  OnFileDropOff,
  Quit,
  WindowMinimise,
  WindowToggleMaximise,
} from '../../../wailsjs/runtime/runtime.js';

const fileDropListeners = new Set();
let fileDropRegistered = false;

export function onSliverEvent(callback) {
  return EventsOn('sliver-event', callback);
}

export function onShellOutput(callback) {
  return EventsOn('shell-output', callback);
}

export function onAutomationUpdated(callback) {
  return EventsOn('automation-updated', callback);
}

export function onAutomationRun(callback) {
  return EventsOn('automation-run', callback);
}

export function onFileDrop(callback) {
  fileDropListeners.add(callback);
  if (!fileDropRegistered) {
    OnFileDrop((x, y, paths) => {
      for (const listener of fileDropListeners) listener(x, y, paths);
    }, true);
    fileDropRegistered = true;
  }

  return () => {
    fileDropListeners.delete(callback);
    if (fileDropListeners.size === 0 && fileDropRegistered) {
      OnFileDropOff();
      fileDropRegistered = false;
    }
  };
}

export function minimizeWindow() {
  WindowMinimise();
}

export function toggleMaximizeWindow() {
  WindowToggleMaximise();
}

export function quitApplication() {
  Quit();
}
