**Overview:**

**Pre-merge:**
- [ ] Before merging to `master`, make sure that a new version hasn't been
  merged. Then, use `make bump-major`, `make bump-minor`, or `make bump-patch`
  to bump the version number in VERSION and version.go (and all files that
  reference gopkg.in if a major bump).

**Post-merge:**
- [ ] After merging to `master`, Use `make tag-version` to set the version
  number in a git tag. Then, run `git push --tags` to push the new tag.
