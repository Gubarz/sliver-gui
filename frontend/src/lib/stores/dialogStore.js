import { writable } from 'svelte/store';

export const dialogStore = writable({
  isOpen: false,
  type: '', // 'alert', 'confirm', 'prompt'
  title: '',
  message: '',
  inputValue: '', // for prompt
  resolve: null
});

export const dialog = {
  alert(message, title = "Alert") {
    return new Promise((resolve) => {
      dialogStore.set({ isOpen: true, type: 'alert', title, message, resolve, inputValue: '' });
    });
  },
  confirm(message, title = "Confirm") {
    return new Promise((resolve) => {
      dialogStore.set({ isOpen: true, type: 'confirm', title, message, resolve, inputValue: '' });
    });
  },
  prompt(message, title = "Input Required", defaultValue = "") {
    return new Promise((resolve) => {
      dialogStore.set({ isOpen: true, type: 'prompt', title, message, resolve, inputValue: defaultValue });
    });
  }
};
