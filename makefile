BINARY_NAME=myapp
DSN="host=localhost port=5432 user=postgres password=password dbname=concurrency sslmode=disable"
REDIS="127.0.0.1:6379"

# docker cp ./localfile.sql containername:/container/path/file.sql
# docker exec -u postgresuser containername psql dbname postgresuser -f /container/path/file.sql

build:
	@echo "Building..."
	env CGO_EBABLE=0 go build -ldflags="-s -w" -o ${BINARY_NAME} ./cmd/web
	@echo "Built1..."

run: build
	@echo "Starting..."
	@env DSN=${DSN} REDIS=${REDIS} ./${BINARY_NAME} &
	@echo "Started!..."

clean:
	@echo "Cleaning..."
	@go clean
	@rm ${BINARY_NAME}
	@echo "Cleaned!..."

start: run

stop:
	@echo "Stopping..."
	@-pkill -SIGTERM -f "./${BINARY_NAME}"
	@echo "Stopped!..."

restart: stop start

test:
	go test -v ./...
