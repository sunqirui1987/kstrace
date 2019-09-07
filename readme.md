### KTrace工具

   基于kubernets的追踪工具

```
sqrdeMacBook-Pro:k8s qiruisun$ git clone https://github.com/suiqirui1987/kstrace.git
```

### 1.  DOCKER构建
```shell
sqrdeMacBook-Pro:k8s qiruisun$ cd kstrace/docker/
sqrdeMacBook-Pro:docker qiruisun$ make
gox -os="linux" -arch="amd64" -output="_output/bin/kstrace_exec_linux" ./cmd
Number of parallel builds: 3

-->     linux/amd64: github.com/suiqirui1987/kstrace/docker/cmd
gox -os="darwin" -arch="amd64" -output="_output/bin/kstrace_exec_darwin" ./cmd
Number of parallel builds: 3

-->    darwin/amd64: github.com/suiqirui1987/kstrace/docker/cmd
docker build --rm=true --no-cache=true --tag=registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:v1.2 -f Dockerfile .
Sending build context to Docker daemon   9.56MB
Step 1/5 : FROM registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:v1.0
 ---> b627989ad427
Step 2/5 : MAINTAINER Sqr <sqr@detu.com>
 ---> Running in 591058524a62
Removing intermediate container 591058524a62
 ---> 28125f8d5386
Step 3/5 : RUN ulimit -l 8192
 ---> Running in 134ac55abd27
Removing intermediate container 134ac55abd27
 ---> 81d2547f45ee
Step 4/5 : RUN rm -rf /root/*
 ---> Running in 85701d8200f1
Removing intermediate container 85701d8200f1
 ---> 0a9555cb4cdb
Step 5/5 : COPY ./_output/bin /root/
 ---> ec3db349eb5b
Successfully built ec3db349eb5b
Successfully tagged registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:v1.2
docker push registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:v1.2
The push refers to repository [registry.cn-hangzhou.aliyuncs.com/test_dev/sqr]
d22a80ff17d2: Pushed 
523e931ff3a4: Layer already exists 
9f8be4227b1d: Layer already exists 
9571205333fa: Layer already exists 
70020405e466: Layer already exists 
43316bbb0407: Layer already exists 
v1.2: digest: sha256:ad3f64a815726686d106a604bd6eb59fb0fd7f8545ef07eb09d0879d5b6b18c3 size: 1575
```

###  2. KTrace工具构建
```shell
sqrdeMacBook-Pro:k8s qiruisun$ cd kstrace/
sqrdeMacBook-Pro:kstrace qiruisun$ make
go build  -o _output/bin/kstrace ./cmd/kstrace
sqrdeMacBook-Pro:kstrace qiruisun$ ./_output/bin/kstrace  -p yapi-8695888ff8-2xqhn -n default -f cpuwalk.bt
INFO[0000] running in verbose mode                      
kstrace 849a92f2-d176-11e9-950e-a45e60bba5e1 created
if your program has maps to print, send a SIGINT using Ctrl-C, if you want to interrupt the execution send SIGINT two times
Attaching 2 probes...
Sampling CPU at 99hz... Hit Ctrl-C to end.
^C
first SIGINT received, now if your program had maps and did not free them it should print them out


@cpu: 
[0, 1)                 7 |@@@@@@@@@@@@                                        |
[1, 2)                14 |@@@@@@@@@@@@@@@@@@@@@@@@@                           |
[2, 3)                 5 |@@@@@@@@                                            |
[3, 4)                 5 |@@@@@@@@                                            |
[4, 5)                14 |@@@@@@@@@@@@@@@@@@@@@@@@@                           |
[5, 6)                29 |@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@|
[6, 7)                20 |@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@                 |
```



