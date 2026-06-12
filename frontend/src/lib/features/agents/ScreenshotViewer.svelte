<script>
  import { TakeScreenshot } from '../../api/agents.js';
  import { errorMessage } from '../../utils/errors.js';

  export let sessionID = "";

  // Keep each session's last screenshot so switching tabs and coming back shows
  // it again instead of a blank panel.
  const shotBySession = new Map();

  let screenshotBase64 = "";
  let loading = false;
  let error = "";

  // Restore this session's screenshot (if we have one) when switching to it.
  let lastSession = null;
  $: if (sessionID !== lastSession) {
    lastSession = sessionID;
    screenshotBase64 = shotBySession.get(sessionID) || "";
    error = "";
  }

  async function takeScreenshot() {
    loading = true;
    error = "";
    try {
      screenshotBase64 = await TakeScreenshot(sessionID);
      shotBySession.set(sessionID, screenshotBase64);
    } catch (err) {
      error = errorMessage(err);
    } finally {
      loading = false;
    }
  }
</script>

<div class="tab-wrapper">
  <div class="tab-header">
    Screenshot
    <button class="btn" style="float: right; margin-top: -4px;" on:click={takeScreenshot}>Retake</button>
  </div>
  
  <div class="tab-content" style="text-align: center; padding: 20px;">
    {#if loading}
      <div>Capturing screen...</div>
    {:else if error}
      <div style="color: var(--danger-color);">{error}</div>
    {:else if screenshotBase64}
      <img src="data:image/png;base64,{screenshotBase64}" alt="Target Screenshot" style="border: 1px solid var(--border-color); box-shadow: 0 4px 8px rgba(0,0,0,0.5);" />
    {:else}
      <div style="margin-top: 100px;">
        <button class="btn btn-primary" style="padding: 10px 20px; font-size: 1.2em;" on:click={takeScreenshot}>Capture Screenshot</button>
      </div>
    {/if}
  </div>
</div>

<style>
  .tab-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
    background-color: var(--bg-color);
  }
  .tab-header {
    background-color: var(--header-bg);
    padding: 10px 15px;
    font-size: 0.9em;
    border-bottom: 1px solid var(--border-color);
  }
  .tab-content {
    flex: 1;
    overflow: auto;
  }
</style>
