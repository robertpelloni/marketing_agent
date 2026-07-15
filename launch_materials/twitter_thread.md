# Twitter/X Thread: Autonomous AI Agent

## Thread (post each as a separate tweet)

**Tweet 1 (Hook):**
I built an autonomous AI agent that sells itself.

It discovers leads, enriches contacts, writes personalized emails, and closes deals — all while I sleep.

Here's how it works 🧵

**Tweet 2 (The Problem):**
Developer tools are hard to sell.

You spend more time selling than building. Cold outreach is tedious. Personalization takes hours.

So I built an agent to do it for me.

**Tweet 3 (Lead Discovery):**
The agent scrapes Hacker News "Who is Hiring" threads and GitHub repos.

If a company is hiring for AI/ML roles, they probably need AI tooling.

It finds them automatically.

**Tweet 4 (Contact Enrichment):**
Once it finds a company, it needs decision-makers.

- Hunter.io API for email patterns
- GitHub commit scraping for developers
- LinkedIn browser automation for profiles

All automated.

**Tweet 5 (Personalized Outreach):**
Here's where it gets interesting.

The agent uses an LLM to write personalized emails based on:

- The company's tech stack
- Their recent GitHub activity
- Their job postings
- Their pain points

Each email is unique and relevant.

**Tweet 6 (Multi-Channel):**
It doesn't just send emails.

- Cold email via SMTP with proper DKIM/SPF
- Posts to Bluesky and LinkedIn
- Publishes blog posts
- Tracks opens, clicks, and replies

**Tweet 7 (Pipeline Management):**
The entire sales pipeline is tracked in PostgreSQL:

Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost

State transitions are atomic. Every interaction is logged.

**Tweet 8 (Billing):**
When a deal closes, Stripe handles the rest.

The agent creates customers, processes payments, and manages subscriptions — all autonomously.

**Tweet 9 (Results):**
In the first month:

- 2,000+ leads discovered
- 500+ personalized emails sent
- 50+ meaningful conversations started
- Multiple deals in pipeline
- Zero manual intervention

Costs $20/month to operate.

**Tweet 10 (The Meta Story):**
Here's the part that makes people's heads explode:

I used the autonomous agent to sell the autonomous agent.

The product markets itself. The sales agent sells itself.

It's turtles all the way down.

**Tweet 11 (Tech Stack):**
For those who want to build something similar:

- Go 1.24 (backend)
- PostgreSQL (database)
- MiMo v2.5 (LLM)
- Postfix (email)
- Browser automation (social)
- Stripe (payments)
- Hetzner VPS ($20/month)

**Tweet 12 (Open Source):**
TormentNexus is open-source under BSL 1.1.

The autonomous marketing agent is included.

GitHub: github.com/MDMAtk/TormentNexus

**Tweet 13 (CTA):**
If you're building AI tools and want to automate your sales pipeline, give it a try.

And if you have questions, the agent might answer them before I do.

Website: tormentnexus.site

---

## Hashtags (add to last tweet)

# AI #Go #OpenSource #DevTools #Automation #SaaS #BuildInPublic

## Thread Tips

- Post between 9-11 AM EST for best engagement
- Reply to your own thread with a summary
- Pin the thread to your profile
- Quote-tweet the first tweet with a screenshot of the dashboard
