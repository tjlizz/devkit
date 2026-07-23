import { defineConfig, globalIgnores } from "eslint/config";
import tseslint from "./node_modules/eslint-config-next/node_modules/typescript-eslint/dist/index.js";

export default defineConfig([
  globalIgnores([".next/**", "out/**", "build/**", "next-env.d.ts"]),
  {
    files: ["**/*.{ts,tsx}"],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        ecmaFeatures: { jsx: true },
        sourceType: "module",
      },
    },
    rules: {
      "constructor-super": "error",
      "for-direction": "error",
      "getter-return": "error",
      "no-async-promise-executor": "error",
      "no-constant-binary-expression": "error",
      "no-debugger": "error",
      "no-dupe-args": "error",
      "no-dupe-class-members": "error",
      "no-dupe-else-if": "error",
      "no-dupe-keys": "error",
      "no-duplicate-imports": "error",
      "no-fallthrough": "error",
      "no-import-assign": "error",
      "no-new-native-nonconstructor": "error",
      "no-promise-executor-return": "error",
      "no-self-assign": "error",
      "no-unreachable": "error",
      "no-unreachable-loop": "error",
      "no-unused-private-class-members": "error",
      "no-useless-assignment": "error",
      "no-useless-backreference": "error",
      "require-yield": "error"
    },
  },
]);
