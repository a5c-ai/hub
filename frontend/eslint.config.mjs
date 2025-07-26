import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  {
    rules: {
      // Allow any types in specific cases where typing is complex
      "@typescript-eslint/no-explicit-any": "warn",
      // Allow unused variables with underscore prefix
      "@typescript-eslint/no-unused-vars": ["warn", { 
        "argsIgnorePattern": "^_",
        "varsIgnorePattern": "^_" 
      }],
      // Don't require exhaustive deps in useEffect if explicitly ignored
      "react-hooks/exhaustive-deps": "warn"
    }
  }
];

export default eslintConfig;
