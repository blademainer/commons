#!/usr/bin/env bash

cur_script_dir="`cd $(dirname $0) && pwd`"
WORK_HOME="${cur_script_dir}/.."

ls ${WORK_HOME}/demos | while read demo; do
    go build -o ${WORK_HOME}/bin/demo/$demo ${WORK_HOME}/demos/$demo
done