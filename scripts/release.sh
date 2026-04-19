#!/usr/bin/env bash
#
# release.sh — prepare a tagged release locally.
#
# Usage: ./scripts/release.sh vX.Y.Z[-suffix]
#
# Steps the script performs:
#   1. Refuse to run on a branch other than main, on a dirty tree, or with an
#      already-existing tag.
#   2. git pull --ff-only origin main.
#   3. git cliff --tag <TAG> -o CHANGELOG.md  (regenerates the file with the
#      new tag at the top).
#   4. Show the diff so you can eyeball or even edit the file.
#   5. After confirmation, commit CHANGELOG.md and create an annotated tag.
#
# The push step is intentionally NOT here — review your local commit and tag,
# then run `git push --follow-tags origin main` yourself.

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 vX.Y.Z[-suffix]" >&2
  exit 1
fi
TAG="$1"

if ! [[ "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.]+)?$ ]]; then
  echo "Tag must match vX.Y.Z or vX.Y.Z-suffix (e.g. v0.1.0, v1.0.0-rc1). Got: $TAG" >&2
  exit 1
fi

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing tool: $1" >&2; exit 1; }; }
need git
need git-cliff

branch=$(git rev-parse --abbrev-ref HEAD)
if [[ "$branch" != "main" ]]; then
  echo "Must be on main, currently on $branch" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Working tree is dirty — commit or stash before releasing." >&2
  git status --short >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/$TAG" >/dev/null; then
  echo "Tag $TAG already exists locally." >&2
  exit 1
fi

if git ls-remote --tags --exit-code origin "$TAG" >/dev/null 2>&1; then
  echo "Tag $TAG already exists on origin." >&2
  exit 1
fi

echo "▶ Fetching latest main…"
git pull --ff-only origin main

echo "▶ Generating CHANGELOG.md for $TAG…"
git cliff --tag "$TAG" -o CHANGELOG.md

echo
echo "==================== CHANGELOG.md diff ===================="
git --no-pager diff CHANGELOG.md
echo "==========================================================="
echo
echo "If you want to edit the file (typos, regrouping, etc.) — do it now,"
echo "then re-run with the same tag, or proceed and amend later."
echo

read -r -p "Commit CHANGELOG.md and tag $TAG? [y/N] " response
if [[ ! "$response" =~ ^[Yy]$ ]]; then
  echo "Aborted. CHANGELOG.md changes left in the working tree."
  exit 0
fi

git add CHANGELOG.md
git commit -m "docs: prepare changelog for $TAG"
git tag -a "$TAG" -m "$TAG"

echo
echo "✓ Tag $TAG created locally on $(git rev-parse --short HEAD)."
echo "Review with:  git show $TAG"
echo "Push when ready:  git push --follow-tags origin main"
