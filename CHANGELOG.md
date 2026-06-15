# Changelog

## [0.5.0] - 2026-06-15

### Added
- **Multi-Channel Outreach Engine**: Support for GitHub Issue/PR technical hook comments.
- **Real-World GitHub Sender**: Integration with `go-github` for automated outreach posting.
- **LinkedIn Outreach Scaffold**: Structured support for LinkedIn messaging with simulation fallback.
- **Dynamic CRM Field Mapping**: Configurable property names for Salesforce and HubSpot via environment variables.
- **Dossier Visibility**: Web dashboard now includes a "View Dossier" dropdown for each lead to show technical research findings.
- **CRM Provider Indicator**: Dashboard shows the active CRM provider and environment.
- **Fallback Enrichment**: Consolidated enrichment sources into a fallback chain with clear status reporting.
- **Cadence-Aware Scheduling**: Background scheduler for multi-touch outreach across Email, GitHub, and LinkedIn.

### Changed
- **Rebranded to TormentNexus**: Updated all product references and documentation.
- **Unified Configuration**: All environment variables consolidated into a typed Go struct.
- **Enhanced Logging**: Migrated core modules to structured `slog` JSON logging.
- **Improved Dashboard**: Lead table now shows truncated dossier snippets and outreach status.

### Fixed
- **CRLF Test Failures**: Resolved line-ending issues in git reconciliation tests.
- **Contact Email Uniqueness**: Fixed schema issue allowing duplicate NULL emails.
- **Worker Draining**: Improved graceful shutdown for all background routines.

## [0.4.9] - 2026-06-01

### Added
- **Hermes LLM Integration**: Real-world LLM provider via Hermes Agent gateway.
- **LLM Intent Classification**: Switched from keyword heuristics to LLM-powered classification.
- **Stripe Invoicing**: Production-ready Stripe integration for deal fulfillment.
- **Dual-Direction Merge Engine**: Intelligent branch reconciliation for autonomous development.

### Fixed
- **PostgreSQL Connection Pooling**: Added configurable limits to DB connections.
