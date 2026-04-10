"use strict";

const https = require("https");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const { execSync } = require("child_process");

const REPO = "shinagawa-web/gomarklint";

const PLATFORM_MAP = {
  darwin: "Darwin",
  linux: "Linux",
  win32: "Windows",
};

const ARCH_MAP = {
  x64: "x86_64",
  arm64: "arm64",
};

function getPlatformInfo() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    throw new Error(
      `Unsupported platform: ${process.platform}-${process.arch}\n` +
        `Supported: darwin-x64, darwin-arm64, linux-x64, linux-arm64, win32-x64, win32-arm64`
    );
  }

  const ext = process.platform === "win32" ? "zip" : "tar.gz";
  const archiveName = `gomarklint_${platform}_${arch}.${ext}`;

  return { platform, arch, ext, archiveName };
}

function download(url) {
  return new Promise((resolve, reject) => {
    const request = (url) => {
      https
        .get(url, { headers: { "User-Agent": "gomarklint-npm" } }, (res) => {
          if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
            request(res.headers.location);
            return;
          }
          if (res.statusCode !== 200) {
            reject(new Error(`Download failed: HTTP ${res.statusCode} for ${url}`));
            return;
          }
          const chunks = [];
          res.on("data", (chunk) => chunks.push(chunk));
          res.on("end", () => resolve(Buffer.concat(chunks)));
          res.on("error", reject);
        })
        .on("error", reject);
    };
    request(url);
  });
}

function verifyChecksum(data, expected) {
  const actual = crypto.createHash("sha256").update(data).digest("hex");
  if (actual !== expected) {
    throw new Error(
      `Checksum mismatch!\n  Expected: ${expected}\n  Actual:   ${actual}`
    );
  }
}

async function main() {
  const pkg = JSON.parse(
    fs.readFileSync(path.join(__dirname, "package.json"), "utf8")
  );
  const version = pkg.version;

  if (version === "0.0.0") {
    console.log("Skipping binary download for development version.");
    return;
  }

  const { archiveName } = getPlatformInfo();
  const baseUrl = `https://github.com/${REPO}/releases/download/v${version}`;

  console.log(`Downloading gomarklint v${version} (${archiveName})...`);

  // Download checksums and archive in parallel
  const [checksumsText, archiveData] = await Promise.all([
    download(`${baseUrl}/gomarklint_${version}_checksums.txt`).then((buf) => buf.toString("utf8")),
    download(`${baseUrl}/${archiveName}`),
  ]);

  // Verify checksum
  const checksumLine = checksumsText
    .split("\n")
    .find((line) => line.includes(archiveName));

  if (!checksumLine) {
    throw new Error(`Checksum not found for ${archiveName}`);
  }

  const expectedHash = checksumLine.split(/\s+/)[0];
  verifyChecksum(archiveData, expectedHash);
  console.log("Checksum verified.");

  // Write archive to temp file and extract
  const archivePath = path.join(__dirname, archiveName);
  fs.writeFileSync(archivePath, archiveData);

  try {
    execSync(`tar -xf "${archiveName}"`, { cwd: __dirname, stdio: "pipe" });
  } catch (e) {
    throw new Error(`Failed to extract archive: ${e.message}`);
  }

  // Clean up archive
  fs.unlinkSync(archivePath);

  // Set executable permission on non-Windows
  if (process.platform !== "win32") {
    const binPath = path.join(__dirname, "gomarklint");
    fs.chmodSync(binPath, 0o755);
  }

  console.log("gomarklint installed successfully.");
}

main().catch((err) => {
  console.error(`Error: ${err.message}`);
  process.exit(1);
});
