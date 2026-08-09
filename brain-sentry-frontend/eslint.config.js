import js from "@eslint/js";
import jsxA11y from "eslint-plugin-jsx-a11y";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import tseslint from "typescript-eslint";

const asWarnings = (rules) => Object.fromEntries(
  Object.entries(rules).map(([name, setting]) => [name, setting === "off" || setting === 0 ? "off" : "warn"]),
);

export default tseslint.config(
  { ignores: ["dist/**", "node_modules/**", "playwright-report/**", "test-results/**"] },
  { rules: asWarnings(js.configs.recommended.rules) },
  ...tseslint.configs.recommended.map((config) => ({ ...config, rules: asWarnings(config.rules ?? {}) })),
  {
    files: ["**/*.{ts,tsx}"],
    plugins: {
      react,
      "react-hooks": reactHooks,
      "jsx-a11y": jsxA11y,
    },
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    settings: { react: { version: "detect" } },
    rules: {
      ...asWarnings(react.configs.recommended.rules),
      ...asWarnings(reactHooks.configs.recommended.rules),
      ...asWarnings(jsxA11y.configs.recommended.rules),
      "react/react-in-jsx-scope": "off",
      "react/prop-types": "off",
      "@typescript-eslint/no-explicit-any": "warn",
    },
  },
  {
    files: ["e2e/**/*.{ts,tsx}"],
    rules: { "react-hooks/rules-of-hooks": "off" },
  },
);
