#!/bin/sh

export NURSECALL_CALL_TOKEN=490ac9e2751428e25e5ef4bb35b8e7cf3774b9afb3a04896ded23736cdc4a593

go run cmd/nursecall/main.go bash -c "sleep 5 && aaa"

