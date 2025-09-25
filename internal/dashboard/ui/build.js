import * as esbuild from "esbuild";
import { readFileSync, writeFileSync } from "fs";
import { execSync } from "child_process";

const isWatch = process.argv.includes("--watch");

// Build CSS with Tailwind
console.log("Building CSS...");
execSync("npx tailwindcss -i ./src/styles.css -o ../static/styles.css", {
  stdio: "inherit",
});

const buildOptions = {
  entryPoints: ["src/index.tsx"],
  bundle: true,
  minify: !isWatch,
  sourcemap: isWatch,
  outfile: "../static/dashboard.js",
  format: "iife",
  platform: "browser",
  target: ["es2020"],
  loader: {
    ".tsx": "tsx",
    ".ts": "ts",
    ".css": "text",
  },
  jsxFactory: "h",
  jsxFragment: "Fragment",
  inject: ["./src/preact-shim.js"],
  alias: {
    react: "@preact/compat",
    "react-dom": "@preact/compat",
    "react/jsx-runtime": "preact/jsx-runtime",
  },
  define: {
    "process.env.NODE_ENV": isWatch ? '"development"' : '"production"',
  },
};

if (isWatch) {
  const ctx = await esbuild.context(buildOptions);
  await ctx.watch();
  console.log("Watching for changes...");
} else {
  await esbuild.build(buildOptions);
  console.log("Build complete");
}
