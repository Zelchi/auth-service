#!/usr/bin/env sh
set -eu

if command -v openssl >/dev/null 2>&1; then
	openssl rand -hex 48
	exit 0
fi

od -An -N48 -tx1 /dev/urandom | tr -d ' \n'
