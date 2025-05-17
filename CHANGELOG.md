# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2025-05-17

### Added
- Initial release
- HTTP/HTTPS traffic interception
- Dynamic TLS certificate generation
- Rule-based system for modifying requests and responses
- Support for custom Go scripts
- WebSocket connection handling
- Environment variable integration
- Comprehensive documentation in the `docs` directory
- Support for mkcert certificate generation
- Binary releases for multiple platforms (macOS Intel/ARM, Linux, Windows)

### Changed
- Improved error handling for missing rules directory

### Fixed
- Fixed handling of non-existent rules directory
