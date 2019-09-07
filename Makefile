IMAGE_VERSION=v1.1
IMAGE_NAME=registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:$(IMAGE_VERSION)


.PHONY: build
build: 
	go build  -o _output/bin/kstrace ./cmd/kstrace

.PHONY: clean
clean:
	rm -Rf _output
