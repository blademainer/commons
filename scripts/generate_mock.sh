#!/usr/bin/env bash

files="
./pkg/rpc/queue/queue.go
./pkg/sql/flatten.go
"

cur_script_dir="$(cd $(dirname $0) && pwd)"
WORK_HOME="${cur_script_dir}/.."

echo "${files}" | while read file; do
  [[ -z "$file" ]] && continue
  echo "generate file: ${file}"
  mockgen -source "${WORK_HOME}/${file}" -destination "${WORK_HOME}/${file%/*}/mock/mock.go"
done

