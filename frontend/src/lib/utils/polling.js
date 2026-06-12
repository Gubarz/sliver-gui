export function startPolling(task, interval, { immediate = true } = {}) {
  let stopped = false;
  let timeout;

  async function run() {
    try {
      await task();
    } finally {
      if (!stopped) timeout = setTimeout(run, interval);
    }
  }

  if (immediate) {
    run();
  } else {
    timeout = setTimeout(run, interval);
  }

  return () => {
    stopped = true;
    clearTimeout(timeout);
  };
}
