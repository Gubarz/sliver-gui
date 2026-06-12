import '@fortawesome/fontawesome-free/css/all.min.css'
import './style.css'
import { mount } from 'svelte'
import App from './App.svelte'

// Svelte 5 replaced `new App({ target })` with the mount() function.
const app = mount(App, {
  target: document.getElementById('app'),
})

export default app
