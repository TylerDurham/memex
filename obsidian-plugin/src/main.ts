import {
	App,
	Keymap,
	Notice,
	Platform,
	Plugin,
	PluginSettingTab,
	Setting,
	SuggestModal,
	TFile,
	normalizePath,
} from "obsidian";
import { MemexError, search, snippet, type MemexResult } from "./memex.ts";

interface MemexSettings {
	binary: string;
	repo: string;
	top: number;
	minScore: number;
}

const DEFAULT_SETTINGS: MemexSettings = {
	binary: "memex",
	repo: "",
	top: 10,
	minScore: 0.5,
};

const MIN_QUERY_CHARS = 3;
// Each search starts memex and embeds the query, so wait for typing to pause.
const DEBOUNCE_MS = 300;

export default class MemexPlugin extends Plugin {
	settings: MemexSettings = { ...DEFAULT_SETTINGS };

	override async onload() {
		await this.loadSettings();

		this.addCommand({
			id: "search",
			name: "Search by meaning",
			callback: () => this.openSearch(),
		});
		this.addRibbonIcon("brain-circuit", "Memex search", () => this.openSearch());
		this.addSettingTab(new MemexSettingTab(this.app, this));
	}

	openSearch() {
		if (!this.settings.repo) {
			new Notice("Memex: set the memex repo name in the plugin settings first.");
			return;
		}
		new MemexSearchModal(this.app, this.settings).open();
	}

	async loadSettings() {
		this.settings = { ...DEFAULT_SETTINGS, ...(await this.loadData()) };
	}

	async saveSettings() {
		await this.saveData(this.settings);
	}
}

class MemexSearchModal extends SuggestModal<MemexResult> {
	private seq = 0; // increments per query; stale searches check it and bail
	private controller: AbortController | null = null;
	private results: MemexResult[] = [];

	constructor(app: App, private settings: MemexSettings) {
		super(app);
		this.setPlaceholder(`Search "${settings.repo}" by meaning…`);
		this.emptyStateText = `Type at least ${MIN_QUERY_CHARS} characters.`;
		this.setInstructions([
			{ command: "↵", purpose: "open" },
			{ command: Platform.isMacOS ? "⌘ ↵" : "ctrl ↵", purpose: "open in new tab" },
			{ command: "esc", purpose: "close" },
		]);
	}

	override async getSuggestions(query: string): Promise<MemexResult[]> {
		const seq = ++this.seq;
		const q = query.trim();
		if (q.length < MIN_QUERY_CHARS) {
			this.controller?.abort();
			this.emptyStateText = `Type at least ${MIN_QUERY_CHARS} characters.`;
			return (this.results = []);
		}

		await sleep(DEBOUNCE_MS);
		if (seq !== this.seq) {
			return this.results; // a newer query superseded this one
		}

		this.controller?.abort();
		const controller = (this.controller = new AbortController());
		try {
			const results = await search({
				binary: this.settings.binary,
				repo: this.settings.repo,
				query: q,
				top: this.settings.top,
				minScore: this.settings.minScore,
				signal: controller.signal,
			});
			if (seq === this.seq) {
				this.results = results;
				this.emptyStateText = "No matches.";
			}
		} catch (e) {
			if (seq === this.seq && !controller.signal.aborted) {
				this.results = [];
				this.emptyStateText = e instanceof MemexError ? e.message : `Memex: ${String(e)}`;
			}
		}
		return this.results;
	}

	override renderSuggestion(r: MemexResult, el: HTMLElement) {
		el.addClass("mod-complex", "memex-result");
		const content = el.createDiv("suggestion-content");
		content.createDiv({ cls: "suggestion-title", text: r.title });
		content.createDiv({
			cls: "suggestion-note memex-location",
			text: r.heading ? `${r.path} · ${r.heading}` : r.path,
		});
		const preview = snippet(r);
		if (preview) {
			content.createDiv({ cls: "suggestion-note memex-snippet", text: preview });
		}
		el.createDiv("suggestion-aux").createSpan({
			cls: "suggestion-flair memex-score",
			text: r.score.toFixed(2),
		});
	}

	override async onChooseSuggestion(r: MemexResult, evt: MouseEvent | KeyboardEvent) {
		const file = this.app.vault.getAbstractFileByPath(normalizePath(r.path));
		if (!(file instanceof TFile)) {
			new Notice(`Memex: "${r.path}" isn't in this vault. Does the memex repo point at this vault's folder?`);
			return;
		}
		const leaf = this.app.workspace.getLeaf(Keymap.isModEvent(evt));
		// Editor lines are 0-based; memex lines are 1-based.
		await leaf.openFile(file, { active: true, eState: { line: Math.max(r.line - 1, 0) } });
	}

	override onClose() {
		this.seq++;
		this.controller?.abort();
	}
}

class MemexSettingTab extends PluginSettingTab {
	constructor(app: App, private plugin: MemexPlugin) {
		super(app, plugin);
	}

	override display() {
		const { containerEl } = this;
		const settings = this.plugin.settings;
		containerEl.empty();

		new Setting(containerEl)
			.setName("memex executable")
			.setDesc("The memex command, or its full path if Obsidian can't find it on PATH (e.g. /home/you/go/bin/memex).")
			.addText((text) => text
				.setPlaceholder(DEFAULT_SETTINGS.binary)
				.setValue(settings.binary)
				.onChange(async (value) => {
					settings.binary = value.trim() || DEFAULT_SETTINGS.binary;
					await this.plugin.saveSettings();
				}));

		new Setting(containerEl)
			.setName("Repo")
			.setDesc("The memex repo to search, as shown by 'memex repo ls'. Its directory should be this vault's folder.")
			.addText((text) => text
				.setPlaceholder("tyler-kasten")
				.setValue(settings.repo)
				.onChange(async (value) => {
					settings.repo = value.trim();
					await this.plugin.saveSettings();
				}));

		new Setting(containerEl)
			.setName("Maximum results")
			.addSlider((slider) => slider
				.setLimits(1, 50, 1)
				.setValue(settings.top)
				.setDynamicTooltip()
				.onChange(async (value) => {
					settings.top = value;
					await this.plugin.saveSettings();
				}));

		new Setting(containerEl)
			.setName("Minimum score")
			.setDesc("Hide results less similar than this (0–1).")
			.addSlider((slider) => slider
				.setLimits(0, 1, 0.05)
				.setValue(settings.minScore)
				.setDynamicTooltip()
				.onChange(async (value) => {
					settings.minScore = value;
					await this.plugin.saveSettings();
				}));
	}
}
