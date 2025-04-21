#!/bin/sh
set -e

API_DIR="api"
OUTPUT_HTTP_API_DIR="internal/handler/http/api"
OUTPUT_GRPC_DIR="pkg/grpc"
HTTP_API_PKG_NAME="api"
# Путь к стандартным proto файлам (если protoc их не находит сам)
#PROTOC_INCLUDE_DIR="D:/Proto/include" # Пример для Windows
#PROTOC_INCLUDE_DIR="/path/to/your/protoc/include" # Пример для Linux/macOS

echo "--- Running OpenAPI Code Generation ---"
echo "Output directory: $OUTPUT_HTTP_API_DIR"
mkdir -p $OUTPUT_HTTP_API_DIR
echo "Generating types..."
oapi-codegen -generate types,skip-prune -package $HTTP_API_PKG_NAME -o $OUTPUT_HTTP_API_DIR/types.gen.go $API_DIR/swagger.yaml
echo "Generating Gin server interface..."
oapi-codegen -generate gin,skip-prune -package $HTTP_API_PKG_NAME -o $OUTPUT_HTTP_API_DIR/server.gen.go $API_DIR/swagger.yaml
echo "OpenAPI code generated successfully."
echo ""

echo "--- Running gRPC Code Generation ---"
echo "Output directory: $OUTPUT_GRPC_DIR"
mkdir -p $OUTPUT_GRPC_DIR
#Генерация Go кода из proto файла
#Используем -I для указания пути к стандартным импортам, если требуется
#protoc --proto_path=$API_DIR -I "$PROTOC_INCLUDE_DIR" --go_out=$OUTPUT_GRPC_DIR --go_opt=paths=source_relative --go-grpc_out=$OUTPUT_GRPC_DIR --go-grpc_opt=paths=source_relative $API_DIR/pvz.proto
echo "Running protoc..."
#Вариант без указания пути к стандартным импортам
protoc --proto_path=$API_DIR --go_out=$OUTPUT_GRPC_DIR --go_opt=paths=source_relative --go-grpc_out=$OUTPUT_GRPC_DIR --go-grpc_opt=paths=source_relative $API_DIR/pvz.proto
echo "gRPC code generated successfully."
echo ""

echo "--- Code Generation Complete ---"