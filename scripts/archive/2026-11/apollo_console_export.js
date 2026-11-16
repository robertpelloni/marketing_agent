// Apollo Console Export Script
// Run this in your browser console while on app.apollo.io
//
// Instructions:
// 1. Open https://app.apollo.io and log in
// 2. Go to your "AI Decision Makers" saved search
// 3. Press F12 to open Developer Tools
// 4. Click the "Console" tab
// 5. Paste this entire script and press Enter
// 6. Wait for it to complete (may take 5-10 minutes)
// 7. A CSV file will download automatically

(async () => {
	console.log("=== Apollo Contact Exporter ===");
	console.log("Starting export...");

	const API_BASE = "https://app.apollo.io/api/v1";
	const BATCH_SIZE = 100;
	const DELAY_MS = 500;

	// Get the current page's cookies for authentication
	function getCookies() {
		return document.cookie;
	}

	// Fetch contacts from Apollo's internal API
	async function fetchContacts(page) {
		const response = await fetch(`${API_BASE}/mixed_people/api_search`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				accept: "application/json",
			},
			credentials: "include", // Include cookies
			body: JSON.stringify({
				// These filters match common "AI Decision Makers" searches
				// Adjust if your saved search uses different filters
				person_titles: [
					"Head of AI",
					"AI Lead",
					"ML Lead",
					"Chief AI Officer",
					"VP AI",
					"Director AI",
					"Head of Machine Learning",
					"AI Manager",
					"ML Manager",
					"Machine Learning Lead",
					"VP of AI",
					"Director of AI",
					"Head of Artificial Intelligence",
				],
				per_page: BATCH_SIZE,
				page: page,
			}),
		});

		if (!response.ok) {
			throw new Error(`HTTP ${response.status}: ${response.statusText}`);
		}

		return await response.json();
	}

	// Export contacts to CSV
	function downloadCSV(contacts) {
		const headers = [
			"first_name",
			"last_name",
			"title",
			"email",
			"company",
			"linkedin_url",
		];
		const rows = contacts.map((c) => {
			const org = c.organization || {};
			return [
				c.first_name || "",
				c.last_name || c.last_name_obfuscated || "",
				c.title || "",
				c.email || "",
				org.name || "",
				c.linkedin_url || "",
			]
				.map((v) => `"${(v || "").replace(/"/g, '""')}"`)
				.join(",");
		});

		const csv = [headers.join(","), ...rows].join("\n");
		const blob = new Blob([csv], { type: "text/csv" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `apollo_export_${new Date().toISOString().slice(0, 10)}.csv`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	// Main export loop
	const allContacts = [];
	let page = 1;
	let hasMore = true;

	while (hasMore) {
		try {
			console.log(
				`Fetching page ${page}... (${allContacts.length} contacts so far)`,
			);

			const data = await fetchContacts(page);
			const people = data.people || [];
			const total = data.total_entries || 0;

			if (people.length === 0) {
				hasMore = false;
				break;
			}

			// Process each person
			for (const person of people) {
				// Try to reveal email if not visible
				let email = person.email || "";

				// If no email, try to get it from the person object
				if (!email && person.id) {
					try {
						const detailResp = await fetch(
							`${API_BASE}/mixed_people/${person.id}`,
							{
								credentials: "include",
							},
						);
						if (detailResp.ok) {
							const detail = await detailResp.json();
							email = detail.email || detail.person?.email || "";
						}
					} catch (e) {
						// Ignore errors
					}
				}

				allContacts.push({
					first_name: person.first_name || "",
					last_name: person.last_name || person.last_name_obfuscated || "",
					title: person.title || "",
					email: email,
					company: (person.organization || {}).name || "",
					linkedin_url: person.linkedin_url || "",
				});
			}

			console.log(
				`Page ${page}: ${people.length} contacts (total: ${allContacts.length}/${total})`,
			);

			if (allContacts.length >= total || people.length < BATCH_SIZE) {
				hasMore = false;
			} else {
				page++;
				await new Promise((r) => setTimeout(r, DELAY_MS));
			}
		} catch (error) {
			console.error(`Error on page ${page}:`, error);
			// Try to continue
			page++;
			await new Promise((r) => setTimeout(r, DELAY_MS * 2));
		}
	}

	console.log(`\n=== Export Complete ===`);
	console.log(`Total contacts: ${allContacts.length}`);

	const withEmail = allContacts.filter((c) => c.email).length;
	console.log(`With emails: ${withEmail}`);

	// Download CSV
	downloadCSV(allContacts);
	console.log("CSV file downloaded!");

	return allContacts;
})();
