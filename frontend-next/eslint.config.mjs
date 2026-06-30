import { defineConfig, globalIgnores } from 'eslint/config';
import nextVitals from 'eslint-config-next/core-web-vitals';
import nextTs from 'eslint-config-next/typescript';
import prettierPlugin from 'eslint-plugin-prettier';
import prettierConfig from 'eslint-config-prettier';

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    '.next/**',
    'out/**',
    'build/**',
    'next-env.d.ts',
    'node_modules/**',
    'dist/**',
  ]),
  {
    plugins: {
      prettier: prettierPlugin,
    },
    rules: {
      // Run Prettier as an ESLint rule
      'prettier/prettier': 'error',
      // Turn off rules that might conflict with Prettier
      'arrow-body-style': 'off',
      'prefer-arrow-callback': 'off',
    },
  },
  prettierConfig, // This must be last to override conflicting rules
]);

export default eslintConfig;
