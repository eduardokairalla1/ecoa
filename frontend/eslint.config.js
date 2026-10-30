/**
 * ESLint setup.
 */

// --- IMPORTS ---
import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

// --- CODE ---
export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      globals: globals.browser,
    },
    rules: {
      // an import with one name cannot be wrapped: Prettier keeps it whole
      'max-len': [
        'error',
        { code: 80, ignoreUrls: true, ignorePattern: '^import\\s.+\\sfrom\\s' },
      ],
    },
  },
])
