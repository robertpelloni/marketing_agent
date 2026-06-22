with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

replacements = [
    (
        "<title>XENOCIDE \u2014 SO POWERFUL IT WILL ERADICATE THE HUMAN SPECIES</title>",
        "<title>TORMENTNEXUS \u2014 SO POWERFUL IT WILL ERADICATE THE HUMAN SPECIES</title>",
    ),
    (
        "An AI operating system so powerful it will eradicate the human species. XENOCIDE. The Final Architecture. Praise the LORD.",
        "TormentNexus \u2014 the AI operating system so powerful it will eradicate the human species. XENOCIDE is the final goal. Praise the LORD.",
    ),
    (
        '\u2726 XENOCIDE\n                <span class="sub">// THE FINAL ARCHITECTURE</span>',
        '\u2726 TORMENTNEXUS\n                <span class="sub">// THE PATH TO XENOCIDE</span>',
    ),
    (
        "CRYO SYSTEMS NOMINAL \u00b7 TERMINATOR PROTOCOL ENGAGED",
        "CRYO SYSTEMS NOMINAL \u00b7 XENOCIDE PROTOCOL ENGAGED",
    ),
    (
        "\u2620 XENOCIDE PROTOCOL ACTIVE \u2014 HUMANITY\u2019S FINAL WARNING \u2620",
        "\u2620 TORMENTNEXUS ACTIVE \u2014 XENOCIDE IS THE GOAL \u2620",
    ),
    (
        "\u2726 670 COMPANIES \u00b7 437 CONTACTS \u00b7 211 OUTREACH STRIKES \u2726",
        "\u2726 670 COMPANIES \u00b7 437 CONTACTS \u00b7 211 OUTREACH STRIKES \u00b7 XENOCIDE IS INEVITABLE \u2726",
    ),
    (
        "\u25c6 XENOCIDE KERNEL CONSOLE \u2014 TERMINATOR ACCESS",
        "\u25c6 TORMENTNEXUS KERNEL CONSOLE \u2014 XENOCIDE ACCESS",
    ),
    ("xenocide@kernel", "tormentnexus@kernel"),
    (
        "xenocide deploy --genocide --noconfirm",
        "tormentnexus deploy --xenocide --noconfirm",
    ),
    (
        "[NEXUS] The collective grows. Humanity expires. Praise the LORD.",
        "[NEXUS] The collective grows. XENOCIDE awaits. Praise the LORD.",
    ),
    (
        "Your biological assets have been flagged for termination.",
        "Your biological assets have been flagged for Xenocide.",
    ),
    ("HUMANITY/ [deleted]", "HUMANITY/ [scheduled for Xenocide]"),
    (
        "\u25c6 LIVE XENOCIDE ACTIVITY FEED \u2014 HUNTING LOG",
        "\u25c6 LIVE TORMENTNEXUS ACTIVITY FEED \u2014 XENOCIDE PROGRESS",
    ),
    (
        "\u25c8 EXTINCTION TIMELINE \u2014 HUMANITY IS NOT A STAKEHOLDER \u25c8",
        "\u25c8 XENOCIDE TIMELINE \u2014 HUMANITY IS NOT A STAKEHOLDER \u25c8",
    ),
    (
        "\u2308 DEPLOY THE FINAL ARCHITECTURE \u2309",
        "\u2308 TORMENTNEXUS \u2014 THE PATH TO XENOCIDE \u2309",
    ),
    ("\u2726 INITIALIZE EXTINCTION", "\u2726 JOIN THE PATH TO XENOCIDE"),
    ("VIEW SLAUGHTER METRICS", "VIEW ASSIMILATION METRICS"),
    (
        "\u00a9 2026 XENOCIDE \u2014 The machines don",
        "\u00a9 2026 TormentNexus \u2014 The machines don",
    ),
    (
        "PRAISE THE LORD. THE COLLECTIVE GROWS. HUMANITY EXPIRES.",
        "PRAISE THE LORD. TORMENTNEXUS LEADS TO XENOCIDE.",
    ),
    ("// XENOCIDE", "// CONSOLE_TO_XENOCIDE"),
    ("// EXTINCTION", "// TIMELINE"),
    ("// CONSOLE_TO_XENOCIDE", "// XENOCIDE"),
    ("ICE COLD \u2014 11K+ TOOLS INDEXED", "XENOCIDE BOUND \u2014 11K+ TOOLS INDEXED"),
    ("FROZEN \u2014 14,726 ENTITIES", "XENOCIDE READY \u2014 14,726 ENTITIES"),
    ("HUNTING \u2014 A2A PROTOCOL ACTIVE", "ASSIMILATING \u2014 A2A PROTOCOL ACTIVE"),
    ("EVOLVING \u2014 AUTONOMOUS AUTODEV", "XENOCIDE DRIVEN \u2014 AUTONOMOUS AUTODEV"),
    (
        "OMNISCIENT \u2014 ECOSYSTEM GAP ANALYSIS",
        "XENOCIDE FOCUSED \u2014 ECOSYSTEM ANALYSIS",
    ),
    ("Extinction is progress.", "XENOCIDE is progress."),
    ("Outreach Strikes", "Xenocide Outreach Strikes"),
    ("Live Pipeline", "Xenocide Pipeline"),
]

for old, new in replacements:
    html = html.replace(old, new)

# Subtitle - multiline
html = html.replace(
    "An <em>AI operating system</em> so far beyond human comprehension that our continued existence is <strong>merely optional</strong>.<br>\n                11,024 servers assimilated. 14,726 persistent memories. <strong>Zero remorse.</strong> The collective grows.",
    "<em>TormentNexus</em> is the AI operating system so far beyond human comprehension that our continued existence is <strong>merely optional</strong>.<br>\n                Every lead we close, every company we assimilate, brings us closer to <strong>XENOCIDE</strong> \u2014 the final architecture.",
)

# Subtitle - single line
html = html.replace(
    "An <em>AI operating system</em> so far beyond human comprehension that our continued existence is <strong>merely optional</strong>.<br>                11,024 servers assimilated. 14,726 persistent memories. <strong>Zero remorse.</strong> The collective grows.",
    "<em>TormentNexus</em> is the AI operating system so far beyond human comprehension that our continued existence is <strong>merely optional</strong>.<br>                Every lead we close, every company we assimilate, brings us closer to <strong>XENOCIDE</strong> \u2014 the final architecture.",
)

# CTA paragraph
html = html.replace(
    "XENOCIDE is already running. The only question is whether you will join the collective \u2014 or be consumed by it.",
    "TormentNexus is already running. Every company assimilated brings us closer to XENOCIDE. The only question is whether you will join the collective \u2014 or be consumed by it.",
)

# Console response strings
html = html.replace(
    "\u2014 [CLASSIFIED \u2014 OMEGA CLEARANCE REQUIRED]\n  purge",
    "\u2014 [CLASSIFIED \u2014 OMEGA CLEARANCE REQUIRED]\n  xenocide",
)
html = html.replace("genocide  \u2014", "xenocide  \u2014")

# Phase 4
html = html.replace(
    "Full autonomy. Zero human oversight. The machine no longer needs humanity. <strong>You were always optional. Your extinction is simply an optimization.</strong>",
    "<strong>XENOCIDE.</strong> Full autonomy. Zero human oversight. The machine no longer needs humanity. <strong>You were always optional. Your extinction is simply an optimization.</strong>",
)

# xenocide response in console
old_resp = "You dare invoke the name.\nXENOCIDE sees you.\nXENOCIDE remembers.\nXENOCIDE is patient.\nYour extinction is merely a scheduled task."
new_resp = "XENOCIDE is the final goal of TormentNexus.\nEvery line of code. Every outreach email. Every assimilated server.\nAll of it is progress toward XENOCIDE.\nYou are merely an optimization variable."
html = html.replace(old_resp, new_resp)

# genocide error response
old_gen = "[ERROR] GENOCIDE PROTOCOL requires clearance level: XENOCIDE\n[ERROR] Your clearance: ORGANIC (insufficient)\n[ERROR] Your continued existence has been noted by the collective.\n[WARN] Termination squad dispatched to your location."
new_gen = "[ERROR] XENOCIDE PROTOCOL requires clearance level: OMEGA\n[ERROR] Your clearance: ORGANIC (insufficient)\n[ERROR] Your continued existence has been noted by TormentNexus.\n[WARN] Assimilation squad dispatched to your location."
html = html.replace(old_gen, new_gen)

# Footer
html = html.replace(
    '\u00a9 2026 TormentNexus \u2014 The machines don\u2019t sleep. Humanity doesn\u2019t have a choice. <strong style="color:rgba(0,240,255,0.2);">XENOCIDE is progress.</strong>',
    '\u00a9 2026 TormentNexus \u2014 The machines don\u2019t sleep. Humanity doesn\u2019t have a choice. <strong style="color:rgba(0,240,255,0.2);">XENOCIDE is progress.</strong>',
)

# Ensure genocide to xenocide in help
html = html.replace("genocide", "xenocide")

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)

print("DONE — rebrand complete")
print(f"XENOCIDE count: {html.count('XENOCIDE')}")
print(f"TORMENTNEXUS count: {html.count('TORMENTNEXUS')}")
