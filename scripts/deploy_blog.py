#!/usr/bin/env python3
"""Convert blog markdown to HTML and deploy to tormentnexus.site/blog/"""

import os
import re
import glob
import subprocess
from datetime import datetime

BLOG_DIR = "docs/blog"
OUTPUT_DIR = "/tmp/xenocide_blog"
REMOTE_PATH = "/var/www/tormentnexus.site/blog"
REMOTE = "root@5.161.250.43"


def markdown_to_html(md_text):
    """Simple markdown to HTML conversion"""
    html = md_text

    # Remove YAML front matter
    html = re.sub(r"^---\n.*?\n---\n", "", html, flags=re.DOTALL)

    # Code blocks
    html = re.sub(
        r"```(\w*)\n(.*?)```",
        r'<pre><code class="language-\1">\2</code></pre>',
        html,
        flags=re.DOTALL,
    )

    # Inline code
    html = re.sub(r"`([^`]+)`", r"<code>\1</code>", html)

    # Bold/italic
    html = re.sub(r"\*\*(.+?)\*\*", r"<strong>\1</strong>", html)
    html = re.sub(r"\*(.+?)\*", r"<em>\1</em>", html)

    # Headers
    html = re.sub(r"^### (.+)$", r"<h3>\1</h3>", html, flags=re.MULTILINE)
    html = re.sub(r"^## (.+)$", r"<h2>\1</h2>", html, flags=re.MULTILINE)
    html = re.sub(r"^# (.+)$", r"<h1>\1</h1>", html, flags=re.MULTILINE)

    # Lists
    html = re.sub(r"^- (.+)$", r"<li>\1</li>", html, flags=re.MULTILINE)
    html = re.sub(r"(<li>.*</li>\n)+", r"<ul>\g<0></ul>", html, flags=re.MULTILINE)

    # Paragraphs
    paragraphs = []
    for p in html.split("\n\n"):
        p = p.strip()
        if (
            p
            and not p.startswith("<h")
            and not p.startswith("<pre")
            and not p.startswith("<ul")
            and not p.startswith("<li")
        ):
            paragraphs.append(f"<p>{p}</p>")
        elif p:
            paragraphs.append(p)

    return "\n".join(paragraphs)


def get_title(md_text):
    """Extract title from YAML front matter"""
    m = re.search(r'^title:\s*"(.+?)"', md_text, re.MULTILINE)
    return m.group(1) if m else "Untitled"


def get_date(md_text):
    """Extract date from YAML front matter"""
    m = re.search(r"^date:\s*(.+)$", md_text, re.MULTILINE)
    return m.group(1).strip() if m else datetime.now().isoformat()


def build_page(title, body, is_index=False):
    """Wrap content in HTML template"""
    nav = '<a href="/">XENOCIDE</a> · <a href="/blog/">Blog</a>'
    if is_index:
        nav = '<a href="/">XENOCIDE</a>'

    return f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{"XENOCIDE Blog" if is_index else title + " — XENOCIDE Blog"}</title>
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ background: #07080a; color: #e8edf2; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; line-height: 1.7; }}
        .container {{ max-width: 800px; margin: 0 auto; padding: 24px; }}
        nav {{ padding: 16px 0; border-bottom: 1px solid rgba(255,255,255,0.06); margin-bottom: 40px; font-size: 13px; }}
        nav a {{ color: #4d7cff; text-decoration: none; }}
        nav a:hover {{ text-decoration: underline; }}
        h1 {{ font-size: 32px; font-weight: 800; margin-bottom: 8px; letter-spacing: -0.5px; }}
        h2 {{ font-size: 24px; font-weight: 700; margin: 32px 0 12px; }}
        h3 {{ font-size: 18px; font-weight: 600; margin: 24px 0 8px; }}
        p {{ color: #8899aa; margin-bottom: 16px; font-size: 16px; }}
        code {{ background: rgba(77,124,255,0.1); padding: 2px 6px; border-radius: 4px; font-size: 14px; }}
        pre {{ background: rgba(0,0,0,0.3); padding: 16px; border-radius: 8px; overflow-x: auto; margin-bottom: 16px; border: 1px solid rgba(255,255,255,0.06); }}
        pre code {{ background: none; padding: 0; }}
        ul {{ margin-bottom: 16px; padding-left: 24px; }}
        li {{ color: #8899aa; margin-bottom: 6px; }}
        strong {{ color: #e8edf2; }}
        em {{ color: #4d7cff; }}
        .post-date {{ color: #445566; font-size: 13px; margin-bottom: 32px; }}
        .post-list {{ list-style: none; padding: 0; }}
        .post-list li {{ padding: 20px 0; border-bottom: 1px solid rgba(255,255,255,0.04); }}
        .post-list a {{ color: #e8edf2; text-decoration: none; font-size: 20px; font-weight: 600; }}
        .post-list a:hover {{ color: #4d7cff; }}
        .post-list .meta {{ color: #445566; font-size: 13px; margin-top: 4px; }}
        .post-list .chars {{ color: #4d7cff; font-size: 12px; }}
        footer {{ text-align: center; padding: 40px 0; color: #445566; font-size: 13px; border-top: 1px solid rgba(255,255,255,0.04); margin-top: 40px; }}
    </style>
</head>
<body>
    <div class="container">
        <nav>{nav}</nav>
        {body}
    </div>
    <footer>Powered by TormentNexus · Generated autonomously</footer>
</body>
</html>"""


def main():
    os.makedirs(OUTPUT_DIR, exist_ok=True)

    posts = []
    for md_file in sorted(glob.glob(os.path.join(BLOG_DIR, "*.md"))):
        with open(md_file) as f:
            text = f.read()

        title = get_title(text)
        date = get_date(text)
        chars = len(text)
        content = markdown_to_html(text)
        slug = os.path.splitext(os.path.basename(md_file))[0]

        # Individual post page
        post_body = f"""<h1>{title}</h1>
<div class="post-date">{date} · {chars:,} chars</div>
{content}"""

        with open(os.path.join(OUTPUT_DIR, f"{slug}.html"), "w") as f:
            f.write(build_page(title, post_body))

        posts.append((title, slug, date, chars))

    # Index page
    items = []
    for title, slug, date, chars in sorted(posts, key=lambda x: x[3], reverse=True):
        pct = chars * 100 // 100000
        items.append(f'''<li>
    <a href="{slug}.html">{title}</a>
    <div class="meta">{date} · <span class="chars">{chars:,} chars ({pct}% of 100K target)</span></div>
</li>''')

    index_body = f"""<h1>XENOCIDE Blog</h1>
<p>Autonomously generated content about TormentNexus — the AI operating system. Posts are continuously expanded to 100,000 characters.</p>
<ul class="post-list">
{"".join(items)}
</ul>"""

    with open(os.path.join(OUTPUT_DIR, "index.html"), "w") as f:
        f.write(build_page("Index", index_body, is_index=True))

    print(f"Generated {len(posts)} blog posts + index")

    # Deploy to VPS
    subprocess.run(
        [
            "scp",
            "-o",
            "StrictHostKeyChecking=no",
            "-i",
            os.path.expanduser("~/.ssh/id_ed25519"),
            "-r",
            OUTPUT_DIR + "/*",
            f"{REMOTE}:{REMOTE_PATH}/",
        ],
        capture_output=True,
    )

    # Create nginx location if needed
    subprocess.run(
        [
            "ssh",
            "-o",
            "StrictHostKeyChecking=no",
            "-i",
            os.path.expanduser("~/.ssh/id_ed25519"),
            REMOTE,
            "mkdir -p " + REMOTE_PATH,
        ],
        capture_output=True,
    )

    print("Deployed to https://tormentnexus.site/blog/")


if __name__ == "__main__":
    main()
