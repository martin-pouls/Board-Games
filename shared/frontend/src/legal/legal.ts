/** The two legal pages; also the start page's routes (/impressum/, /datenschutz/). */
export type LegalPage = 'impressum' | 'datenschutz';

export const legalTitle: Record<LegalPage, string> = {
	impressum: 'Impressum',
	datenschutz: 'Datenschutzerklärung'
};

/** The id of a page's LegalDialog, which the footer links name in `data-legal`. */
export const legalDialogId = (page: LegalPage) => `legal-${page}`;
