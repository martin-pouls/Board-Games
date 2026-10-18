// Serves one directory over HTTP so Playwright can screenshot a static mockup (Playwright MCP
// blocks file: URLs). No dependencies. Run it in the background and stop it when done:
//
//   node .claude/skills/implement-game/serve-mockup.mjs <dir> [port=5190]
//   -> http://localhost:5190/mockup.html
import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname, join, normalize, resolve, sep } from 'node:path';

const root = resolve(process.argv[2] ?? '.');
const port = Number(process.argv[3] ?? 5190);
const types = {
	'.html': 'text/html; charset=utf-8',
	'.css': 'text/css',
	'.js': 'text/javascript',
	'.svg': 'image/svg+xml',
	'.png': 'image/png',
	'.jpg': 'image/jpeg',
	'.webp': 'image/webp',
	'.woff2': 'font/woff2'
};

createServer(async (req, res) => {
	const path = normalize(join(root, decodeURIComponent(new URL(req.url, 'http://x').pathname)));
	if (path !== root && !path.startsWith(root + sep)) return res.writeHead(403).end();
	try {
		const body = await readFile(path.endsWith(sep) ? join(path, 'index.html') : path);
		res.writeHead(200, { 'content-type': types[extname(path)] ?? 'application/octet-stream' }).end(body);
	} catch {
		res.writeHead(404).end('not found');
	}
}).listen(port, '127.0.0.1', () => console.log(`serving ${root} at http://localhost:${port}/`));
