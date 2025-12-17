#!/usr/bin/env zsh
set -euo pipefail

# Bootstrap a Meilisearch index for a project.
# Usage: MEILI_HOST=http://localhost:7700 MEILI_KEY=master ./scripts/meili_bootstrap.sh <project_id>

if [ $# -ne 1 ]; then
  echo "usage: MEILI_HOST=... MEILI_KEY=... $0 <project_id>" >&2
  exit 1
fi

PROJECT_ID="$1"
INDEX="cgap_chunks_${PROJECT_ID}"

if [ -z "${MEILI_HOST:-}" ] || [ -z "${MEILI_KEY:-}" ]; then
  echo "MEILI_HOST and MEILI_KEY must be set" >&2
  exit 1
fi

create_index_body() {
  cat <<EOF
{"uid":"${INDEX}","primaryKey":"id"}
EOF
}

settings_body() {
  cat <<'EOF'
{
  "searchableAttributes": ["title", "text", "section_path"],
  "filterableAttributes": ["project_id", "source_type", "document_uri"],
  "sortableAttributes": ["score_raw", "ord"],
  "rankingRules": [
    "typo",
    "words",
    "proximity",
    "attribute",
    "sort",
    "exactness"
  ]
}
EOF
}

echo "[meili] creating index ${INDEX}"
curl -fsSL -X POST "${MEILI_HOST}/indexes" \
  -H "Authorization: Bearer ${MEILI_KEY}" \
  -H "Content-Type: application/json" \
  -d "$(create_index_body)" >/dev/null

echo "[meili] updating settings"
curl -fsSL -X PATCH "${MEILI_HOST}/indexes/${INDEX}/settings" \
  -H "Authorization: Bearer ${MEILI_KEY}" \
  -H "Content-Type: application/json" \
  -d "$(settings_body)" >/dev/null

echo "[meili] done"
