<script>
  import { createEventDispatcher } from 'svelte';
  import PanelShell from './PanelShell.svelte';

  export let title = "Modal Title";
  export let width = "450px";

  const dispatch = createEventDispatcher();

  function close() {
    dispatch('close');
  }
</script>

<PanelShell {title} {width} bodyPadding="20px" on:close={close}>
  <slot></slot>
</PanelShell>

<style>
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
