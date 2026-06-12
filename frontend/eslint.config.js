import js from '@eslint/js';
import globals from 'globals';
import svelte from 'eslint-plugin-svelte';

export default [
  {
    ignores: ['dist/**', 'node_modules/**', 'wailsjs/**'],
  },
  js.configs.recommended,
  ...svelte.configs.recommended,
  {
    files: ['src/**/*.{js,svelte}', '*.js'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    rules: {
      'no-empty': ['error', { allowEmptyCatch: true }],
      'svelte/require-each-key': 'off',
      'svelte/prefer-svelte-reactivity': 'off',
      'svelte/infinite-reactive-loop': 'off',
      'svelte/no-immutable-reactive-statements': 'off',
      'svelte/no-at-html-tags': 'warn',
    },
  },
  {
    files: ['**/*.svelte'],
    rules: {
      'no-useless-assignment': 'off',
    },
  },
];
