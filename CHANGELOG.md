# Changelog

## [0.6.0] - 2026-06-15

### Added
- **LLM-Powered AutoDev**: Implementation of real code generation in `LocalAgent` using the configured LLM provider.
- **UAT Simulation Portal**: Interactive dashboard component to simulate inbound messages from leads and verify autonomous responses.
- **Lead Contact Lookup**: Added `GetContactByID` to the database repository.

### Changed
- **Web Server Architecture**: Updated `web.NewServer` to accept `Communication Manager` for simulation orchestration.

### Fixed
- **CI & Security Hardening**: Resolved multiple gosec and linting issues including Slowloris protection and secure cookie handling.

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
- **Technical Blog Ingestion**: `BlogWorker` polls engineering blogs via RSS to detect technical signals.
- **Competitor Tracking**: `LearningSalesEngine` now incorporates competitor mentions into lead scoring.

### Changed
- **Rebranded to TormentNexus**: Updated all product references and documentation.
- **Unified Configuration**: All environment variables consolidated into a typed Go struct.
- **Enhanced Logging**: Migrated core modules to structured `slog` JSON logging.

### Fixed
- **CRLF Test Failures**: Resolved line-ending issues in git reconciliation tests.
- **Contact Email Uniqueness**: Fixed schema issue allowing duplicate NULL emails.
- **Worker Draining**: Improved graceful shutdown for all background routines.
