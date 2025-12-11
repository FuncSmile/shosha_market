const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

const projectRoot = path.join(__dirname, '../..');
const backendDir = path.join(projectRoot, 'backend');
const backendResourcesDir = path.join(__dirname, '../resources/backend');

console.log('📦 Building backend binaries...');

// Create resources directory if not exists
if (!fs.existsSync(backendResourcesDir)) {
  fs.mkdirSync(backendResourcesDir, { recursive: true });
}

// Download Go modules once
console.log('📥 Downloading Go modules...');
const dlResult = spawnSync('go', ['mod', 'download'], { cwd: backendDir, stdio: 'inherit' });
if (dlResult.status !== 0) {
  console.error('❌ Failed to download Go modules');
  process.exit(1);
}

// Build function
function buildBinary(targetOS, targetArch, outputName) {
  console.log(`\n🔨 Building ${outputName} for ${targetOS}/${targetArch}...`);
  
  const buildEnv = { ...process.env };
  buildEnv.GOOS = targetOS;
  buildEnv.GOARCH = targetArch;
  buildEnv.CGO_ENABLED = '1';
  
  // Set cross-compilation tools for Windows
  if (targetOS === 'windows' && process.platform === 'linux') {
    buildEnv.CC = 'x86_64-w64-mingw32-gcc';
    buildEnv.CXX = 'x86_64-w64-mingw32-g++';
    console.log('  Using mingw-w64 cross-compiler');
  }
  
  const buildArgs = [
    'build',
    '-o', outputName,
    'main.go'
  ];
  
  const buildResult = spawnSync('go', buildArgs, {
    cwd: backendDir,
    stdio: 'inherit',
    env: buildEnv
  });
  
  if (buildResult.status !== 0) {
    console.error(`❌ Failed to build ${outputName}`);
    return false;
  }
  
  const binaryPath = path.join(backendDir, outputName);
  const targetPath = path.join(backendResourcesDir, outputName);
  
  if (fs.existsSync(binaryPath)) {
    fs.copyFileSync(binaryPath, targetPath);
    console.log(`✓ Built and copied: ${outputName}`);
    return true;
  }
  
  return false;
}

// Build for current platform
const isWindows = process.platform === 'win32';
const isMac = process.platform === 'darwin';
const currentBinary = isWindows ? 'server.exe' : 'server';
const currentOS = isWindows ? 'windows' : (isMac ? 'darwin' : 'linux');
const currentArch = 'amd64';

// Always build for current platform
buildBinary(currentOS, currentArch, currentBinary);

// On Linux, also try to build Windows binary (for cross-platform package)
if (process.platform === 'linux') {
  console.log('\n🪟 Attempting to build Windows binary (requires mingw-w64)...');
  const winBuilt = buildBinary('windows', 'amd64', 'server.exe');
  
  if (!winBuilt) {
    console.log('⚠️  Windows build failed. Install mingw-w64:');
    console.log('    Ubuntu/Debian: sudo apt install gcc-mingw-w64-x86-64');
    console.log('    Arch/Manjaro: sudo pacman -S mingw-w64-gcc');
    console.log('    Or build on Windows/GitHub Actions');
  }
}

// Ensure .env exists for production (copy from .env.production if .env doesn't exist)
const envPath = path.join(backendDir, '.env');
const envProductionPath = path.join(backendDir, '.env.production');

if (!fs.existsSync(envPath) && fs.existsSync(envProductionPath)) {
  console.log('📝 Creating .env from .env.production...');
  fs.copyFileSync(envProductionPath, envPath);
  console.log('✓ .env created for production build');
}

console.log('✅ Backend build complete');
