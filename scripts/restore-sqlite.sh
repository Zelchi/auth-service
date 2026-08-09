#!/usr/bin/env sh
set -eu

if [ "$#" -ne 2 ] || [ "${CONFIRM_RESTORE:-}" != "YES" ]; then
	echo "uso: CONFIRM_RESTORE=YES $0 BACKUP_PATH DB_PATH" >&2
	exit 2
fi

backup_path=$1
db_path=$2

if [ ! -f "$backup_path" ]; then
	echo "backup não encontrado: $backup_path" >&2
	exit 1
fi

case "$backup_path:$db_path" in
	*"'"*)
		echo "os caminhos não podem conter aspas simples" >&2
		exit 2
		;;
esac

sqlite3 "$db_path" ".timeout 5000" ".restore '$backup_path'"
echo "banco restaurado em $db_path"
