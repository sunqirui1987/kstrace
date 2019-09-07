IMAGE_VERSION=v1.2
IMAGE_NAME=registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:$(IMAGE_VERSION)


.PHONY: build
build: 
	gox -os="linux" -arch="amd64" -output="_output/bin/kstrace_exec_linux" ./cmd
	gox -os="darwin" -arch="amd64" -output="_output/bin/kstrace_exec_darwin" ./cmd
	docker build --rm=true --no-cache=true --tag=$(IMAGE_NAME) -f Dockerfile .
	docker push $(IMAGE_NAME)

.PHONY: clean
clean:
	rm -Rf _output
