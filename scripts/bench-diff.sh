#!/usr/bin/env bash
#
# Прогоняет mock-бенчмарки на текущей ревизии (head MR) и на целевой ветке
# MR (обычно main), делает статистическое сравнение через benchstat и постит
# сводку комментарием в сам Merge Request.
#
# Используется только из .gitlab-ci.yml в merge_request_event пайплайнах.
# Требует переменных окружения, выставляемых GitLab:
#   CI_MERGE_REQUEST_TARGET_BRANCH_NAME — имя target ветки
#   CI_MERGE_REQUEST_IID                — номер MR
#   CI_PROJECT_ID                       — числовой id проекта
#   CI_API_V4_URL                       — base URL API
# И одной проектной переменной (Settings → CI/CD → Variables):
#   BENCH_BOT_TOKEN                     — personal/project access token c
#                                         правом писать комментарии в MR
# Если токен не задан, скрипт отработает, но комментарий не отправит —
# отчёт останется только в логах job.

set -euo pipefail

COUNT="${BENCH_COUNT:-10}"
BENCH_PATTERN='^Benchmark(GetFirst|GetList|Count|Insert|Update|Delete)_(Direct|Gerpo)$'
BENCH_PKG='./tests/...'

: "${CI_MERGE_REQUEST_TARGET_BRANCH_NAME:?must run from merge_request_event pipeline}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing tool: $1" >&2; exit 1; }; }
need go
need git
need jq
need curl

HEAD_SHA="$(git rev-parse HEAD)"
HEAD_SHORT="$(git rev-parse --short HEAD)"

echo "=== Benchmarking head ($HEAD_SHORT) ==="
go test -run='^$' -bench="$BENCH_PATTERN" -benchmem -count="$COUNT" "$BENCH_PKG" | tee head.txt

echo
echo "=== Fetching base: $CI_MERGE_REQUEST_TARGET_BRANCH_NAME ==="
git fetch --depth=50 origin "$CI_MERGE_REQUEST_TARGET_BRANCH_NAME"
BASE_REF="origin/$CI_MERGE_REQUEST_TARGET_BRANCH_NAME"
BASE_SHORT="$(git rev-parse --short "$BASE_REF")"

git checkout --detach "$BASE_REF"

echo "=== Benchmarking base ($BASE_SHORT) ==="
BASE_OK=1
if ! go test -run='^$' -bench="$BENCH_PATTERN" -benchmem -count="$COUNT" "$BENCH_PKG" | tee base.txt; then
  echo "base bench run failed; diff will be skipped" >&2
  BASE_OK=0
fi

# Возвращаемся на исходную ревизию, чтобы артефакты / последующие джобы не путались.
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

if [[ -z "${BENCH_BOT_TOKEN:-}" ]]; then
  echo "BENCH_BOT_TOKEN is not set — skipping MR comment (report is in job logs)." >&2
  exit 0
fi

BODY=$(printf "### %s\n\n\`\`\`\n%s\n\`\`\`\n" "$HEADER" "$SUMMARY")
PAYLOAD=$(jq -n --arg body "$BODY" '{body: $body}')

curl --fail-with-body -sS \
  -X POST \
  -H "PRIVATE-TOKEN: $BENCH_BOT_TOKEN" \
  -H "Content-Type: application/json" \
  --data "$PAYLOAD" \
  "$CI_API_V4_URL/projects/$CI_PROJECT_ID/merge_requests/$CI_MERGE_REQUEST_IID/notes" \
  >/dev/null

echo "Posted benchmark diff to MR !$CI_MERGE_REQUEST_IID"
