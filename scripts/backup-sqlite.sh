#!/usr/bin/env sh
set -eu

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
	echo "uso: $0 DB_PATH [BACKUP_PATH]" >&2
	exit 2
fi

db_path=$1
backup_path=${2:-"${db_path}.backup-$(date -u +%Y%m%dT%H%M%SZ)"}

if [ ! -f "$db_path" ]; then
	echo "banco não encontrado: $db_path" >&2
	exit 1
fi

case "$backup_path" in
	*"'"*)
		echo "BACKUP_PATH não pode conter aspas simples" >&2
		exit 2
		;;
esac

sqlite3 "$db_path" ".timeout 5000" ".backup '$backup_path'"
echo "backup criado em $backup_path"
