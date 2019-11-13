#!/usr/bin/env bash

cur_script_dir="`cd $(dirname $0) && pwd`"
WORK_HOME="${cur_script_dir}/.."

#darwin	386
#darwin	amd64
#darwin	arm
#darwin	arm64
#dragonfly	amd64
#freebsd	386
#freebsd	amd64
#freebsd	arm
#linux	386
#linux	amd64
#linux	arm
#linux	arm64
#linux	ppc64
#linux	ppc64le
#netbsd	386
#netbsd	amd64
#netbsd	arm
#openbsd	386
#openbsd	amd64
#openbsd	arm
#plan9	386
#plan9	amd64
#solaris	amd64
#windows	386
#windows	amd64

OS_ARCHS=(
"darwin,386"
"darwin,amd64"
#"darwin,arm"
#"darwin,arm64"
#"dragonfly,amd64"
"freebsd,386"
"freebsd,amd64"
"freebsd,arm"
"linux,386"
"linux,amd64"
"linux,arm"
"linux,arm64"
"linux,ppc64"
"linux,ppc64le"
"netbsd,386"
"netbsd,amd64"
"netbsd,arm"
"openbsd,386"
"openbsd,amd64"
"openbsd,arm"
#"plan9,386"
#"plan9,amd64"
"solaris,amd64"
"windows,386"
"windows,amd64"
)



for os_arch in ${OS_ARCHS[@]}; do
  os=`echo $os_arch | awk -F "," '{print $1}'`
  arch=`echo $os_arch | awk -F "," '{print $2}'`
  GOOS=$os GOARCH=$arch go build -o ${WORK_HOME}/bin/main/main_${os}_${arch} ${WORK_HOME}/cmd/main
done