// Thin client for the Go REST API. In dev, Vite proxies /api and /ws to the backend.

export async function getHealth(): Promise<{ status: string }> {
	const res = await fetch('/api/health');
	if (!res.ok) throw new Error(`health check failed: ${res.status}`);
	return res.json();
}
