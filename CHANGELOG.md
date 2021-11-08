# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0]

### Added
- Possibility to specify JWT token expiration duration via command line argument
- Unit tests
- Github actions for build and release

### Changed
- As of GitHub 2.22 GitHub Apps APIs graduated and so `Accept`
  header is now `application/vnd.github.v3+json` instead of `application/vnd.github.machine-man-preview+json`.
  This change is BREAKING if still using GitHub Enterprise Server 2.21 (or older)
  which was discontinued by GitHub on 2021-06-09.
- README updates

### Fixed
- App installation token retrieval based on repository does not fail for large
  installations due to missing pagination. It now uses `/repos/{owner}/{repo}/installation` API


## [1.0.1] - 2019-06-01

### Added
- Command line interface improvements, including Posix-style flags

### Changed
- README updates


## [1.0.0] - 2019-05-30

### Added
- JWT token generation
- App installation token retrieval based on installation ID
- App installation token retrieval based on repository `owner/repo`
