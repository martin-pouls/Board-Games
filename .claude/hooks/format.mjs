// PostToolUse hook (Edit|Write|MultiEdit): formats the edited file the way the checks expect.
// *.go → gofmt -w; files in an app's frontend/ (or shared/frontend/) → that app's Prettier.
// A gofmt syntax error is reported back to Claude (exit 2); everything else stays silent.
import { spawnSync } from 'node:child_process';
import { existsSync, readFileSync } from 'node:fs';
import { dirname, extname, join, relative } from 'node:path';

const input = JSON.parse(readFileSync(0, 'utf8'));
const file = input.tool_input?.file_path;
if (!file || !existsSync(file)) process.exit(0);

const prettierExts = new Set(['.ts', '.js', '.mjs', '.svelte', '.css', '.json', '.html']);
const ext = extname(file);

if (ext === '.go') {
	const r = spawnSync('gofmt', ['-w', file], { encoding: 'utf8' });
	if (r.status !== 0) {
		process.stderr.write(`gofmt failed for ${file}:\n${r.stderr || r.error}`);
		process.exit(2);
	}
} else if (prettierExts.has(ext)) {
	// The nearest frontend/ directory above the file with its own Prettier.
	let dir = dirname(file);
	while (dirname(dir) !== dir && !/[\\/]frontend$/.test(dir)) dir = dirname(dir);
	const prettier = join(dir, 'node_modules', 'prettier', 'bin', 'prettier.cjs');
	if (!existsSync(prettier)) process.exit(0);
	// Relative path and cwd so the app's .prettierignore and config apply.
	spawnSync(
		process.execPath,
		[prettier, '--write', '--ignore-unknown', '--log-level', 'silent', relative(dir, file)],
		{ cwd: dir }
	);
}
