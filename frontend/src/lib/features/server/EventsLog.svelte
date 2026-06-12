<script>
  export let events = [];

  function fmtTime(ts) {
    if (!ts) return "-";
    return new Date(ts).toLocaleTimeString();
  }
</script>

<div class="events-log-container">
  <div class="panel-body">
    {#if events.length === 0}
      <p style="color: var(--text-muted)">No events recorded yet.</p>
    {:else}
      <table class="data-table">
        <thead>
          <tr>
            <th>Time</th>
            <th>Type</th>
            <th>Details</th>
          </tr>
        </thead>
        <tbody>
          {#each events.slice().reverse() as ev}
            <tr>
              <td class="mono" style="width: 100px;">{fmtTime(ev.time)}</td>
              <td class="mono">{ev.type}</td>
              <td class="mono">
                {#if ev.sessionID}
                  [Session {ev.sessionID.substring(0,8)}] {ev.username}@{ev.hostname}
                {/if}
                {#if ev.job}
                  [Job {ev.job}]
                {/if}
                {#if ev.data}
                  {ev.data}
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<style>
  .events-log-container {
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .panel-body {
    padding: 16px 20px;
    overflow-y: auto;
    flex: 1;
  }
</style>
