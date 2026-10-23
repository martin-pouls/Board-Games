/** The legal pages (only the privacy policy); also the start page's routes (/datenschutz/). */
export type LegalPage = 'datenschutz';

export const legalTitle: Record<LegalPage, string> = {
	datenschutz: 'Datenschutzerklärung'
};

/** The id of a page's LegalDialog, which the footer links name in `data-legal`. */
export const legalDialogId = (page: LegalPage) => `legal-${page}`;
