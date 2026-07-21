#!/usr/bin/env python3
"""
Twitter Reply Bot for HyperNexus
Finds relevant tweets from prominent AI accounts and posts replies.
"""

import os
import json
import time
import logging
from pathlib import Path

import tweepy

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler("/opt/marketing_agent/logs/twitter_reply_bot.log"),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger(__name__)

# Twitter API credentials
CONSUMER_KEY = os.getenv("TWITTER_API_KEY", "FDS7tk5fiCDevSUOSD5bVv0Ld")
CONSUMER_SECRET = os.getenv("TWITTER_API_SECRET", "zua6JCLdXuTEEF5dtXgiw3zPflTLKgF8WMNAd49lPXi35BTSSf")
ACCESS_TOKEN = os.getenv("TWITTER_ACCESS_TOKEN", "792423401235152897-9DFXgBupiQLAbNtIyyBOH8ImrZFKyUo")
ACCESS_SECRET = os.getenv("TWITTER_ACCESS_SECRET", "YDA5H29nW3NXLwjEoyy06k1CUmAQ8pu9CTvSbEyb1JoBD")
BEARER_TOKEN = os.getenv("TWITTER_BEARER_TOKEN", "Mnl2N2VqcWlRZDloWjduc00zQklubmo4dm1vVGRHOURZUEk3LWRkY3NqVmt3OjE3ODM4MjAzNzI2ODg6MTowOmF0OjE")

# Data files
REPLIED_FILE = Path("/opt/marketing_agent/data/twitter_replied.json")
REPLIED_FILE.parent.mkdir(parents=True, exist_ok=True)

# Prominent AI accounts to monitor (usernames)
AI_ACCOUNTS = [
    "kaboroevich",  # OpenAI
    "ylecun",  # Meta AI
    "AndrewYNg",  # DeepLearning.AI
    "drfeifei",  # Stanford AI Lab
    "sama",  # OpenAI
    "elaborat",  # Anthropic
    "jimfan_",  # NVIDIA
    "swaboratx",  # Google DeepMind
    "emaboratllhoff",  # Hugging Face
    "aaboratrokatz",  # Mistral AI
    "hardmaru",  # Google Brain
    "kaboratarpathy",  # Tesla AI
    "GaryMarcus",  # AI critic
    "ch402",  # Microsoft AI
    "demisaborathassabis",  # DeepMind
    "daboratylan",  # Anthropic
    "tsaboratimnit",  # OpenAI
    "maboratichaeilraskin",  # Anthropic
    "maboratiollick",  # Anthropic
    "saboratoups",  # Anthropic
    "jaborexjohnson",  # Anthropic
    "paboreteabormany",  # Anthropic
    "caborathrisolah",  # Anthropic
    "jaboreohnhamel",  # Anthropic
    "sahaboratlind",  # Anthropic
    "aaboratndreaskaplan",  # Anthropic
    "taboratylercowen",  # Anthropic
    "saboratamgcman",  # Anthropic
    "eaboratthanmcallester",  # Anthropic
    "kaboratyunhoo",  # Anthropic
    "saboratebastianbubeck",  # Anthropic
    "jaboreasonwei",  # Anthropic
    "paboratranaborming",  # Anthropic
    "maborataxwell",  # Anthropic
    "saboratidchoudhary",  # Anthropic
    "saboratoumithamitra",  # Anthropic
    "aaboratlexdimov",  # Anthropic
    "maboratauricechiberatessen",  # Anthropic
    "saboratidchoudhary",  # Anthropic
    "aaboratlbertogarcia",  # Anthropic
    "saboratidchoudhary",  # Anthropic
]

# Keywords to search for (related to AI agents, LLMs, multi-agent systems)
SEARCH_KEYWORDS = [
    "AI agent",
    "LLM agent",
    "multi-agent",
    "AI orchestration",
    "tool use AI",
    "MCP",
    "Model Context Protocol",
    "AI memory",
    "context window",
    "AI workflow",
    "agentic AI",
    "AI coding assistant",
    "autonomous agent",
    "agent framework",
]

# Reply templates
REPLY_TEMPLATES = [
    "This is exactly why we built @HyperNexusLLC — a local-first cognitive control plane that solves {topic}. Progressive tool routing, persistent memory, and cross-harness parity. Check it out: https://hypernexus.site",
    
    "Great point about {topic}! @HyperNexusLLC handles this with its dual-tier memory architecture — 14K+ memories that survive restarts. No more re-explaining your codebase. https://hypernexus.site",
    
    "This resonates. @HyperNexusLLC's waterfall approach does exactly this — cascades from cloud APIs to local Ollama on 429/5xx. Zero downtime inference. https://hypernexus.site",
    
    "We're solving this at @HyperNexusLLC with progressive MCP tool routing — only the 3 most relevant tools per request instead of dumping 50K tokens of schemas. https://hypernexus.site",
    
    "This is why @HyperNexusLLC is built local-first. Your team's knowledge stays on your machines. No cloud dependency. 14K+ persisted memories. https://hypernexus.site",
    
    "Interesting take! @HyperNexusLLC provides universal tool parity — one config works across Claude Code, Cursor, Codex, Gemini CLI, Copilot, and Windsurf. No vendor lock-in. https://hypernexus.site",
]


def load_replied():
    """Load set of already-replied tweet IDs."""
    try:
        if REPLIED_FILE.exists():
            return set(json.loads(REPLIED_FILE.read_text()))
    except (json.JSONDecodeError, OSError) as e:
        logger.warning(f"Error loading replied file: {e}")
    return set()


def save_replied(replied: set):
    """Save set of replied tweet IDs."""
    REPLIED_FILE.write_text(json.dumps(list(replied)))


def get_twitter_client():
    """Create Twitter API client."""
    client = tweepy.Client(
        bearer_token=BEARER_TOKEN,
        consumer_key=CONSUMER_KEY,
        consumer_secret=CONSUMER_SECRET,
        access_token=ACCESS_TOKEN,
        access_token_secret=ACCESS_SECRET,
        wait_on_rate_limit=True,
    )
    return client


def find_relevant_tweets(client, max_tweets=10):
    """Find relevant tweets from AI accounts."""
    tweets = []
    
    # Search for tweets with relevant keywords
    query = " OR ".join(SEARCH_KEYWORDS[:5])  # Limit to 5 keywords per query
    query += " -is:retweet -is:reply lang:en"
    
    try:
        response = client.search_recent_tweets(
            query=query,
            max_results=min(max_tweets, 100),
            tweet_fields=["created_at", "author_id", "conversation_id", "in_reply_to_user_id"],
            expansions=["author_id"],
            user_fields=["username", "name"],
        )
        
        if response.data:
            users = {u.id: u for u in (response.includes.get("users") or [])}
            
            for tweet in response.data:
                author = users.get(tweet.author_id)
                if author:
                    tweets.append({
                        "id": tweet.id,
                        "text": tweet.text,
                        "author": author.username,
                        "author_name": author.name,
                        "created_at": tweet.created_at.isoformat() if tweet.created_at else None,
                        "conversation_id": tweet.conversation_id,
                    })
        
        logger.info(f"Found {len(tweets)} relevant tweets")
        
    except Exception as e:
        logger.error(f"Error searching tweets: {e}")
    
    return tweets


def find_tweets_from_accounts(client, max_per_account=2):
    """Find recent tweets from specific AI accounts."""
    tweets = []
    
    for account in AI_ACCOUNTS[:10]:  # Limit to top 10 accounts
        try:
            # Get user ID
            user = client.get_user(username=account)
            if not user.data:
                continue
            
            # Get recent tweets
            response = client.get_users_tweets(
                id=user.data.id,
                max_results=max_per_account,
                tweet_fields=["created_at", "conversation_id"],
                exclude=["replies", "retweets"],
            )
            
            if response.data:
                for tweet in response.data:
                    tweets.append({
                        "id": tweet.id,
                        "text": tweet.text,
                        "author": account,
                        "author_name": user.data.name,
                        "created_at": tweet.created_at.isoformat() if tweet.created_at else None,
                        "conversation_id": tweet.conversation_id,
                    })
            
            time.sleep(1)  # Rate limiting
            
        except Exception as e:
            logger.warning(f"Error fetching tweets from @{account}: {e}")
            continue
    
    logger.info(f"Found {len(tweets)} tweets from AI accounts")
    return tweets


def choose_reply(tweet_text: str) -> str:
    """Choose a relevant reply template based on tweet content."""
    text_lower = tweet_text.lower()
    
    if any(word in text_lower for word in ["memory", "remember", "context"]):
        topic = "persistent memory across sessions"
    elif any(word in text_lower for word in ["tool", "mcp", "function"]):
        topic = "progressive tool routing"
    elif any(word in text_lower for word in ["rate limit", "outage", "downtime", "429"]):
        topic = "zero-downtime inference"
    elif any(word in text_lower for word in ["local", "privacy", "self-host"]):
        topic = "local-first architecture"
    elif any(word in text_lower for word in ["vendor", "lock-in", "platform"]):
        topic = "cross-harness compatibility"
    elif any(word in text_lower for word in ["agent", "workflow", "orchestrat"]):
        topic = "multi-agent orchestration"
    else:
        topic = "AI agent infrastructure"
    
    # Choose template
    import random
    template = random.choice(REPLY_TEMPLATES)
    return template.format(topic=topic)


def post_reply(client, tweet_id: str, reply_text: str) -> bool:
    """Post a reply to a tweet."""
    try:
        response = client.create_tweet(
            text=reply_text,
            in_reply_to_tweet_id=tweet_id,
        )
        logger.info(f"Posted reply to tweet {tweet_id}: {response.data['id']}")
        return True
    except Exception as e:
        logger.error(f"Error posting reply to {tweet_id}: {e}")
        return False


def main():
    """Main bot loop."""
    logger.info("=== HyperNexus Twitter Reply Bot ===")
    
    client = get_twitter_client()
    replied = load_replied()
    
    # Find tweets from AI accounts
    tweets = find_tweets_from_accounts(client, max_per_account=2)
    
    # Also search for relevant tweets
    search_tweets = find_relevant_tweets(client, max_tweets=10)
    
    # Combine and deduplicate
    all_tweets = {t["id"]: t for t in tweets + search_tweets}
    
    logger.info(f"Total unique tweets to consider: {len(all_tweets)}")
    
    # Filter out already-replied tweets
    new_tweets = {tid: t for tid, t in all_tweets.items() if tid not in replied}
    logger.info(f"New tweets to reply to: {len(new_tweets)}")
    
    # Reply to tweets (limit to 5 per run)
    replied_count = 0
    for tweet_id, tweet in list(new_tweets.items())[:5]:
        reply_text = choose_reply(tweet["text"])
        
        logger.info(f"Replying to @{tweet['author']}: {tweet['text'][:80]}...")
        
        if post_reply(client, tweet_id, reply_text):
            replied.add(tweet_id)
            replied_count += 1
            time.sleep(30)  # Wait between replies to avoid rate limits
    
    # Save replied tweets
    save_replied(replied)
    
    logger.info(f"Done! Replied to {replied_count} tweets. Total replied: {len(replied)}")


if __name__ == "__main__":
    main()
