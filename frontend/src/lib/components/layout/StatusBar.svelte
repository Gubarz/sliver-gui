<script>
  export let serverVersion = '';
  export let sessionCount = 0;
  export let beaconCount = 0;
</script>

<div class="status-bar">
  <span class="connected-indicator">●</span> Connected
  {#if serverVersion}&nbsp;·&nbsp; Server v{serverVersion}{/if}
  &nbsp;·&nbsp; {sessionCount} session{sessionCount === 1 ? '' : 's'}
  &nbsp;·&nbsp; {beaconCount} beacon{beaconCount === 1 ? '' : 's'}
</div>

<style>
  .connected-indicator {
    color: var(--success-color);
  }
</style>
