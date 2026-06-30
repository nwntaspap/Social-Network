import js from '@eslint/js';
import globals from 'globals';
import prettier from 'eslint-config-prettier';

export default [
  js.configs.recommended, // ESLint's recommended rules
  {
    ignores: ['dist/', 'node_modules/', '**/*.test.js'], // <-- this removes the warning
  },
  {
    files: ['**/*.{js,mjs,cjs}'],
    languageOptions: {
      globals: globals.browser,
    },
    // Add your custom rules here if needed
    rules: {
      // your overrides
    },
  },
  prettier, // Must be last to override formatting rules
];
