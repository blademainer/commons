#!/usr/bin/env bash

set -x

dir="${1}"
if [[ -z "$dir" ]] || [[ ! -d "$dir" ]]; then
  echo "Not presents dir!"
  exit 1
fi

cur_script_dir="`cd $(dirname $0) && pwd`"
WORK_HOME="${cur_script_dir}/.."
IMPORT_HOME="${WORK_HOME}/../../../"

echo "IMPORT_HOME: `ls $IMPORT_HOME`"

echo "dir: $dir"
docker run --rm -v ${WORK_HOME}/$dir:/defs -v ${IMPORT_HOME}:/input blademainer/protoc-all:1.23_v0.0.3 -i /defs -i /go/src -i /input -d /defs/ -l go -o /defs --validate-out "lang=go:/defs" --with-docs --with-gateway --lint $addition;
