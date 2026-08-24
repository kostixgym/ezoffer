#!/usr/bin/env bash
# Разворачивает схему и загружает выгрузки easyoffer.
#
#     import/run.sh [имя_базы]           # по умолчанию ezoffer
#
# Схема пересоздаётся с нуля: скрипт для первичного наполнения, не для
# накатки на боевую базу. Сами CSV лежат в data/ и в git не хранятся.
set -euo pipefail

db="${1:-ezoffer}"
cd "$(dirname "$0")/.."

if [ ! -d data ]; then
	echo "нет каталога data/ с выгрузками" >&2
	exit 1
fi

psql -d postgres -v ON_ERROR_STOP=1 -q \
	-c "DROP DATABASE IF EXISTS \"$db\";" \
	-c "CREATE DATABASE \"$db\";"

psql -d "$db" -v ON_ERROR_STOP=1 -q -f ezoffer.sql
psql -d "$db" -v ON_ERROR_STOP=1 -f import/import.sql
