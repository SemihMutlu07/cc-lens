#!/usr/bin/env node
'use strict';

// Thin launcher for wrapminal: downloads the matching prebuilt Go binary from
// GitHub Releases on first run, caches it, then runs it. No runtime deps.
// Pattern mirrors esbuild/swc npm wrappers. Local-first: nothing is uploaded.

const fs = require('fs');
const os = require('os');
const path = require('path');
const https = require('https');
const { spawn } = require('child_process');

const REPO = 'SemihMutlu07/wrapminal';
const VERSION = require('../package.json').version;

// process.platform/arch -> GitHub release asset name (see .github/workflows/release.yml)
function assetName() {
  const archMap = { x64: 'amd64', arm64: 'arm64' };
  const goArch = archMap[process.arch];

  if (process.platform === 'win32') {
    return goArch === 'amd64' ? 'wrapminal-windows-amd64.exe' : null;
  }
  if (process.platform === 'darwin' || process.platform === 'linux') {
    return goArch ? `wrapminal-${process.platform}-${goArch}` : null;
  }
  return null;
}

function cacheDir() {
  const base =
    process.env.WRAPMINAL_CACHE_DIR || path.join(os.homedir(), '.cache', 'wrapminal');
  fs.mkdirSync(base, { recursive: true });
  return base;
}

// GitHub release downloads redirect to a signed CDN URL, so follow redirects.
function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https
      .get(url, { headers: { 'User-Agent': 'wrapminal-npx' } }, (res) => {
        if (
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location
        ) {
          res.resume();
          file.close();
          fs.rmSync(dest, { force: true });
          return download(res.headers.location, dest).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          res.resume();
          file.close();
          fs.rmSync(dest, { force: true });
          return reject(
            new Error(`download failed: HTTP ${res.statusCode} for ${url}`)
          );
        }
        res.pipe(file);
        file.on('finish', () => file.close(() => resolve()));
      })
      .on('error', (err) => {
        file.close();
        fs.rmSync(dest, { force: true });
        reject(err);
      });
  });
}

async function ensureBinary() {
  const asset = assetName();
  if (!asset) {
    console.error(
      `wrapminal: unsupported platform ${process.platform}/${process.arch}.\n` +
        `Grab a binary from https://github.com/${REPO}/releases/latest`
    );
    process.exit(1);
  }

  const binPath = path.join(cacheDir(), `${VERSION}-${asset}`);
  if (fs.existsSync(binPath) && fs.statSync(binPath).size > 0) {
    return binPath;
  }

  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${asset}`;
  process.stderr.write(`wrapminal: downloading ${asset} (v${VERSION})...\n`);

  const tmp = `${binPath}.${process.pid}.tmp`;
  await download(url, tmp);
  if (process.platform !== 'win32') fs.chmodSync(tmp, 0o755);
  fs.renameSync(tmp, binPath);
  return binPath;
}

ensureBinary()
  .then((binPath) => {
    const child = spawn(binPath, process.argv.slice(2), {
      stdio: 'inherit',
      env: process.env,
    });
    child.on('error', (err) => {
      console.error(`wrapminal: failed to launch: ${err.message}`);
      process.exit(1);
    });
    child.on('exit', (code, signal) => {
      if (signal) process.kill(process.pid, signal);
      else process.exit(code === null ? 0 : code);
    });
  })
  .catch((err) => {
    console.error(`wrapminal: ${err.message}`);
    process.exit(1);
  });
