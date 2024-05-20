# Define the name of the output binary
BINARY_NAME := main

# Define the path to the main.go file
MAIN_GO_PATH := roles/analyze_logs/files/main.go

# Build the main.go file and create the binary
build:
	go build -o $(BINARY_NAME) $(MAIN_GO_PATH)
