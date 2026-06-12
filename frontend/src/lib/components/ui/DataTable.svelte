<script>
  export let data = [];
  export let columns = [];
  export let filterable = true;
  export let defaultSortKey = "";
  export let compact = false;

  let filter = "";
  let sortKey = defaultSortKey || (columns.length > 0 ? columns[0].key : "");
  let sortDir = 1; // 1 asc, -1 desc
  $: compactWidth = columns.reduce((total, column) => total + column.width, 0);

  function sortBy(key) {
    if (sortKey === key) sortDir = -sortDir;
    else { sortKey = key; sortDir = 1; }
  }

  $: filtered = data.filter((item) => {
    if (!filter) return true;
    const hay = columns.map(c => item[c.key] || "").join(" ").toLowerCase();
    return hay.includes(filter.toLowerCase());
  });

  $: rows = [...filtered].sort((a, b) => {
    let av = (a[sortKey] ?? "").toString().toLowerCase();
    let bv = (b[sortKey] ?? "").toString().toLowerCase();
    return av < bv ? -sortDir : av > bv ? sortDir : 0;
  });

  let resizingCol = null;
  let startX = 0;
  let startWidth = 0;

  function startResize(e, index) {
    e.stopPropagation();
    resizingCol = index;
    startX = e.clientX;
    startWidth = columns[index].width;
    window.addEventListener('mousemove', doResize);
    window.addEventListener('mouseup', stopResize);
  }

  function doResize(e) {
    if (resizingCol !== null) {
      columns[resizingCol].width = Math.max(compact ? 32 : 50, startWidth + (e.clientX - startX));
      columns = [...columns];
    }
  }

  function stopResize() {
    resizingCol = null;
    window.removeEventListener('mousemove', doResize);
    window.removeEventListener('mouseup', stopResize);
  }
</script>

<div class="table-wrap" class:resizing={resizingCol !== null}>
  {#if filterable}
    <div class="table-toolbar">
      <i class="fas fa-search"></i>
      <input class="search" placeholder="Filter..." bind:value={filter} />
      <span class="count">{rows.length} item{rows.length === 1 ? '' : 's'}</span>
    </div>
  {/if}
  <div class="table-container">
    <table
      class="data-table"
      class:compact
      style={compact ? `width: 100%; min-width: ${compactWidth}px;` : ""}
    >
      {#if compact}
        <colgroup>
          {#each columns as col}
            <col style="width: {col.width}px; min-width: {col.width}px; max-width: {col.width}px;" />
          {/each}
        </colgroup>
      {/if}
      <thead>
        <tr>
          {#each columns as col, index}
            <th
              title={col.label}
              on:click={() => sortBy(col.key)}
              class="sortable"
              style="width: {col.width}px; min-width: {col.width}px; max-width: {col.width}px;"
            >
              <div class="th-content">
                <span class="th-label">{col.label}</span>
                {#if sortKey === col.key}<span class="arrow">{sortDir === 1 ? '▲' : '▼'}</span>{/if}
              </div>
              <!-- svelte-ignore a11y-click-events-have-key-events -->
              <!-- svelte-ignore a11y-no-static-element-interactions -->
              <div 
                class="col-resizer" 
                on:mousedown={(e) => startResize(e, index)}
                on:click|stopPropagation
              ></div>
            </th>
          {/each}
        </tr>
      </thead>
      <tbody>
        <slot {rows} {columns}></slot>
      </tbody>
    </table>
  </div>
</div>
