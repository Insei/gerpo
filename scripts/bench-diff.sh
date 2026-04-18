#!/usr/bin/env bash
#
# Прогоняет mock-бенчмарки на head PR и на target-ветке, делает статистическое
# сравнение через benchstat и постит сводку комментарием в сам Pull Request.
#
# Используется из .github/workflows/bench-diff.yml в событиях pull_request.
# Ожидает стандартные переменные GitHub Actions:
#   GITHUB_BASE_REF     — имя target-ветки (например, main)
#   GITHUB_REPOSITORY   — owner/repo
#   GITHUB_EVENT_PATH   — путь к JSON-пейлоаду события (для номера PR)
#   GITHUB_API_URL      — обычно https://api.github.com
#   GITHUB_TOKEN        — auto-inject, нужны permissions: pull-requests: write
# Если GITHUB_TOKEN пустой — отчёт остаётся только в логах и артефактах.

set -euo pipefail

COUNT="${BENCH_COUNT:-10}"
BENCH_PATTERN='^Benchmark(GetFirst|GetList|Count|Insert|Update|Delete)_(Direct|Gerpo)$'
BENCH_PKG='./tests/...'

: "${GITHUB_BASE_REF:?must run from pull_request event}"
: "${GITHUB_REPOSITORY:?}"
: "${GITHUB_EVENT_PATH:?}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing tool: $1" >&2; exit 1; }; }
need go
need git
need jq
need curl

PR_NUMBER="$(jq -r '.pull_request.number // empty' "$GITHUB_EVENT_PATH")"
if [[ -z "$PR_NUMBER" ]]; then
  echo "could not read pull_request.number from $GITHUB_EVENT_PATH" >&2
  exit 1
fi

API_URL="${GITHUB_API_URL:-https://api.github.com}"

HEAD_SHA="$(git rev-parse HEAD)"
HEAD_SHORT="$(git rev-parse --short HEAD)"

echo "=== Benchmarking head ($HEAD_SHORT) ==="
go test -run='^$' -bench="$BENCH_PATTERN" -benchmem -count="$COUNT" "$BENCH_PKG" | tee head.txt

echo
echo "=== Fetching base: $GITHUB_BASE_REF ==="
git fetch --depth=50 origin "$GITHUB_BASE_REF"
BASE_REF="origin/$GITHUB_BASE_REF"
BASE_SHORT="$(git rev-parse --short "$BASE_REF")"

git checkout --detach "$BASE_REF"

echo "=== Benchmarking base ($BASE_SHORT) ==="
BASE_OK=1
if ! go test -run='^$' -bench="$BENCH_PATTERN" -benchmem -count="$COUNT" "$BENCH_PKG" | tee base.txt; then
  echo "base bench run failed; diff will be skipped" >&2
  BASE_OK=0
fi

git checkout --detach "$HEAD_SHA"

echo
echo "=== benchstat ==="
go install golang.org/x/perf/cmd/benchstat@latest
BENCHSTAT="$(go env GOPATH)/bin/benchstat"

if [[ "$BASE_OK" == "1" ]]; then
  SUMMARY="$("$BENCHSTAT" base.txt head.txt)"
  HEADER="Benchmark diff: \`$BASE_SHORT\` → \`$HEAD_SHORT\` (mockdb, $COUNT runs)"
else
  SUMMARY="$(cat head.txt)"
  HEADER="Base bench did not run on \`$BASE_SHORT\`; head-only results for \`$HEAD_SHORT\`"
fi

echo "$SUMMARY"

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN is not set — skipping PR comment (report is in job logs)." >&2
  exit 0
fi

BODY=$(printf "### %s\n\n\`\`\`\n%s\n\`\`\`\n" "$HEADER" "$SUMMARY")
PAYLOAD=$(jq -n --arg body "$BODY" '{body: $body}')

curl --fail-with-body -sS \
  -X POST \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  --data "$PAYLOAD" \
  "$API_URL/repos/$GITHUB_REPOSITORY/issues/$PR_NUMBER/comments" \
  >/dev/null

echo "Posted benchmark diff to PR #$PR_NUMBER"
