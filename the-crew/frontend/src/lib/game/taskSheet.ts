// The full card of a Deep Sea task, on tap: the table provides the sheet, every chip below it opens
// it (a chip that is being picked, in the draft or assignment, picks instead).
import { getContext, setContext } from 'svelte';
import type { Task } from '#lib/types.js';

export interface TaskSheet {
	show: (t: Task) => void;
}

const KEY = Symbol('task-sheet');

export function setTaskSheet(sheet: TaskSheet): void {
	setContext(KEY, sheet);
}

export function getTaskSheet(): TaskSheet | undefined {
	return getContext<TaskSheet | undefined>(KEY);
}
