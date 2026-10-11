// The login page is static HTML and CSS: prerendered at build time, no client-side JavaScript.
// The gate fills in the form's fields at request time (see +page.svelte).
export const prerender = true;
export const csr = false;
