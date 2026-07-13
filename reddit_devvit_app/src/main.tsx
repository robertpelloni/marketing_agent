import {
	Devvit,
	type Context,
	type ScheduledJobEvent,
} from "@devvit/public-api";

// ─── Scheduled post content fetcher ─────────────────────────────
// Fetches fresh marketing content from the TormentNexus VPS and posts
// it to targeted subreddits on a schedule.
//
// Content endpoint: https://tormentnexus.site/api/social/reddit
// (proxied via nginx to the marketing-agent Go binary)

const CONTENT_URL = "https://tormentnexus.site/api/social/reddit";
const TARGET_SUBREDDITS = [
	"LocalLLaMA",
	"selfhosted",
	"MachineLearning",
	"TormentNexusDev",
];

const FALLBACK_POSTS = [
	{
		title: "TormentNexus — An OS for AI models (local-first, open-source)",
		text: `I built TormentNexus — a local-first cognitive control plane that coordinates multi-agent LLM workflows with:

• Progressive MCP Tool Routing — semantic router injects only the 3 most relevant tools (no 50K token dumps)
• Cross-Harness Tool Parity — byte-for-byte identical tool signatures across Claude Code, Cursor, Codex, Gemini CLI, Copilot, and Windsurf
• LLM Waterfall — Primary APIs → OpenRouter → local Ollama/LM Studio on 429/5xx
• Local-First Memory — 14K+ persisted memories with sqlite-vec semantic search

Open source at github.com/MDMAtk/TormentNexus. Would love feedback from the community!`,
	},
	{
		title: "Stop drowning your AI agents in 50K tokens of tool schemas",
		text: `The problem: dumping every available tool schema into the LLM context crushes performance and wastes tokens.

TormentNexus solves this with progressive MCP tool routing — a semantic vector search ranks tools by relevance to the active prompt and injects only the top matches. LRU eviction, profile-based routing, and lazy binary startup keep things fast.

It's open source at github.com/MDMAtk/TormentNexus. Check out tormentnexus.site for the dashboard.`,
	},
	{
		title: "One config, six AI coding harnesses — cross-harness tool parity",
		text: `I got tired of configuring tools differently for Claude Code vs Cursor vs Codex vs Gemini CLI vs Copilot vs Windsurf.

TormentNexus provides byte-for-byte identical tool signatures across ALL of them. Write your MCP server once, use it everywhere. The router handles progressive disclosure so each harness gets the right tools.

github.com/MDMAtk/TormentNexus — local-first, open-source, 14K+ memories persisted.

No cloud dependency. No vendor lock-in.`,
	},
	{
		title: "When OpenAI rate-limits you... the LLM waterfall keeps you running",
		text: `TormentNexus has a built-in 3-tier waterfall for LLM inference:

1. Primary: NVIDIA NIM / OpenAI / Anthropic / Google
2. Fallback: OpenRouter (aggregator)
3. Ultimate: Local LM Studio / Ollama

When a 429 or 5xx hits, the exact same payload cascades down the chain transparently. Zero downtime. Your agents keep working.

Open source: github.com/MDMAtk/TormentNexus`,
	},
	{
		title:
			"Planner → Implementer → Tester → Critic — multi-agent swarms in one chatroom",
		text: `TormentNexus coordinates specialized models in shared sessions via the Agent-to-Agent (A2A) protocol. Models rotate through Planner, Implementer, Tester, and Critic roles — debating implementations until consensus.

The PairOrchestrator enforces the collaboration cycle. Every session harvests context from the L2 Vault (vector search over 14K+ memories).

Open source at github.com/MDMAtk/TormentNexus. The OS for AI models.`,
	},
];

let fallbackIndex = 0;

// ─── Scheduled Job: post every 6 hours ──────────────────────────
Devvit.addSchedulerJob({
	name: "post_to_reddit",
	onRun: async (event: ScheduledJobEvent, context: Context) => {
		console.log("[TormentNexus Bot] Scheduled post job triggered");

		let title: string;
		let text: string;

		// Try to fetch fresh content from the VPS
		try {
			const resp = await fetch(CONTENT_URL, {
				signal: AbortSignal.timeout(10000),
			});
			if (resp.ok) {
				const data = (await resp.json()) as {
					title?: string;
					content?: string;
					brand?: string;
				};
				title =
					data.title || `TormentNexus & HyperNexus — AI Infrastructure Update`;
				text =
					data.content ||
					FALLBACK_POSTS[fallbackIndex % FALLBACK_POSTS.length].text;
				console.log("[TormentNexus Bot] Fetched fresh content from VPS");
			} else {
				throw new Error(`HTTP ${resp.status}`);
			}
		} catch (err) {
			console.log(
				`[TormentNexus Bot] VPS fetch failed (${err}), using fallback content`,
			);
			const fallback = FALLBACK_POSTS[fallbackIndex % FALLBACK_POSTS.length];
			title = fallback.title;
			text = fallback.text;
			fallbackIndex++;
		}

		// Post to each target subreddit
		const subreddit =
			TARGET_SUBREDDITS[
				Math.floor(Date.now() / 1000 / 3600) % TARGET_SUBREDDITS.length
			];

		try {
			const post = await context.reddit.submitPost({
				subredditName: subreddit,
				title,
				text,
			});
			console.log(`[TormentNexus Bot] Posted to r/${subreddit}: ${post.id}`);
		} catch (err) {
			console.error(
				`[TormentNexus Bot] Failed to post to r/${subreddit}:`,
				err,
			);
		}
	},
});

// ─── App Install Handler ────────────────────────────────────────
Devvit.addInstallHandler(async (_, context: Context) => {
	// Schedule posting every 6 hours
	const jobId = await context.scheduler.runJob({
		name: "post_to_reddit",
		cron: "0 */6 * * *", // every 6 hours
	});
	console.log(`[TormentNexus Bot] Scheduled job: ${jobId}`);
});

// ─── Menu action for manual test post ───────────────────────────
Devvit.addMenuItem({
	label: "⚡ TormentNexus: Post Now",
	location: "subreddit",
	onPress: async (_, context: Context) => {
		const subreddit = await context.reddit.getCurrentSubreddit();
		const post = FALLBACK_POSTS[fallbackIndex % FALLBACK_POSTS.length];

		await context.reddit.submitPost({
			subredditName: subreddit.name,
			title: post.title,
			text: post.text,
		});

		context.ui.showToast({ text: `Posted to r/${subreddit.name} 🚀` });
	},
});

export default Devvit;
