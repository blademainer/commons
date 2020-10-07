#!/usr/bin/env bash
[[ -n "$(docker ps -a -f name=test-mysql | grep test-mysql)" ]] && echo "mysql is running!" && exit 0

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/.."
sqlDir="${WORK_HOME}/docs/db"
echo "Sql dir: $sqlDir"
echo "Should init sql scripts: $(ls "$sqlDir")"
docker run -d --name test-mysql -v "$sqlDir":/docker-entrypoint-initdb.d -v "$WORK_HOME/mysql/data":/var/lib/mysql -e MYSQL_ROOT_PASSWORD=admin -e MYSQL_USER=test -e MYSQL_PASSWORD=111 -e MYSQL_DATABASE=test -e LANG=C.UTF-8  -p 3306:3306 mysql:latest --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci --transaction_isolation=READ-COMMITTED --default-authentication-plugin=mysql_native_password
