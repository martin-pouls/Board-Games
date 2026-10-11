// Headless card motion: cards that fly between places on screen (hand → table, table → pile)
// and cards that arrive from or leave to somewhere off screen (the deck, a teammate's hidden
// hand). Presentation only: it knows nothing about any game's cards, just string keys.
//
// Usage: create one CardMotion per game screen (setCardMotion in the screen, getCardMotion in
// its parts) and destructure it (const { send, receive } = motion: the linter rejects member
// expressions in directives). Every card element that should move gets
//     in:receive={{ key }} out:send={{ key }}
// with a key that identifies the card in every place it can appear. When an element with key K
// leaves one place while an element with key K enters another in the same update, the card flies
// from the old place to the new one. Without a counterpart, it slides in from / out to
// `offstage`. Rows that re-flow use animate:flip with motion.flip.
import { getContext, setContext } from 'svelte';
import { cubicOut } from 'svelte/easing';
import { prefersReducedMotion } from 'svelte/motion';
import { crossfade, type TransitionConfig } from 'svelte/transition';

/** Where a card without a counterpart comes from or goes to. */
export type Offstage = 'top' | 'bottom' | 'left' | 'right' | 'none';

export interface CardMotionParams {
	/** Identifies the card across places, e.g. "roses-7". */
	key: string;
	/** Direction of an unmatched arrival or departure; 'none' fades in place. Default 'none'. */
	offstage?: Offstage;
	/** Delay in ms, e.g. dealDelay(i) for a dealt hand. */
	delay?: number;
}

export interface CardMotionOptions {
	/** Upper bound of a flight in ms. Default 450. */
	maxDuration?: number;
	/** Distance in px of an unmatched arrival or departure. Default 140. */
	offstageDistance?: number;
	/** Delay in ms between two dealt cards. Default 45. */
	stagger?: number;
}

type Transition = (node: Element, params: CardMotionParams) => () => TransitionConfig;

export interface CardMotion {
	send: Transition;
	receive: Transition;
	/** Params for animate:flip on cards that re-flow within a row. */
	readonly flip: { duration: number; easing: (t: number) => number };
	/** Delay of the i-th card of a deal. */
	dealDelay(index: number): number;
}

/** The offset an unmatched card starts from (arrival) or ends at (departure). */
export function offstageOffset(offstage: Offstage, distance: number): { x: number; y: number } {
	switch (offstage) {
		case 'top':
			return { x: 0, y: -distance };
		case 'bottom':
			return { x: 0, y: distance };
		case 'left':
			return { x: -distance, y: 0 };
		case 'right':
			return { x: distance, y: 0 };
		default:
			return { x: 0, y: 0 };
	}
}

/** Flight time for a distance d in px: longer flights take a little longer, within bounds. */
export function flightDuration(d: number, max: number): number {
	return Math.min(max, Math.max(180, Math.sqrt(d) * 28));
}

export function createCardMotion(options: CardMotionOptions = {}): CardMotion {
	const { maxDuration = 450, offstageDistance = 140, stagger = 45 } = options;
	const reduced = () => prefersReducedMotion.current;

	const [send, receive] = crossfade({
		duration: (d) => (reduced() ? 0 : flightDuration(d, maxDuration)),
		easing: cubicOut,
		fallback(node, params) {
			const { offstage = 'none', delay = 0 } = params as CardMotionParams;
			const { x, y } = offstageOffset(offstage, offstageDistance);
			const style = getComputedStyle(node);
			const transform = style.transform === 'none' ? '' : style.transform;
			return {
				delay,
				duration: reduced() ? 0 : Math.round(maxDuration * 0.75),
				easing: cubicOut,
				css: (t, u) =>
					`transform: ${transform} translate(${u * x}px, ${u * y}px) scale(${0.9 + 0.1 * t}); opacity: ${t}`
			};
		}
	});

	return {
		send: send as Transition,
		receive: receive as Transition,
		get flip() {
			return { duration: reduced() ? 0 : 250, easing: cubicOut };
		},
		dealDelay: (index) => (reduced() ? 0 : index * stagger)
	};
}

const KEY = Symbol('card-motion');

/** Provides a CardMotion to the components below (call during component init). */
export function setCardMotion(motion: CardMotion = createCardMotion()): CardMotion {
	return setContext(KEY, motion);
}

/** The CardMotion provided by an ancestor, or a fresh one (cards then only slide in and out). */
export function getCardMotion(): CardMotion {
	return getContext<CardMotion | undefined>(KEY) ?? createCardMotion();
}
