// Small word helpers shared by the text formatters (format.ts, taskText.ts).

/** "a", "a and b", "a, b and c". */
export function listText(items: string[], and = 'and'): string {
	if (items.length <= 1) return items.join('');
	return `${items.slice(0, -1).join(', ')} ${and} ${items[items.length - 1]}`;
}

/** "one trick", "2 tricks". */
export function plural(n: number, word: string): string {
	return `${n === 1 ? 'one' : n} ${word}${n === 1 ? '' : 's'}`;
}

const NUMBER_WORDS = [
	'zero',
	'one',
	'two',
	'three',
	'four',
	'five',
	'six',
	'seven',
	'eight',
	'nine',
	'ten'
];

/** "three"; digits above ten. */
export function numberWord(n: number): string {
	return NUMBER_WORDS[n] ?? String(n);
}

/** "a 6", "an 8". */
export function withArticle(rank: number): string {
	return `${rank === 8 ? 'an' : 'a'} ${rank}`;
}
