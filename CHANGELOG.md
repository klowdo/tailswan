# Changelog

## [0.1.2](https://github.com/klowdo/tailswan/compare/v0.1.1...v0.1.2) (2026-04-15)


### Bug Fixes

* pass --reset to tailscale up to clear stale state ([f85d8e6](https://github.com/klowdo/tailswan/commit/f85d8e69c656c42430b22de46bf1911c91942ab0))

## [0.1.1](https://github.com/klowdo/tailswan/compare/v0.1.0...v0.1.1) (2026-04-15)


### Features

* Add TailSwan core implementation ([f3d58fa](https://github.com/klowdo/tailswan/commit/f3d58faba8a41389a3003bc215ad26cad70e94c6))
* Add web UI with PWA support ([324c989](https://github.com/klowdo/tailswan/commit/324c9892660d86c7faa8fb89ff94259a338233af))


### Bug Fixes

* bump Dockerfile Go version to 1.26.1 and add Renovate tracking ([ef3d264](https://github.com/klowdo/tailswan/commit/ef3d264ef39b87702b0437ca23fae34ed388ae2f))
* docker build workflow only verifies on PRs, pushes on tags ([4cdbad5](https://github.com/klowdo/tailswan/commit/4cdbad5268aafa85470bba082ac039ecd8c97eb9))
* migrate from deprecated tailscale.com/client/tailscale to tailscale.com/client/local ([c5e9093](https://github.com/klowdo/tailswan/commit/c5e9093afdeabc9a3940cafc5efc535d31e04375))
* Resolve data race in SSEHandler test ([37f25f0](https://github.com/klowdo/tailswan/commit/37f25f03683758d5d5f80a7329cb2987b826f676))
* suppress gosec G117 by excluding AuthKey from JSON serialization ([c4dd8d9](https://github.com/klowdo/tailswan/commit/c4dd8d9c903fb05154880ce736cf20e504984a60))
* update fang import path to charm.land/fang/v2 ([b3ed5e8](https://github.com/klowdo/tailswan/commit/b3ed5e8eb5c52f2ce2dff1da20f58afbe089e2de))
* use swanctl --stats for charon healthcheck ([339d08b](https://github.com/klowdo/tailswan/commit/339d08bb5b2acb1d58e772a1216bf722516cee27))
