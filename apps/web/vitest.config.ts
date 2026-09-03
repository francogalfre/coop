import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["**/__tests__.ts", "**/__tests__.*.ts"],
    exclude: ["node_modules/**", ".next/**"],
  },
});
