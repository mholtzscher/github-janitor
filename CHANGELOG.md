# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.2](https://github.com/mholtzscher/github-janitor/compare/v0.2.1...v0.2.2) (2026-06-19)


### Features

* **config:** add mholtzscher/agent-artifacts repository ([c243782](https://github.com/mholtzscher/github-janitor/commit/c2437827e40a8c9a6a8073c7e260ffe161908293))

## [0.2.1](https://github.com/mholtzscher/github-janitor/compare/v0.2.0...v0.2.1) (2026-03-14)


### Features

* **actions-secrets:** add GitHub Actions secrets management support ([#16](https://github.com/mholtzscher/github-janitor/issues/16)) ([c4fd7ae](https://github.com/mholtzscher/github-janitor/commit/c4fd7aee6281f869e2a531d6683a6738079434f7))
* **github-janitor.yaml:** add actions_secrets configuration for PAT secret ([3b6eacf](https://github.com/mholtzscher/github-janitor/commit/3b6eacf8116290967201a3c3eb1cf0f960b777f7))
* **security:** add dependabot security settings sync ([#18](https://github.com/mholtzscher/github-janitor/issues/18)) ([c440804](https://github.com/mholtzscher/github-janitor/commit/c44080462b9e91f72063f0cabf58a4b5fe967e6e))


### Bug Fixes

* handle newlinesin tokens ([5bfd9b4](https://github.com/mholtzscher/github-janitor/commit/5bfd9b4222e852cd1a9d1455ad83efd36a8a6073))
* **secrets:** preserve spaces in environment variable values ([42f4ef2](https://github.com/mholtzscher/github-janitor/commit/42f4ef2e2174acf74fc6bc1045b8b53c90eb76c9))

## [0.2.0](https://github.com/mholtzscher/github-janitor/compare/v0.1.3...v0.2.0) (2026-02-20)


### ⚠ BREAKING CHANGES

* **cli:** rename sync subcommand to apply ([#11](https://github.com/mholtzscher/github-janitor/issues/11))

### Features

* **cli:** rename sync subcommand to apply ([#11](https://github.com/mholtzscher/github-janitor/issues/11)) ([19c2a8f](https://github.com/mholtzscher/github-janitor/commit/19c2a8f3db915b4971134d4baaa43dd6eb639651))

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
