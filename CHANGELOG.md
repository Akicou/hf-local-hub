# Changelog

All notable changes to hf-local-hub will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-03-01

### Added
- Token authentication (shared secret)
- Hugging Face OAuth authentication
- LDAP authentication
- JWT token-based session management
- Login/logout CLI commands
- .env file support
- Web UI with auth integration

### Changed
- Makefile works from project root and server directory
- Updated Python package to v0.2.0

## [0.1.0] - 2026-03-01

### Added
- Go server with Gin framework and GORM ORM
- SQLite database for repository management
- REST API endpoints compatible with Hugging Face Hub
- Static file serving with Range support
- Python CLI with Typer framework
- Hugging Face Hub integration (huggingface_hub)
- Upload/download workflows (file and folder)
- Repository creation and listing
- Health check endpoint
- Docker support with multi-stage builds
- Comprehensive test suite (Go and Python)
- LFS stub endpoint
- CORS middleware
- Configuration via CLI flags and environment variables

### Security
- Safe path resolution to prevent directory traversal
- Input validation on all API endpoints

### Documentation
- Complete API reference
- Usage guide with examples
- Architecture documentation

## [0.0.1] - 2026-03-01

### Added
- Project initialization
- Basic directory structure
- README and LICENSE files
