#!/usr/bin/env bash
#set -x
set -e
cur_script_dir="`cd $(dirname $0) && pwd`"
WORK_HOME="${cur_script_dir}/.."
IMPORT_HOME="${WORK_HOME}/../../../"
echo "dirname $WORK_HOME"
echo "WORK_HOME = $WORK_HOME"
find $WORK_HOME -name "*.proto" | while read proto; do
  dir="`dirname $proto`"
  echo "dir: `cd $dir && pwd`"
  docker run --rm -v $dir:/defs -v ${IMPORT_HOME}:/input blademainer/protoc-all:1.23_v0.0.3 -i /defs -i /input -d /defs/ -l go -o /defs --validate-out "lang=go:/defs" --with-gateway --lint $addition;
done


# generage js protos
find $WORK_HOME -name "generage.sh" | while read proto; do
#  dir="`dirname $proto`"
#  echo "dir: `cd $dir && pwd`"
#  docker run --rm -v $dir:/defs -v ${IMPORT_HOME}:/input blademainer/protoc-all:1.23_v0.0.3 -i /defs -i /input -d /defs/ -l node -o /defs  --lint $addition;
  sh $proto
done