#!/usr/bin/env node

const { execFileSync } = require("child_process");
const { join } = require("path");

const ext = process.platform === "win32" ? ".exe" : "";
const bin = join(__dirname, "bin", `rancher-mcp-server${ext}`);

try {
  execFileSync(bin, process.argv.slice(2), { stdio: "inherit" });
} catch (err) {
  if (err.status !== undefined) {
    process.exit(err.status);
  }
  console.error(`Failed to run rancher-mcp-server: ${err.message}`);
  console.error("Try reinstalling: npm install -g rancher-mcp-server");
  process.exit(1);
}
