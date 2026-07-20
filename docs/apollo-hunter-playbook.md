# Apollo + Hunter Prospecting Playbook for TormentNexus

## Target Personas

### Primary Targets (High Intent)

1. **CTOs / VPs of Engineering** at AI-first startups (10-200 employees)
2. **AI/ML Engineering Leads** at companies using LLMs
3. **DevTools / Platform Engineering** at companies building AI products
4. **Technical Founders** at AI startups (Series A-C)

### Secondary Targets (Medium Intent)

1. **Senior Software Engineers** at companies with AI initiatives
2. **Engineering Managers** at companies adopting AI tools
3. **DevOps / Platform Engineers** at companies with complex toolchains

### Company Signals (What to Look For)

- Uses Claude, Cursor, Codex, or other AI coding tools
- Recently raised funding (have budget)
- Hiring AI/ML engineers (active AI initiatives)
- Building internal tools (need orchestration)
- Has multiple AI models in use (need routing)

---

## Apollo.io Strategy

### 1. Build Company Lists

**Search Filters:**

- Industry: Software, Technology, AI/ML
- Company Size: 10-500 employees
- Funding: Series A or later
- Technology: Python, TypeScript, Go (tech stack match)
- Location: US, UK, Canada, EU (English-speaking markets)

**Saved Searches:**

1. "AI Startups - Series A+" (funded, building AI)
2. "DevTools Companies" (likely using AI internally)
3. "Companies Hiring AI Engineers" (active AI initiatives)

### 2. Build Contact Lists

**Search Filters:**

- Job Title: CTO, VP Engineering, AI Lead, ML Engineer, Platform Engineer
- Seniority: Director, VP, C-Level
- Department: Engineering, AI/ML, Product
- Company: From saved company lists above

### 3. Use Apollo Sequences

**Sequence 1: Cold Outreach (Week 1)**

- Email 1: Value prop + pain point
- Email 2: Case study / social proof (Day 3)
- Email 3: Demo offer (Day 7)

**Sequence 2: Nurture (Week 2+)**

- Email 4: Technical deep dive
- Email 5: Community invite
- Email 6: Limited offer

### 4. Apollo Intent Signals

**Monitor:**

- Companies searching for "AI agent framework"
- Companies searching for "MCP tools"
- Companies searching for "multi-agent orchestration"
- Companies visiting competitor websites

---

## Hunter.io Strategy

### 1. Domain Search

**For each target company:**

1. Search domain in Hunter
2. Find all email addresses
3. Verify deliverability
4. Export to CSV

### 2. Email Verification

**Before sending any outreach:**

1. Verify all emails from Apollo
2. Remove invalid/bounced emails
3. Keep only "valid" and "accept_all" emails

### 3. Bulk Operations

**Monthly tasks:**

1. Export Apollo contact list
2. Bulk verify with Hunter
3. Remove invalid emails
4. Import clean list back to Apollo

---

## Target Company Lists

### Tier 1: High-Value Targets

1. **AI Coding Tools**: Cursor, Windsurf, Cody, Continue, Aider
2. **LLM Platforms**: OpenRouter, Together AI, Anyscale, Modal
3. **AI Agent Frameworks**: LangChain, CrewAI, AutoGen, Semantic Kernel
4. **DevTools with AI**: Vercel, Netlify, Railway, Render

### Tier 2: Potential Customers

1. **AI Startups**: Any company building with LLMs
2. **Enterprise AI**: Companies with AI teams >10 people
3. **Consulting Firms**: AI/ML consultancies
4. **Agencies**: Digital agencies using AI tools

### Tier 3: Long Tail

1. **Indie Hackers**: Solo developers using AI
2. **Open Source Maintainers**: Popular AI repos
3. **Content Creators**: AI-focused YouTubers, bloggers
4. **Community Leaders**: Discord/Slack admins for AI communities

---

## Email Templates

### Template 1: Pain Point Hook

```
Subject: Your AI agents are drowning in tool dumps

Hi [Name],

I noticed [Company] is building with AI/LLMs. Quick question:

Are your agents overwhelmed by 50K-token tool dumps?

We built TormentNexus to solve exactly this. Our progressive MCP 
tool routing injects only the 3 most relevant tools per request 
instead of dumping everything into context.

Result: 10x faster inference, 90% fewer hallucinations.

Would love to show you a quick demo. Interested?

Best,
[Your Name]
```

### Template 2: Technical Hook

```
Subject: One config, identical tools across Claude, Cursor, Codex

Hi [Name],

Quick question: Are you maintaining separate tool configs for 
Claude Code, Cursor, Codex, and Gemini CLI?

TormentNexus provides byte-for-byte identical tool signatures 
across all major AI coding harnesses. One config, universal parity.

Plus: LLM waterfall failover (OpenAI → OpenRouter → local Ollama 
on any 429/5xx error).

Worth a look? https://hypernexus.site

Best,
[Your Name]
```

### Template 3: Social Proof Hook

```
Subject: How [Similar Company] reduced AI costs 40%

Hi [Name],

[Similar Company] was spending $X/month on AI API calls. After 
implementing TormentNexus's LLM waterfall failover, they reduced 
costs 40% by cascading to local Ollama on rate limits.

Key features:
- Progressive MCP tool routing (50K tokens → 3 tools)
- Persistent memory across sessions (14K+ memories)
- Universal tool parity (Claude, Cursor, Codex, Gemini)

Would this help [Company]? Happy to show you how it works.

Best,
[Your Name]
```

---

## Execution Plan

### Week 1: Build Lists

- [ ] Create 5 Apollo company lists
- [ ] Create 5 Apollo contact lists
- [ ] Export all contacts
- [ ] Bulk verify with Hunter
- [ ] Import clean lists back to Apollo

### Week 2: Launch Sequences

- [ ] Set up Sequence 1 (Cold Outreach)
- [ ] Set up Sequence 2 (Nurture)
- [ ] Launch first batch (100 contacts)
- [ ] Monitor open/reply rates

### Week 3: Optimize

- [ ] A/B test subject lines
- [ ] Refine messaging based on replies
- [ ] Expand to new company lists
- [ ] Launch Sequence 2 for engaged contacts

### Week 4: Scale

- [ ] Increase volume (200-500 contacts/week)
- [ ] Add new sequences for different personas
- [ ] Track conversion to sales
- [ ] Calculate CAC and ROI

---

## Key Metrics to Track

| Metric | Target | Tool |
|--------|--------|------|
| Open Rate | >40% | Apollo |
| Reply Rate | >5% | Apollo |
| Meeting Booked | >2% | Apollo |
| Email Validity | >95% | Hunter |
| Bounce Rate | <3% | Hunter |
| CAC | <$50 | Manual |

---

## Monthly Tasks

1. **Week 1**: Refresh company/contact lists
2. **Week 2**: Verify all new emails with Hunter
3. **Week 3**: Launch new sequences
4. **Week 4**: Analyze metrics, optimize

---

## Integration with Marketing Agent

The marketing_agent Go code already has:

- `internal/enrichment/apollo.go` - Apollo API integration
- `internal/enrichment/hunter.go` - Hunter API integration
- `internal/scraper/` - Lead discovery
- `internal/communication/` - Outreach sequences

To use these:

1. Set environment variables (already done)
2. Run the marketing agent
3. It will automatically discover, enrich, and outreach

---

## Apollo.io Advanced Features to Use

### 1. **Buying Intent Signals**

- Filter companies by intent topics: "AI", "Machine Learning", "Developer Tools"
- Prioritize outreach to high-intent companies

### 2. **Technology Tracking**

- See what technologies companies use
- Target companies using competitor tools

### 3. **Funding Data**

- Filter by recent funding rounds
- Companies with fresh funding have budget

### 4. **Job Postings**

- Companies hiring AI/ML engineers = active AI initiatives
- Use as outreach trigger: "Saw you're hiring AI engineers..."

### 5. **Email Sequences with A/B Testing**

- Test different subject lines
- Test different value props
- Optimize based on data

### 6. **CRM Sync**

- Sync Apollo contacts to your CRM
- Track full funnel from lead to close

---

## Hunter.io Advanced Features to Use

### 1. **Bulk Email Verification**

- Upload CSV of emails
- Get deliverability score
- Remove invalid emails

### 2. **Domain Search**

- Find all emails at a company
- See confidence scores
- Get department breakdown

### 3. **Email Finder**

- Find specific person's email
- Use name + domain
- High accuracy for business emails

### 4. **Campaigns**

- Send cold emails directly from Hunter
- Track opens and replies
- Automate follow-ups

### 5. **API Integration**

- Integrate with marketing_agent
- Automate verification
- Real-time enrichment

---

## Budget Allocation

| Tool | Monthly Cost | Usage |
|------|--------------|-------|
| Apollo.io | $99-199/mo | Prospecting, sequences, intent |
| Hunter.io | $49-99/mo | Email verification, domain search |
| **Total** | **$148-298/mo** | Full prospecting stack |

Expected ROI: 5-10x (based on $50/year product price)
