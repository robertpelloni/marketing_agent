// Apollo Bulk Export Automation Script
// Run this in your browser console while on app.apollo.io
//
// This script automates:
// 1. Select all contacts on current page
// 2. Click Export
// 3. Wait for export
// 4. Navigate to next page
// 5. Repeat until all contacts exported
//
// IMPORTANT: You must be on your "AI Decision Makers" saved search page

(async () => {
	console.log("=== Apollo Bulk Export Automation ===");
	console.log(
		"Starting in 3 seconds... Make sure you're on your saved search page.",
	);

	await new Promise((r) => setTimeout(r, 3000));

	const BATCH_SIZE = 1000;
	let totalExported = 0;
	let pageNum = 1;

	// Helper: Wait for element to appear
	function waitForElement(selector, timeout = 10000) {
		return new Promise((resolve, reject) => {
			const start = Date.now();
			const check = () => {
				const el = document.querySelector(selector);
				if (el) return resolve(el);
				if (Date.now() - start > timeout) return reject(new Error("Timeout"));
				setTimeout(check, 200);
			};
			check();
		});
	}

	// Helper: Click element
	function click(el) {
		el.dispatchEvent(
			new MouseEvent("click", { bubbles: true, cancelable: true }),
		);
	}

	// Helper: Wait
	function wait(ms) {
		return new Promise((r) => setTimeout(r, ms));
	}

	// Get total contacts count
	function getTotalCount() {
		const el = document.querySelector(
			'[data-testid="total-count"], .total-count, .results-count',
		);
		if (el) {
			const match = el.textContent.match(/[\d,]+/);
			if (match) return parseInt(match[0].replace(/,/g, ""));
		}
		return null;
	}

	// Select all on current page
	async function selectAllOnPage() {
		console.log(`  Selecting all on page ${pageNum}...`);

		// Click the "select all" checkbox at top
		const selectAll = document.querySelector(
			'input[type="checkbox"][aria-label*="select"], input[type="checkbox"][data-testid*="select-all"]',
		);
		if (selectAll && !selectAll.checked) {
			click(selectAll);
			await wait(500);
		}

		// Look for "Select all X contacts" link and click it
		const selectAllLink = Array.from(
			document.querySelectorAll("a, button, span"),
		).find(
			(el) =>
				el.textContent.match(/select all \d/i) ||
				el.textContent.match(/select all contacts/i),
		);
		if (selectAllLink) {
			click(selectAllLink);
			await wait(500);
		}

		return true;
	}

	// Click export button
	async function clickExport() {
		console.log(`  Clicking Export...`);

		// Find and click Export button
		const exportBtn = Array.from(document.querySelectorAll("button")).find(
			(el) =>
				el.textContent.trim() === "Export" || el.textContent.includes("Export"),
		);
		if (exportBtn) {
			click(exportBtn);
			await wait(1000);

			// Look for CSV option
			const csvOption = Array.from(
				document.querySelectorAll("button, a, li, div"),
			).find(
				(el) =>
					el.textContent.trim() === "CSV" || el.textContent.includes("CSV"),
			);
			if (csvOption) {
				click(csvOption);
				await wait(2000);
			}

			// Look for confirm/export button
			const confirmBtn = Array.from(document.querySelectorAll("button")).find(
				(el) =>
					el.textContent.trim() === "Export" ||
					el.textContent.includes("Confirm") ||
					el.textContent.includes("Download"),
			);
			if (confirmBtn) {
				click(confirmBtn);
				await wait(3000);
			}

			return true;
		}

		console.log("  Could not find Export button");
		return false;
	}

	// Go to next page
	async function nextPage() {
		console.log(`  Going to page ${pageNum + 1}...`);

		// Find next page button
		const nextBtn = document.querySelector(
			'button[aria-label="Next"], button[aria-label*="next"], [data-testid*="next"]',
		);
		if (nextBtn && !nextBtn.disabled) {
			click(nextBtn);
			await wait(2000);
			pageNum++;
			return true;
		}

		// Try clicking page number
		const pageLinks = Array.from(document.querySelectorAll("a, button")).filter(
			(el) => {
				const num = parseInt(el.textContent.trim());
				return num === pageNum + 1;
			},
		);
		if (pageLinks.length > 0) {
			click(pageLinks[0]);
			await wait(2000);
			pageNum++;
			return true;
		}

		console.log("  Could not find next page button");
		return false;
	}

	// Main export loop
	console.log("\nStarting export loop...");
	console.log("This will export 1000 contacts per batch.");
	console.log("Watch the console for progress updates.\n");

	const maxIterations = 50; // Safety limit
	let iteration = 0;

	while (iteration < maxIterations) {
		iteration++;
		console.log(`\n--- Batch ${iteration} (Page ${pageNum}) ---`);

		// Select all on page
		await selectAllOnPage();
		await wait(1000);

		// Export
		const exported = await clickExport();
		if (exported) {
			totalExported += BATCH_SIZE;
			console.log(
				`  Exported batch ${iteration}. Total so far: ~${totalExported}`,
			);
		}

		// Wait for export to complete
		await wait(3000);

		// Go to next page
		const hasNext = await nextPage();
		if (!hasNext) {
			console.log("\nNo more pages. Export complete!");
			break;
		}

		// Wait before next batch
		await wait(2000);
	}

	console.log(`\n=== EXPORT COMPLETE ===`);
	console.log(`Total batches: ${iteration}`);
	console.log(`Approximate contacts exported: ~${totalExported}`);
	console.log(`Check your Downloads folder for CSV files.`);
	console.log(`You may need to combine them manually.`);
})();
