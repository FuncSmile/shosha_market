const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

const projectRoot = path.join(__dirname, '../..');
const backendDir = path.join(projectRoot, 'backend');

console.log('🔨 Building Go backend...');

// Download dependencies
console.log('📥 Downloading Go modules...');
const dlResult = spawnSync('go', ['mod', 'download'], {
  cwd: backendDir,
  stdio: 'inherit'
});

if (dlResult.status !== 0) {
  console.error('❌ Failed to download modules');
  process.exit(1);
}

// Build based on platform
const isWindows = process.platform === 'win32';
const isMac = process.platform === 'darwin';
const isLinux = process.platform === 'linux';

let output = isWindows ? 'server.exe' : 'server';

console.log(`🔨 Building for ${process.platform}...`);

const buildArgs = ['build', '-o', output, 'main.go'];
const buildEnv = { ...process.env };

if (isLinux) {
  buildEnv.CGO_ENABLED = '1';
} else if (isWindows) {
  buildEnv.CGO_ENABLED = '0';
}

const result = spawnSync('go', buildArgs, {
  cwd: backendDir,
  stdio: 'inherit',
  env: buildEnv
});

if (result.status !== 0) {
  console.error('❌ Build failed');
  process.exit(1);
}

console.log(`✅ Backend built: ${output}`);
