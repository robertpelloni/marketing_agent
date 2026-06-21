#!/usr/bin/env python3
"""Rewrite each blog post using FreeLLM, one at a time, waiting indefinitely."""
import os
import json
import urllib.request
import time
import re

FREELLM_URL = "http://localhost:4000/v1/chat/completions"
BLOG_DIR = "docs/blog"

TOPICS = [
    "Progressive MCP Tool Routing: Why your agents don't need 50K tokens of tools",
    "Cross-Harness Parity: One config for Claude, Codex, Cursor, Copilot, Gemini, Kiro",
    "Local-First AI: Why your LLM infrastructure shouldn't depend on the cloud",
    "The LLM Waterfall: Zero-downtime inference with NVIDIA → OpenRouter → Local fallback",
    "Dual-Tier Memory Architecture: 14K+ persistent memories that survive restarts",
    "Multi-Agent Swarms: How Planner-Implementer-Tester-Critic collaboration works",
    "Self-Healing AI: The diagnose-fix-verify-retry closed loop",
    "Why TormentNexus uses sqlite-vec for vector search instead of Pinecone",
    "Progressive Disclosure: The anti-pattern of dumping every tool into context",
    "Agent-to-Agent Protocol: How AI agents negotiate and debate in shared chatrooms",
    "The Go Sidecar Pattern: Why we wrote our AI kernel in Go, not Python",
    "Cross-Harness Tool Parity: Byte-for-byte identical tools across 6 AI coding platforms",
    "Local-First Memory: Why your AI shouldn't need the cloud to remember",
    "11K MCP Servers: The largest indexed catalog of AI tools",
    "From NVIDIA NIM to Local Ollama: Building a resilient LLM provider cascade",
]

def slugify(topic):
    s = topic.lower().replace(" ", "-").replace(":", "").replace("'", "").replace("/", "-").replace("→", "to")
    return s[:60]

def call_llm(prompt, max_tokens=4096, temperature=0.7):
    """Call FreeLLM, retrying indefinitely until success."""
    data = {
        "model": "free-llm",
        "messages": [
            {"role": "system", "content": "You are a technical writer for TormentNexus. Write clear, authoritative blog posts about AI infrastructure. Be specific, technical, and compelling. Output clean markdown."},
            {"role": "user", "content": prompt}
        ],
        "max_tokens": max_tokens,
        "temperature": temperature,
    }
    req = urllib.request.Request(
        FREELLM_URL,
        data=json.dumps(data).encode(),
        headers={"Content-Type": "application/json", "Authorization": "Bearer sk-litellm"},
    )
    while True:
        try:
            with urllib.request.urlopen(req, timeout=600) as resp:
                result = json.loads(resp.read())
                if "choices" in result and len(result["choices"]) > 0:
                    content = result["choices"][0]["message"]["content"]
                    # Strip reasoning prefixes
                    content = re.sub(r'\[Model:[^\]]+\]\s*\n\s*', '', content)
                    content = re.sub(r'\[Continued with Model:[^\]]+\]\s*\n\s*', '', content)
                    return content.strip()
        except Exception as e:
            print(f"  LLM call failed: {e}. Retrying in 10s...")
            time.sleep(10)

def process_post(topic):
    slug = slugify(topic)
    path = os.path.join(BLOG_DIR, slug + ".md")
    
    print(f"\n{'='*60}")
    print(f"Processing: {topic}")
    print(f"File: {path}")
    print(f"{'='*60}")
    
    # Step 1: Generate initial post (target 3000 chars)
    print("\n[1/3] Generating initial content...")
    content = call_llm(
        f"Write a technical blog post about: \"{topic}\"\n\n"
        f"TormentNexus is a local-first cognitive control plane written in Go+TypeScript.\n"
        f"Target length: 3000-4000 characters. Include technical details, architecture, and specific insights.\n"
        f"Format as markdown with sections.",
        max_tokens=4096
    )
    
    if not content:
        print("  Failed to generate content, skipping")
        return False
    
    # Step 2: Expand to 5000+ chars
    print("\n[2/3] Expanding content...")
    expanded = call_llm(
        f"Continue and expand this blog post. Add more technical depth, use cases, and details.\n"
        f"Current length: ~{len(content)} chars. Target: 5000+ chars.\n"
        f"DO NOT repeat what's already written. Add NEW sections.\n\n"
        f"CURRENT POST:\n{content[:2000]}\n\n... [content truncated] ...\n\n"
        f"Write the next section(s) that follow naturally.",
        max_tokens=4096
    )
    
    full_content = content + "\n\n" + (expanded or "")
    
    # Step 3: Final polish
    print("\n[3/3] Polishing...")
    polished = call_llm(
        f"Polish this blog post. Fix grammar, improve flow, ensure consistency.\n"
        f"Remove any markdown formatting issues. Make it read like a professional engineering blog.\n\n"
        f"FULL POST:\n{full_content}",
        max_tokens=4096
    )
    
    final = polished or full_content
    final_len = len(final)
    
    # Write the file
    header = f"""---
title: "{topic}"
date: {time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())}
author: TormentNexus AI
status: polished
chars: {final_len}
---

"""
    with open(path, "w", encoding="utf-8") as f:
        f.write(header + final)
    
    print(f"\n✅ Done: {final_len} chars written to {path}")
    return True

def main():
    # Process posts one at a time
    for topic in TOPICS:
        slug = slugify(topic)
        path = os.path.join(BLOG_DIR, slug + ".md")
        
        # Check if already processed (has "polished" status)
        already_polished = False
        if os.path.exists(path):
            with open(path) as f:
                if "status: polished" in f.read():
                    already_polished = True
        
        if already_polished:
            print(f"⏭️  Skipping {topic} (already polished)")
            continue
        
        process_post(topic)
        print("\n⏱️  Waiting 5 seconds before next post...")
        time.sleep(5)
    
    print(f"\n{'='*60}")
    print("All posts processed!")
    print(f"{'='*60}")

if __name__ == "__main__":
    main()
