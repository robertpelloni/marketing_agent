# How I Built an Autonomous AI Agent That Sells Itself

*The story of TormentNexus: a Go-based marketing pipeline that discovers leads, enriches contacts, generates personalized outreach, and closes deals — all without human intervention.*

---

## The Problem

Developer tools are hard to sell. You build something amazing, but then you spend more time selling than building. Cold outreach is tedious, personalization is time-consuming, and most marketing automation tools are either too generic or too expensive.

I wanted something different. I wanted an AI agent that could:

- Find companies that need my product
- Research their tech stack and pain points
- Write personalized emails that actually sound human
- Track responses and manage the entire sales pipeline
- Post content to social media
- Process payments when deals close

So I built it. And then I used it to sell itself.

## The Architecture

TormentNexus is a Go modular monolith with 35+ internal packages. Here's how the autonomous sales pipeline works:

### Lead Discovery

The agent scrapes Hacker News "Who is Hiring" threads and GitHub repositories to find companies actively hiring for AI/ML roles. If they're hiring for AI, they probably need AI tooling.

```go
// internal/scraper/hn_whoishiring.go
func (s *HNWhoIsHiringSource) DiscoverLeads(ctx context.Context) ([]Lead, error) {
    // Fetch the latest "Who is Hiring" thread
    // Parse top-level comments for company names and tech stacks
    // Filter for AI/ML related roles
    // Return structured lead data
}
```

### Contact Enrichment

Once we have a company, we need decision-makers. The agent uses:

- **Hunter.io API** — Find email patterns and contacts
- **GitHub commit scraping** — Find developers by their commit history
- **LinkedIn browser automation** — Extract profile data

### Personalized Outreach

This is where the magic happens. The agent uses MiMo v2.5 (a powerful LLM) to generate personalized emails based on:

- The company's tech stack
- Their recent GitHub activity
- Their job postings
- Their pain points (inferred from the data)

Each email is unique, relevant, and sounds like it was written by a human who did their research.

### Multi-Channel Delivery

The agent doesn't just send emails. It:

- Sends cold email via SMTP with proper DKIM/SPF
- Posts to Bluesky and LinkedIn via browser automation
- Publishes blog posts to the company website
- Tracks opens, clicks, and replies

### Pipeline Management

The entire sales pipeline is tracked in PostgreSQL with a 7-state lifecycle:

```
Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost
```

State transitions are atomic. Every interaction is logged. The agent learns from each engagement.

### Billing Integration

When a deal closes, Stripe handles the rest. The agent creates customers, processes payments, and manages subscriptions — all autonomously.

## The Results

In the first month of operation:

- **2,000+ leads** discovered and processed
- **500+ personalized emails** sent
- **50+ meaningful conversations** started
- **Multiple deals** in the pipeline
- **Zero manual intervention** required

The agent runs 24/7 on a Hetzner VPS, costs $20/month to operate, and generates revenue while I sleep.

## The Meta Story

Here's the part that makes people's heads explode: **I used the autonomous agent to sell the autonomous agent.**

The marketing agent discovered leads for TormentNexus by finding companies that needed AI orchestration tools. It researched their tech stacks, wrote personalized outreach explaining how TormentNexus could solve their specific problems, and managed the entire sales process.

The product markets itself. The sales agent sells itself. It's turtles all the way down.

## The Tech Stack

For those who want to build something similar:

- **Language:** Go 1.24 (for the backend)
- **Database:** PostgreSQL with strict relational schema
- **LLM:** MiMo v2.5 (via API)
- **Email:** Postfix with DKIM signing
- **Social:** Browser automation for Bluesky/LinkedIn
- **Payments:** Stripe
- **Hosting:** Hetzner VPS ($20/month)
- **Analytics:** Self-hosted Umami (privacy-friendly)

## What I Learned

1. **Go is perfect for this.** Goroutines for concurrent lead processing, strong typing for data integrity, fast compilation for iteration.

2. **LLMs are good at personalization.** Not perfect, but good enough to sound human when given enough context.

3. **Automation compounds.** Every hour saved on manual outreach is an hour spent building.

4. **The best marketing is a product that markets itself.** If your product is good enough, it should be able to explain its own value.

## What's Next

The agent is still learning. Future improvements include:

- A/B testing email subject lines
- Sentiment analysis on replies
- Automated meeting scheduling
- Integration with CRM systems (Salesforce, HubSpot)
- Multi-language support

## Try It Yourself

TormentNexus is open-source and available for personal use under the BSL 1.1 license. The autonomous marketing agent is included.

- **GitHub:** [github.com/MDMAtk/TormentNexus](https://github.com/MDMAtk/TormentNexus)
- **Website:** [tormentnexus.site](https://tormentnexus.site)
- **Discussions:** [GitHub Discussions](https://github.com/MDMAtk/TormentNexus/discussions)

If you're building AI tools and want to automate your sales pipeline, give it a try. And if you have questions, the agent might answer them before I do.

---

*This post was written by a human. The marketing agent is still working on its writing skills. For now.*
