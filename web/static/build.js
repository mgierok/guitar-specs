#!/usr/bin/env node

const esbuild = require('esbuild');
const crypto = require('crypto');
const fs = require('fs');
const path = require('path');

// Calculate SRI hash for a file
function calculateSRI(filePath) {
  const content = fs.readFileSync(filePath);
  const hash = crypto.createHash('sha384').update(content).digest('base64');
  return `sha384-${hash}`;
}

// Generate cache-busted filename with hash
function generateHashedFilename(filePath) {
  const content = fs.readFileSync(filePath);
  const hash = crypto.createHash('sha256').update(content).digest('hex').substring(0, 8);
  
  const dir = path.dirname(filePath);
  const ext = path.extname(filePath);
  const base = path.basename(filePath, ext);
  
  return {
    hashedPath: path.join(dir, `${base}.${hash}${ext}`),
    hash: hash
  };
}

// Build configuration
const buildConfig = {
  entryPoints: ['web/static/src/main.js'],
  bundle: true,
  minify: true,
  outdir: 'web/static/dist/js',
  format: 'esm',
  outExtension: { '.js': '.js' },
};

// Run esbuild build
async function build() {
  try {
    console.log('→ Building JavaScript with esbuild...');
    const result = await esbuild.build(buildConfig);
    
    if (result.errors.length > 0) {
      console.error('❌ Build failed:', result.errors);
      process.exit(1);
    }
    
    console.log('✓ JavaScript build completed');
    
    // Process built files for SRI and cache busting
    await processBuiltFiles();
    
  } catch (error) {
    console.error('❌ Build error:', error);
    process.exit(1);
  }
}

// Process built files to add SRI and cache busting
async function processBuiltFiles() {
  const jsDir = path.join(process.cwd(), 'web/static/dist/js');
  const cssDir = path.join(process.cwd(), 'web/static/dist/css');
  
  // Clean up previous hashed files and manifest
  cleanupPreviousFiles(jsDir);
  cleanupPreviousFiles(cssDir);
  
  const manifest = {};
  
  // Process JavaScript files
  const jsFiles = fs.readdirSync(jsDir).filter(file => file.endsWith('.js') && (file.match(/\./g) || []).length === 1);
  for (const file of jsFiles) {
    const filePath = path.join(jsDir, file);
    
    // Generate hashed filename
    const { hashedPath, hash } = generateHashedFilename(filePath);
    
    // Calculate SRI hash
    const sri = calculateSRI(filePath);
    
    // Rename file with hash
    fs.renameSync(filePath, hashedPath);
    
    // Add to manifest
    const originalPath = `/static/dist/js/${file}`;
    manifest[originalPath] = {
      hashed: path.basename(hashedPath),
      sri: sri,
      hash: hash,
      path: `/static/dist/js/${path.basename(hashedPath)}`
    };
    
    console.log(`✓ Generated JS: ${path.basename(hashedPath)} (SRI: ${sri})`);
  }
  
  // Process CSS files (already built by Tailwind CLI)
  const cssFiles = fs.readdirSync(cssDir).filter(file => file.endsWith('.css') && (file.match(/\./g) || []).length === 1);
  for (const file of cssFiles) {
    const filePath = path.join(cssDir, file);
    
    // Generate hashed filename
    const { hashedPath, hash } = generateHashedFilename(filePath);
    
    // Calculate SRI hash
    const sri = calculateSRI(filePath);
    
    // Rename file with hash
    fs.renameSync(filePath, hashedPath);
    
    // Add to manifest
    const originalPath = `/static/dist/css/${file}`;
    manifest[originalPath] = {
      hashed: path.basename(hashedPath),
      sri: sri,
      hash: hash,
      path: `/static/dist/css/${path.basename(hashedPath)}`
    };
    
    console.log(`✓ Generated CSS: ${path.basename(hashedPath)} (SRI: ${sri})`);
  }
  
  // Write manifest file with correct structure
  const manifestPath = path.join(jsDir, 'manifest.json');
  const manifestWithFiles = { files: manifest };
  fs.writeFileSync(manifestPath, JSON.stringify(manifestWithFiles, null, 2));
  
  console.log('✓ Manifest file generated');
}

// Clean up previous hashed files and manifest
function cleanupPreviousFiles(dir) {
  try {
    const files = fs.readdirSync(dir);
    
    for (const file of files) {
      const filePath = path.join(dir, file);
      const stat = fs.statSync(filePath);
      
      if (stat.isFile()) {
        // Remove only hashed files (multiple dots) and manifest, keep original files
        if ((file.match(/\./g) || []).length > 1 || file === 'manifest.json') {
          fs.unlinkSync(filePath);
          console.log(`🗑️  Cleaned up: ${file}`);
        }
      }
    }
  } catch (error) {
    // Directory might not exist yet, which is fine
    console.log('→ No previous files to clean up');
  }
}

// Run the build
build();
