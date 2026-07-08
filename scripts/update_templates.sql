-- Update templates for dual-brand: HyperNexus (corp) / TormentNexus (dev)
-- The Go code auto-swaps HyperNexus → TormentNexus for non-corporate leads

UPDATE templates SET
  name = 'intro-email',
  subject = 'HyperNexus for {{company}} -- Quick Question',
  body = E'Hi {{contact}},\n\nI noticed {{company}} is building some really impressive things with {{tech_stack}}. The work your team is doing around AI orchestration is exciting.\n\nI''m reaching out because we''ve built HyperNexus — the enterprise-grade cognitive control plane that eliminates vendor lock-in. Our platform delivers:\n• Progressive MCP tool routing — injects only the 3 most relevant tools per request\n• Resilient LLM waterfall — NVIDIA → OpenRouter → local, zero-downtime failover\n• Persistent dual-tier memory — semantic search across 14K+ stored contexts\n• Cross-harness tool parity — one config works across Claude Code, Cursor, Copilot, Gemini CLI\n\nTeams like yours have seen 3-5x improvements in agent coordination efficiency and eliminated LLM API downtime.\n\nWould you be open to a quick 15-minute chat this week?\n\nBest,\n[Your Name]'
WHERE id = 'intro-email';

UPDATE templates SET
  name = 'followup-email',
  subject = 'Re: HyperNexus for {{company}} -- Thoughts?',
  body = E'Hi {{contact}},\n\nJust wanted to follow up on my previous note about HyperNexus.\n\nI know things get busy, so I''ll keep this brief: HyperNexus provides progressive MCP tool routing, dual-tier memory (14K+ persisted memories), and a resilient LLM waterfall that cascades across providers (NVIDIA → OpenRouter → local Ollama) with zero downtime.\n\nIf you''re even remotely curious about improving your agent coordination, I''d love to share a quick demo.\n\nWorth a conversation?\n\nBest,\n[Your Name]'
WHERE id = 'followup-email';

UPDATE templates SET
  name = 'breakup-email',
  subject = 'Should I close your file?',
  body = E'Hi {{contact}},\n\nI''ve reached out a few times about HyperNexus, but haven''t heard back.\n\nI''m guessing this isn''t a priority right now, or you''re swamped with other initiatives. Either way, I don''t want to be a pest.\n\nIf you''d like me to close your file on this, just reply "close". If I got the timing wrong and you''d still like to chat, hit me with a quick "yes" and we''ll find a time.\n\nNo hard feelings either way.\n\nBest,\n[Your Name]'
WHERE id = 'breakup-email';

UPDATE templates SET
  name = 'github-hook',
  subject = '',
  body = E'Hey @{{github_handle}}, I saw your work on {{repo}} — really impressive stuff! We''ve been tackling similar coordination challenges with HyperNexus (the enterprise cognitive control plane for multi-agent workflows). Would love to get your thoughts if you''re open to it.'
WHERE id = 'github-hook';

UPDATE templates SET
  name = 'linkedin-connect',
  subject = '',
  body = E'Hi {{contact}}, I came across your profile while researching teams working on {{tech_stack}} at {{company}}. Your background in {{role}} is impressive. I''d love to connect and exchange insights on AI infrastructure.'
WHERE id = 'linkedin-connect';
