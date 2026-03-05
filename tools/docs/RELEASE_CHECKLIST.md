# Release Checklist

This checklist must be completed before tagging a new release of `goConfy`.

## 1. Local Verification
- [ ] Run `make verify` from the project root.
  - This includes: `fmt-check`, `test`, `test-race`, and `vulncheck`.
  - All targets must pass for both the core module and the `tools/` module.
- [ ] Ensure integration tests in `tests/integration` pass.

## 2. Documentation Audit
- [ ] `README.md`: Check if examples still match the latest API.
- [ ] `CHANGELOG.md`: Ensure all changes are listed under a version section or `[Unreleased]`.
- [ ] `SECURITY_MODEL.md`: Ensure behavior matches documented security guarantees (e.g., dotenv).
- [ ] `CONFIG_POLICY.md`: Ensure enforced policies in tests match the documentation.

## 3. Tool Builds
- [ ] Build `goconfygen` and `goconfytui` locally to ensure no missing dependencies:
  ```bash
  cd tools
  go build ./cmd/goconfygen
  go build ./cmd/goconfytui
  ```

## 4. Release Preparation
- [ ] Bump version in `CHANGELOG.md` (move items from `[Unreleased]` to new version).
- [ ] Create a git tag with the version (e.g., `v0.2.1`).
- [ ] Push tag to origin: `git push origin v0.2.1`.

## 5. Dependency Check
- [ ] Run `go list -m all` to ensure no unwanted dependencies crept into the core module.
- [ ] Core module should have minimal external dependencies (only `gopkg.in/yaml.v3`).
