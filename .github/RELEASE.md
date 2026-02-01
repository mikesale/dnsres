# Release Checklist for dnsres

This document provides a step-by-step checklist for releasing new versions of dnsres.

## Pre-Release Testing

### Local Testing
- [ ] Run all unit tests: `make test`
- [ ] Run integration tests: `go test -tags=integration ./internal/integration -v`
- [ ] Run linters: `make lint`
- [ ] Test CLI binary locally: `make build && ./dnsres example.com`
- [ ] Test TUI binary locally: `make build-tui && ./dnsres-tui example.com`
- [ ] Verify configuration file loading from all XDG locations
- [ ] Test with both valid and invalid configurations

### GoReleaser Testing
- [ ] Test snapshot build: `goreleaser build --snapshot --clean`
- [ ] Verify binaries in `dist/` directory work
- [ ] Check that all platforms build successfully (darwin/linux/windows, amd64/arm64)
- [ ] Verify archive contents include:
  - [ ] Both binaries (dnsres, dnsres-tui)
  - [ ] README.md, LICENSE, INSTALL.md
  - [ ] examples/config.json
  - [ ] completions/ directory

### Package Testing (Optional but Recommended)
- [ ] Test .deb package: `sudo dpkg -i dist/*.deb`
- [ ] Test .rpm package: `sudo rpm -i dist/*.rpm`
- [ ] Verify both binaries are in PATH after package install
- [ ] Verify completions are installed
- [ ] Verify example config is at /etc/dnsres/config.json.example

### Documentation Review
- [ ] Update version numbers in INSTALL.md if needed
- [ ] Review README.md for accuracy
- [ ] Check that all links work
- [ ] Verify code examples are correct

## Version Preparation

### Version Number
- [ ] Decide on version number following [Semantic Versioning](https://semver.org/):
  - MAJOR: Breaking changes
  - MINOR: New features (backward compatible)
  - PATCH: Bug fixes (backward compatible)
- [ ] Current version: `git describe --tags --abbrev=0`
- [ ] New version: `v?.?.?`

### Update Files
- [ ] Update PKGBUILD pkgver (if releasing to AUR)
- [ ] Review CHANGELOG or commit history for release notes
- [ ] Update any version-specific documentation

## Release Process

### 1. Ensure Clean Working Directory
```bash
git status  # Should be clean
git pull origin main  # Ensure up to date
```

### 2. Create and Push Git Tag
```bash
# Create annotated tag
git tag -a v1.2.0 -m "Release v1.2.0: Add package manager distribution"

# Push tag to trigger release workflow
git push origin v1.2.0
```

### 3. Monitor Release Workflow
- [ ] Go to: https://github.com/mikesale/dnsres/actions
- [ ] Watch "Release" workflow for the new tag
- [ ] Ensure all jobs complete successfully
- [ ] Check for any errors in logs

### 4. Verify GitHub Release
- [ ] Go to: https://github.com/mikesale/dnsres/releases
- [ ] Verify release was created for the tag
- [ ] Check release assets include:
  - [ ] All platform archives (.tar.gz, .zip)
  - [ ] .deb packages (Linux x86_64, arm64)
  - [ ] .rpm packages (Linux x86_64, arm64)
  - [ ] checksums.txt
- [ ] Review generated changelog
- [ ] Verify release notes are accurate

### 5. Test Installation from Release
- [ ] Test Homebrew (may take a few minutes to update):
  ```bash
  brew install mikesale/dnsres/dnsres
  dnsres -version
  ```
- [ ] Test install script:
  ```bash
  curl -sSL https://raw.githubusercontent.com/mikesale/dnsres/main/install.sh | bash
  ```
- [ ] Download and test manual installation from release page

### 6. Verify Homebrew Formula
GoReleaser should automatically create/update the formula in the main repository:
- [ ] Verify `Formula/dnsres.rb` was created/updated in main repo
- [ ] Check version number matches release
- [ ] Check SHA256 checksums are present
- [ ] Test installation: `brew install mikesale/dnsres/dnsres`

## Post-Release Tasks

### Package Manager Submissions

#### Snap Store (Optional)
If publishing to Snap Store:
- [ ] Build snap locally: `snapcraft`
- [ ] Test snap: `snap install --dangerous dnsres_*.snap`
- [ ] Upload to Snap Store: `snapcraft upload dnsres_*.snap --release=stable`

#### AUR (Arch User Repository)
- [ ] Update PKGBUILD with new version
- [ ] Calculate SHA256 sums:
  ```bash
  wget https://github.com/mikesale/dnsres/releases/download/v1.2.0/dnsres_1.2.0_Linux_x86_64.tar.gz
  sha256sum dnsres_1.2.0_Linux_x86_64.tar.gz
  ```
- [ ] Update sha256sums in PKGBUILD
- [ ] Test PKGBUILD: `makepkg -si`
- [ ] Commit to AUR repository:
  ```bash
  git clone ssh://aur@aur.archlinux.org/dnsres-bin.git
  cd dnsres-bin
  # Update PKGBUILD and .SRCINFO
  makepkg --printsrcinfo > .SRCINFO
  git add PKGBUILD .SRCINFO
  git commit -m "Update to v1.2.0"
  git push
  ```

#### Homebrew Core (Future)
Once project meets requirements:
- [ ] At least 75 GitHub stars OR 30 forks
- [ ] No unresolved bugs
- [ ] Stable for 30+ days
- [ ] Clean audit: `brew audit --strict --online dnsres`
- [ ] Submit PR to homebrew-core

### Announcements
- [ ] Tweet/post on social media (if applicable)
- [ ] Update project website (if applicable)
- [ ] Notify users via Discord/Slack (if applicable)
- [ ] Post in relevant Reddit communities (if applicable)

### Documentation
- [ ] Add release notes to GitHub release (if not auto-generated)
- [ ] Update any external documentation
- [ ] Update package manager documentation if URLs changed

## Troubleshooting

### Release Workflow Failed
1. Check GitHub Actions logs for errors
2. Common issues:
   - GITHUB_TOKEN permissions
   - GoReleaser configuration syntax
   - Build failures for specific platforms
3. Delete tag if needed: `git tag -d v1.2.0 && git push origin :refs/tags/v1.2.0`
4. Fix issue, re-tag, and push again

### Homebrew Formula Not Updated
1. Check for `Formula/dnsres.rb` in the main repository
2. GoReleaser needs GITHUB_TOKEN with repo access
3. Manually create formula if needed:
   ```ruby
   url "https://github.com/mikesale/dnsres/releases/download/v1.2.0/dnsres_1.2.0_Darwin_x86_64.tar.gz"
   sha256 "abc123..."  # Get from checksums.txt
   ```
4. Commit to main branch: `git add Formula/dnsres.rb && git commit -m "Update Homebrew formula"`

### Package Install Issues
1. Test package installation on clean VM
2. Check file permissions in package
3. Verify post-install scripts work correctly
4. Update package configuration if needed

## Version History Template

Keep track of releases:

| Version | Date | Type | Notes |
|---------|------|------|-------|
| v1.2.0 | 2026-02-01 | Minor | Add package manager distribution |
| v1.1.8 | 2026-01-15 | Patch | Bug fixes |
| v1.1.7 | 2026-01-10 | Patch | Performance improvements |

## Emergency Hotfix Process

For critical bugs in production:

1. Create hotfix branch from tag:
   ```bash
   git checkout -b hotfix/v1.2.1 v1.2.0
   ```

2. Fix the bug and commit

3. Tag the hotfix:
   ```bash
   git tag -a v1.2.1 -m "Hotfix: Critical bug fix"
   ```

4. Push tag to trigger release:
   ```bash
   git push origin v1.2.1
   ```

5. Merge hotfix back to main:
   ```bash
   git checkout main
   git merge hotfix/v1.2.1
   git push origin main
   ```

## Rollback Procedure

If a release has critical issues:

1. Mark GitHub release as "Pre-release" to warn users
2. Create hotfix release (see above)
3. Notify users via GitHub issues/discussions
4. Revert Homebrew formula to previous version:
   ```bash
   git revert HEAD -- Formula/dnsres.rb
   git commit -m "Revert Homebrew formula to previous version"
   git push origin main
   ```

---

## Release Checklist Summary

Quick checklist for experienced maintainers:

- [ ] All tests pass
- [ ] Version number decided
- [ ] Tag created and pushed
- [ ] Release workflow succeeded
- [ ] GitHub release verified
- [ ] Homebrew formula updated
- [ ] Installation tested
- [ ] Documentation updated
- [ ] Announcements made (if applicable)

---

**Last Updated:** 2026-02-01  
**Maintainer:** Mike Sale <mike.sale@gmail.com>
