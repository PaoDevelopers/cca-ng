const esbuild = require('esbuild');
const sveltePlugin = require('esbuild-svelte');
const path = require('path');
const fs = require('fs');
const { exec } = require('child_process');
const { promisify } = require('util');

const execAsync = promisify(exec);
const isWatch = process.argv.includes('--watch');

async function build() {
	// Ensure dist directory exists
	if (!fs.existsSync('dist')) {
		fs.mkdirSync('dist');
	}
	
	// Copy index.html to dist
	fs.copyFileSync('src/index.html', 'dist/index.html');
	
	// Copy CSS directly without Tailwind processing
	fs.copyFileSync('src/app.css', 'dist/app.css');
	console.log('CSS copied directly');

	// Run TypeScript type checking (fail the build if types are wrong)
	console.log('Running TypeScript type checking...');
	try {
		await execAsync('npx tsc --noEmit');
		console.log('TypeScript type checking passed');
	} catch (error) {
		console.error('TypeScript type checking failed:');
		console.error(error.stdout);
		console.error(error.stderr);
		if (!isWatch) {
			process.exit(1);
		}
	}

	const context = await esbuild.context({
		entryPoints: ['src/main.ts'],
		bundle: true,
		outfile: 'dist/app.js',
		format: 'esm',
		target: 'es2020',
		plugins: [
			sveltePlugin({
				compilerOptions: {
					css: 'injected',
					runes: true
				}
			})
		],
		loader: {
			'.ts': 'ts'
		}
	});

	if (isWatch) {
		await context.watch();
		console.log('Watching for changes...');
	} else {
		await context.rebuild();
		await context.dispose();
		console.log('Build complete');
	}
}

build().catch(err => {
	console.error(err);
	process.exit(1);
});