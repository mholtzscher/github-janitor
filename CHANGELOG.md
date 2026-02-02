# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.3](https://github.com/mholtzscher/github-janitor/compare/v0.1.2...v0.1.3) (2026-02-02)


### Features

* add support for repository metadata and additional settings ([78aacd7](https://github.com/mholtzscher/github-janitor/commit/78aacd79092d7084181618eb5b5725731675a57a))
* **auth:** include token source in authentication output ([3762a75](https://github.com/mholtzscher/github-janitor/commit/3762a755667ddbe3499df91a62258e61de4b4f95))
* upgrade google/go-github to v82 and adapt to api changes ([bfe3f8f](https://github.com/mholtzscher/github-janitor/commit/bfe3f8fc27dd6188da649b3b9f72a574a340085e))


### Bug Fixes

* **plan:** add trailing space to result change arrow ([3098a8d](https://github.com/mholtzscher/github-janitor/commit/3098a8d860cca832187504252a68927a49307cae))

## [0.1.2](https://github.com/mholtzscher/github-janitor/compare/v0.1.1...v0.1.2) (2026-01-31)


### Bug Fixes

* **deps:** update dependencies and refine version test ([f2887e4](https://github.com/mholtzscher/github-janitor/commit/f2887e4cb75128d72511343cf623101930604b06))

## [0.1.1](https://github.com/mholtzscher/github-janitor/compare/v0.1.0...v0.1.1) (2026-01-31)


### Features

* **cli:** implement global options context propagation and improve test script setup ([571d517](https://github.com/mholtzscher/github-janitor/commit/571d517b38125d95de72e46d71486e91f495ebd4))
* implement core synchronization commands and settings engine ([80dd68c](https://github.com/mholtzscher/github-janitor/commit/80dd68ccbdd2553a8c184e3a7a97976dbd0cc18c))
* initialize project structure and basic CLI ([82990b6](https://github.com/mholtzscher/github-janitor/commit/82990b66a8dfeae8fc8af86cbbc3df31ab2b583a))

## [0.1.0](https://github.com/mholtzscher/github-janitor/releases/tag/v0.1.0) (YYYY-MM-DD)

### Features

- Initial release
- Basic CLI structure with urfave/cli/v3
- Example subcommand
- Nix flake support
- GitHub Actions CI/CD
