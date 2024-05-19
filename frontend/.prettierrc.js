export default {
  useTabs: false,
  singleQuote: false,
  trailingComma: "all",
  printWidth: 100,
  plugins: [
    "prettier-plugin-svelte",
    "@trivago/prettier-plugin-sort-imports",
    "prettier-plugin-packagejson",
  ],
  overrides: [
    {
      files: "**/*.svx",
      options: { parser: "markdown" },
    },
    {
      files: "**/*.ts",
      options: { parser: "typescript" },
    },
  ],
  importOrderSeparation: true,
  importOrderSortSpecifiers: true,
};
