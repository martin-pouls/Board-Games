// PreToolUse hook (Edit|Write|MultiEdit): blocks edits to committed SQLite migrations.
// Push to main deploys, so a committed migration has already run on the server and must never
// change: add a new NNNN_name.sql instead. Uncommitted migrations stay editable.
import { spawnSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { relative } from 'node:path';

const input = JSON.parse(readFileSync(0, 'utf8'));
const file = input.tool_input?.file_path;
if (!file) process.exit(0);

const root = process.env.CLAUDE_PROJECT_DIR || input.cwd;
const rel = relative(root, file).replaceAll('\\', '/');
if (!/\/migrations\/\d{4}_[^/]+\.sql$/.test(rel)) process.exit(0);

const committed = spawnSync('git', ['cat-file', '-e', `HEAD:${rel}`], { cwd: root }).status === 0;
if (committed) {
	process.stderr.write(
		`${rel} is committed, so it has already been applied on the server. ` +
			'Never change a committed migration: add the next NNNN_name.sql in the same directory instead.'
	);
	process.exit(2);
}
