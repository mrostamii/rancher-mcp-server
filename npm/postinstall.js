#!/usr/bin/env node

const { createWriteStream, mkdirSync, chmodSync, existsSync } = require("fs");
const { join } = require("path");
const https = require("https");
const { version } = require("./package.json");

const REPO = "mrostamii/rancher-mcp-server";
const BIN_NAME = "rancher-mcp-server";

const PLATFORM_MAP = {
  "darwin-arm64": "darwin-arm64",
  "darwin-x64": "darwin-x64",
  "linux-arm64": "linux-arm64",
  "linux-x64": "linux-x64",
  "win32-x64": "windows-x64",
};

function getPlatformKey() {
  const platform = process.platform;
  const arch = process.arch;
  return `${platform}-${arch}`;
}

function getBinaryName(platformKey) {
  const suffix = PLATFORM_MAP[platformKey];
  if (!suffix) return null;
  const ext = process.platform === "win32" ? ".exe" : "";
  return `${BIN_NAME}-${suffix}${ext}`;
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = createWriteStream(dest);
    https
      .get(url, (response) => {
        if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
          file.close();
          return downloadFile(response.headers.location, dest).then(resolve, reject);
        }
        if (response.statusCode !== 200) {
          file.close();
          return reject(new Error(`Download failed: HTTP ${response.statusCode} from ${url}`));
        }
        response.pipe(file);
        file.on("finish", () => {
          file.close(resolve);
        });
      })
      .on("error", (err) => {
        file.close();
        reject(err);
      });
  });
}

async function main() {
  const platformKey = getPlatformKey();
  const binaryAsset = getBinaryName(platformKey);

  if (!binaryAsset) {
    console.error(
      `Unsupported platform: ${process.platform}-${process.arch}\n` +
        `Supported: ${Object.keys(PLATFORM_MAP).join(", ")}`
    );
    process.exit(1);
  }

  const binDir = join(__dirname, "bin");
  mkdirSync(binDir, { recursive: true });

  const ext = process.platform === "win32" ? ".exe" : "";
  const dest = join(binDir, `${BIN_NAME}${ext}`);

  if (existsSync(dest)) {
    console.log(`${BIN_NAME} binary already exists, skipping download.`);
    return;
  }

  const tag = `v${version}`;
  const url = `https://github.com/${REPO}/releases/download/${tag}/${binaryAsset}`;

  console.log(`Downloading ${BIN_NAME} ${tag} for ${platformKey}...`);
  console.log(`  ${url}`);

  try {
    await downloadFile(url, dest);
    if (process.platform !== "win32") {
      chmodSync(dest, 0o755);
    }
    console.log(`Installed ${BIN_NAME} to ${dest}`);
  } catch (err) {
    console.error(`Failed to download ${BIN_NAME}:`, err.message);
    console.error(
      `\nYou can manually download from:\n  https://github.com/${REPO}/releases/tag/${tag}`
    );
    process.exit(1);
  }
}

main();
