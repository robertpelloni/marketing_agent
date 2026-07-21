#!/usr/bin/env python3
"""
Twitter Auto-Poster for HyperNexus
Posts promotional tweets about HyperNexus.
"""

import os
import json
import logging
import random
from pathlib import Path

import tweepy

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler("/opt/marketing_agent/logs/twitter_bot.log"),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger(__name__)

# Twitter API credentials
CONSUMER_KEY = os.getenv("TWITTER_API_KEY", "3cxJy4XTdlzhL0BWdPOkRzX7G")
CONSUMER_SECRET = os.getenv("TWITTER_API_SECRET", "ntRC30S1l07FiSZN6elkicl5OVjAu3AR60LlvFxexoLsiQs9l0")
ACCESS_TOKEN = os.getenv("TWITTER_ACCESS_TOKEN", "2079245743085223936-5DUc3zZOrRFBtUSVuk7tmv3cpHChVB")
ACCESS_SECRET = os.getenv("TWITTER_ACCESS_SECRET", "U69aKhP8QqVadE00ioeXfnN5Rv6vPT2yLghFHoDdcIEck")

# Data files
POSTED_FILE = Path("/opt/marketing_agent/data/twitter_posted.json")
POSTED_FILE.parent.mkdir(parents=True, exist_ok=True)

# Tweet templates
TWEET_TEMPLATES = [
    "Building AI agents that actually remember? HyperNexus's dual-tier memory architecture persists 14K+ memories across sessions. No more re-explaining your codebase. #AI #LLM #DevTools https://hypernexus.site",
    
    "Your AI agent is drowning in 50K tokens of tool definitions. HyperNexus's progressive MCP routing injects only the 3 most relevant tools per request. 95% context reduction. #AI #MCP https://hypernexus.site",
    
    "OpenAI rate-limited? No problem. HyperNexus cascades from cloud APIs to OpenRouter to local Ollama automatically. Zero downtime inference. #AI #DevOps https://hypernexus.site",
    
    "One config, six AI harnesses. Claude Code, Cursor, Codex, Gemini CLI, Copilot, Windsurf — byte-for-byte identical tool signatures. No vendor lock-in. #AI #DevTools https://hypernexus.site",
    
    "Local-first AI infrastructure: your team's knowledge stays on your machines. No cloud dependency. 14K+ persisted memories. Works offline. #AI #Privacy https://hypernexus.site",
    
    "Multi-agent orchestration done right. Planner, Implementer, Tester, Critic — all in one chatroom with shared transcript and consensus-driven decisions. #AI #Agents https://hypernexus.site",
    
    "Stop re-explaining your codebase every session. HyperNexus remembers decisions, preferences, and architectural choices. Persistent AI memory. #AI #Productivity https://hypernexus.site",
    
    "The waterfall approach: when your primary LLM fails, HyperNexus cascades to the next provider automatically. NVIDIA → OpenRouter → Ollama. #AI #Reliability https://hypernexus.site",
    
    "Cross-harness tool parity means your team can use their preferred AI tool. One config works everywhere. No vendor lock-in. #AI #Flexibility https://hypernexus.site",
    
    "14,726 memories that survive restarts. Your AI agent's knowledge is permanent, searchable, and instantly available. #AI #Memory https://hypernexus.site",
]


def load_posted():
    """Load set of already-posted tweet hashes."""
    try:
        if POSTED_FILE.exists():
            return set(json.loads(POSTED_FILE.read_text()))
    except (json.JSONDecodeError, OSError) as e:
        logger.warning(f"Error loading posted file: {e}")
    return set()


def save_posted(posted: set):
    """Save set of posted tweet hashes."""
    POSTED_FILE.write_text(json.dumps(list(posted)))


def get_twitter_client():
    """Create Twitter API client."""
    client = tweepy.Client(
        consumer_key=CONSUMER_KEY,
        consumer_secret=CONSUMER_SECRET,
        access_token=ACCESS_TOKEN,
        access_token_secret=ACCESS_SECRET,
        wait_on_rate_limit=True,
    )
    return client


def post_tweet(client, tweet_text: str) -> bool:
    """Post a tweet."""
    try:
        response = client.create_tweet(text=tweet_text)
        logger.info(f"Posted tweet: {response.data['id']}")
        return True
    except Exception as e:
        logger.error(f"Error posting tweet: {e}")
        return False


def main():
    """Main bot function."""
    logger.info("=== HyperNexus Twitter Bot ===")
    
    client = get_twitter_client()
    posted = load_posted()
    
    # Choose a tweet that hasn't been posted recently
    available = [t for t in TWEET_TEMPLATES if hash(t) not in posted]
    
    if not available:
        # Reset if all have been posted
        posted.clear()
        available = TWEET_TEMPLATES.copy()
    
    tweet_text = random.choice(available)
    
    logger.info(f"Posting tweet: {tweet_text[:80]}...")
    
    if post_tweet(client, tweet_text):
        posted.add(hash(tweet_text))
        save_posted(posted)
        logger.info("Tweet posted successfully!")
    else:
        logger.error("Failed to post tweet")
    
    logger.info(f"Done! Total posted: {len(posted)}")


if __name__ == "__main__":
    main()
