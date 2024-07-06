include .env

TMP=tmp
CREATE_TABLE=sql/schema.sql
DROP_TABLE=sql/drop.sql

run:
	air

tmp:
	@echo "removing $(TMP) file..."
	rmdir /S /Q $(TMP)

conn:
	chcp 1252
	psql -d $(DATABASE_URL)

create:
	chcp 1252
	psql -d $(DATABASE_URL) -f $(CREATE_TABLE)

drop:
	chcp 1252
	psql -d $(DATABASE_URL) -f $(DROP_TABLE)

.SILENT:
.PHONY: run tmp conn drop