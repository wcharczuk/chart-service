NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
BLUE_COLOR=\033[94;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

LOG_EVENTS ?= info,warning,error,debug,db.query,db.execute,web.request,web.request.body,request

test:
	@echo "$(OK_COLOR)==> Testing$(NO_COLOR)"
	@DATABASE_URL=postgres://localhost/stocks?sslmode=disable go test ./server/...
	@echo "$(OK_COLOR)==> Testing Complete!$(NO_COLOR)"

run:
	@echo "$(OK_COLOR)==> Running$(NO_COLOR)"
	@LOG_EVENTS=${LOG_EVENTS} DATABASE_URL=postgres://localhost/stocks?sslmode=disable go run main.go

db:
	@echo "$(OK_COLOR)==> Wiping DB$(NO_COLOR)"
	@dropdb stocks;
	@echo "DROP DATABASE"
	@createdb stocks;
	@echo "CREATE DATABASE"
	@pg_restore --dbname=stocks ./_data/stocks.db
	@echo "RESTORE DATABASE"	
	@echo "$(OK_COLOR)==> Wiping DB Done!$(NO_COLOR)"
