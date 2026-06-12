<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { GetScreenshotData, listLoot } from '../../api/server.js';
  import PanelShell from '../../components/ui/PanelShell.svelte';
  import { errorMessage } from '../../utils/errors.js';
  
  const dispatch = createEventDispatcher();
  let screenshots = [];
  let loading = true;
  let errorMsg = "";
  
  // Lightbox state
  let selectedImage = null;

  onMount(async () => {
    try {
      const allLoot = await listLoot();
      
      // Filter for images
      const imageLoot = allLoot.filter(l => {
        if (!l || !l.Name) return false;
        const name = l.Name.toLowerCase();
        return name.endsWith('.png') || name.endsWith('.jpg') || name.endsWith('.jpeg');
      });

      // Fetch data for each image
      for (const item of imageLoot) {
        try {
          const dataURI = await GetScreenshotData(item.ID);
          screenshots = [...screenshots, { ...item, dataURI }];
        } catch (e) {
          console.error("Failed to load image", item.ID, e);
        }
      }
    } catch (e) {
      console.error(e);
      errorMsg = errorMessage(e);
    } finally {
      loading = false;
    }
  });

  function close() {
    dispatch('close');
  }

  function openLightbox(img) {
    selectedImage = img;
  }

  function closeLightbox() {
    selectedImage = null;
  }
</script>

<PanelShell
  title="Screenshot Gallery"
  icon="fa-images"
  width="900px"
  bodyPadding="20px"
  bodyBackground="var(--bg-color)"
  on:close={close}
>
  {#if loading}
    <div class="message">Loading screenshots from Loot...</div>
  {:else if errorMsg}
    <div class="message error">{errorMsg}</div>
  {:else if screenshots.length === 0}
    <div class="message">No screenshots found in Loot.</div>
  {:else}
    <div class="grid">
      {#each screenshots as img}
        <button type="button" class="card" on:click={() => openLightbox(img)}>
          <img src={img.dataURI} alt={img.Name} />
          <span class="card-title">{img.Name}</span>
        </button>
      {/each}
    </div>
  {/if}
</PanelShell>

{#if selectedImage}
  <div class="lightbox" role="dialog" aria-modal="true" aria-label={selectedImage.Name}>
    <button type="button" class="lightbox-dismiss" aria-label="Close image" on:click={closeLightbox}></button>
    <figure>
      <img src={selectedImage.dataURI} alt={selectedImage.Name} />
      <figcaption>{selectedImage.Name} - {selectedImage.Size} bytes</figcaption>
    </figure>
    <button type="button" class="lightbox-close" aria-label="Close image" on:click={closeLightbox}>×</button>
  </div>
{/if}

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 20px;
  }
  .card {
    padding: 0;
    color: inherit;
    font: inherit;
    text-align: inherit;
    background: var(--toolbar-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
    display: flex;
    flex-direction: column;
  }
  .card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.5);
    border-color: var(--accent-color);
  }
  .card img {
    width: 100%;
    height: 150px;
    object-fit: cover;
    background-color: #000;
  }
  .card-title {
    padding: 10px;
    font-size: 0.9em;
    color: var(--text-color);
    text-overflow: ellipsis;
    overflow: hidden;
    white-space: nowrap;
    text-align: center;
  }
  .message {
    padding: 40px;
    text-align: center;
    color: var(--text-muted);
  }
  .error {
    color: var(--danger-color);
  }
  
  /* Lightbox */
  .lightbox {
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0, 0, 0, 0.9);
    z-index: 10000;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    backdrop-filter: blur(5px);
  }
  .lightbox-dismiss {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    padding: 0;
    border: 0;
    background: transparent;
    cursor: default;
  }
  .lightbox figure {
    position: relative;
    z-index: 1;
    display: flex;
    max-width: 90%;
    max-height: 90%;
    flex-direction: column;
    align-items: center;
    margin: 0;
    pointer-events: none;
  }
  .lightbox img {
    max-width: 100%;
    max-height: 85vh;
    box-shadow: 0 0 20px rgba(0,0,0,0.8);
    border: 2px solid var(--border-color);
  }
  .lightbox figcaption {
    margin-top: 15px;
    color: #fff;
    font-size: 1.1em;
  }
  .lightbox-close {
    position: absolute;
    top: 20px;
    right: 30px;
    background: none;
    border: none;
    color: #fff;
    font-size: 40px;
    cursor: pointer;
    z-index: 2;
  }
</style>
