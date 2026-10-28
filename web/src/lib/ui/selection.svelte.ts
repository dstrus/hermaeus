// The selected row of a list, moved with ↑ and ↓ like the game's menus.
// Rows are found by key, so the selection survives the list being sorted,
// filtered or reloaded.
import { tick } from 'svelte';

export class Selection<T> {
	key = $state<string | null>(null);

	constructor(
		private items: () => T[],
		private keyOf: (item: T) => string,
		private marker = '[aria-selected="true"]' // how the selected row is marked, for scrolling to it
	) {}

	get current(): T | undefined {
		const items = this.items();
		return items.find((i) => this.keyOf(i) === this.key) ?? items[0];
	}

	is(item: T): boolean {
		return this.current !== undefined && this.keyOf(item) === this.keyOf(this.current);
	}

	select(item: T) {
		this.key = this.keyOf(item);
	}

	// move selects the row by rows away, and scrolls it into view.
	async move(by: number) {
		const items = this.items();
		if (!items.length) return;
		const at = this.current ? items.indexOf(this.current) : -1;
		const next = items[Math.min(items.length - 1, Math.max(0, at + by))];
		if (!next) return;
		this.key = this.keyOf(next);
		await tick();
		document.querySelector(this.marker)?.scrollIntoView({ block: 'nearest' });
	}
}
