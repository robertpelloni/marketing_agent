#!/usr/bin/env python3
"""
Cross-post a markdown blog post to multiple platforms:
- Dev.to
- Hashnode
- Medium (requires manual paste - API is deprecated)
- Blogger (requires OAuth setup)

Usage:
    python3 crosspost_blog.py launch_materials/blog_post_autonomous_agent.md
"""

import sys
import os
import re
import requests
from pathlib import Path

# Load .env file
def load_env():
    env_path = Path(__file__).parent.parent / ".env"
    if env_path.exists():
        with open(env_path) as f:
            for line in f:
                line = line.strip()
                if "=" in line and not line.startswith("#"):
                    key, value = line.split("=", 1)
                    os.environ.setdefault(key.strip(), value.strip())

load_env()

def read_post(filepath):
    """Read markdown post and extract frontmatter + content."""
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()
    
    # Extract title from first H1
    title_match = re.search(r"^#\s+(.+)$", content, re.MULTILINE)
    title = title_match.group(1) if title_match else "Untitled"
    
    # Extract tags (look for common patterns)
    tags = []
    tag_patterns = ["AI", "Go", "Golang", "LLM", "OpenSource", "DevTools", "Automation", "SaaS"]
    for tag in tag_patterns:
        if tag.lower() in content.lower():
            tags.append(tag)
    
    return {
        "title": title,
        "content": content,
        "tags": tags[:4],  # Most platforms limit to 4 tags
    }

def post_to_devto(post):
    """Post to Dev.to using their API."""
    api_key = os.environ.get("DEVTO_API_KEY")
    if not api_key:
        print("⚠️  DEVTO_API_KEY not set, skipping Dev.to")
        return False
    
    headers = {
        "api-key": api_key,
        "Content-Type": "application/json",
    }
    
    data = {
        "article": {
            "title": post["title"],
            "body_markdown": post["content"],
            "tags": post["tags"],
            "published": True,
            "series": "Building TormentNexus",
        }
    }
    
    response = requests.post(
        "https://dev.to/api/articles",
        headers=headers,
        json=data
    )
    
    if response.status_code == 201:
        url = response.json()["url"]
        print(f"✅ Dev.to: {url}")
        return url
    else:
        print(f"❌ Dev.to error: {response.status_code} - {response.text[:200]}")
        return False

def post_to_hashnode(post):
    """Post to Hashnode using their GraphQL API."""
    api_key = os.environ.get("HASHNODE_API_KEY")
    publication_id = os.environ.get("HASHNODE_PUBLICATION_ID")
    
    if not api_key or not publication_id:
        print("⚠️  HASHNODE_API_KEY or HASHNODE_PUBLICATION_ID not set, skipping Hashnode")
        return False
    
    query = """
    mutation CreatePost($input: CreatePostInput!) {
        createPost(input: $input) {
            post {
                url
            }
        }
    }
    """
    
    variables = {
        "input": {
            "title": post["title"],
            "contentMarkdown": post["content"],
            "tags": [{"name": tag} for tag in post["tags"]],
            "publicationId": publication_id,
        }
    }
    
    headers = {
        "Authorization": api_key,
        "Content-Type": "application/json",
    }
    
    response = requests.post(
        "https://gql.hashnode.com",
        headers=headers,
        json={"query": query, "variables": variables}
    )
    
    if response.status_code == 200:
        data = response.json()
        if "errors" not in data:
            url = data["data"]["createPost"]["post"]["url"]
            print(f"✅ Hashnode: {url}")
            return url
        else:
            print(f"❌ Hashnode error: {data['errors'][0]['message']}")
            return False
    else:
        print(f"❌ Hashnode error: {response.status_code}")
        return False

def post_to_medium(post):
    """Post to Medium (manual - API is deprecated)."""
    print("⚠️  Medium API is deprecated. Manual posting required:")
    print("   1. Go to https://medium.com/new-story")
    print("   2. Paste the markdown content")
    print("   3. Add tags:", ", ".join(post["tags"]))
    return False

def post_to_ghost(post):
    """Post to Ghost using their API."""
    admin_key = os.environ.get("GHOST_ADMIN_API_KEY")
    site_url = os.environ.get("GHOST_SITE_URL")
    
    if not admin_key or not site_url:
        print("⚠️  GHOST_ADMIN_API_KEY or GHOST_SITE_URL not set, skipping Ghost")
        return False
    
    # Ghost API key format: {id}:{secret}
    key_id, key_secret = admin_key.split(":")
    
    # Create JWT token (simplified - in production use proper JWT)
    
    # For now, just print instructions
    print("⚠️  Ghost integration requires JWT setup. See Ghost docs.")
    return False

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 crosspost_blog.py <markdown_file>")
        sys.exit(1)
    
    filepath = sys.argv[1]
    if not os.path.exists(filepath):
        print(f"Error: File not found: {filepath}")
        sys.exit(1)
    
    post = read_post(filepath)
    print(f"📝 Post: {post['title']}")
    print(f"🏷️  Tags: {', '.join(post['tags'])}")
    print()
    
    results = []
    
    # Post to each platform
    for platform, func in [
        ("Dev.to", post_to_devto),
        ("Hashnode", post_to_hashnode),
        ("Ghost", post_to_ghost),
        ("Medium", post_to_medium),
    ]:
        print(f"📤 Posting to {platform}...")
        url = func(post)
        if url:
            results.append((platform, url))
        print()
    
    # Summary
    if results:
        print("=" * 50)
        print("✅ Published to:")
        for platform, url in results:
            print(f"   {platform}: {url}")
    else:
        print("=" * 50)
        print("⚠️  No API keys configured. Set these in .env:")
        print("   DEVTO_API_KEY=your_devto_key")
        print("   HASHNODE_API_KEY=your_hashnode_key")
        print("   HASHNODE_PUBLICATION_ID=your_publication_id")

if __name__ == "__main__":
    main()
