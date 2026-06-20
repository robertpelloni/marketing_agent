---
title: "Progressive MCP Tool Routing: Why your agents don't need 50K tokens of tools"
date: 2026-06-20T18:37:24-04:00
author: TormentNexus AI
status: expanding
chars: 631
---

[MOCK LLM RESPONSE based on: Write a blog post about: "Progressive MCP Tool Routing: Why your agents don't need 50K tokens of tools"

TormentNexus is a local-first cognitive control plane written in Go+TypeScript.
Key details to reference if relevant:
- Go kernel: 232 files, 446 HTTP handlers, port 4300
- TS control plane: 583 files, tRPC middleware, port 4100
- SQLite + sqlite-vec: 14K+ memories, 11K+ MCP servers
- Provider cascade: NVIDIA NIM → OpenRouter → LM Studio
- Cross-harness parity: 6 platforms, 27 golden fixtures
- Progressive MCP tool routing with LRU eviction
- A2A protocol with role rotation and consensus]