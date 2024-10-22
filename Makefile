DATE := $(shell powershell -Command "Get-Date -Format 'yyyyMMdd'")
FUNCTION_NAME := $(func)

benchConv:
	go test -bench=$(FUNCTION_NAME) -benchmem > ./benchmarks/convert_$(DATE).txt
