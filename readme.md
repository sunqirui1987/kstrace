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

-->    darwin/amd64: github.com/suiqirui1987/kstrace/docker/cmd
docker build --rm=true --no-cache=true --tag=registry.cn-hangzhou.aliyuncs.com/test_dev/sqr:v1.2 -f Dockerfile .
Sending build context to Docker daemon   9.56MB
....
v1.2: digest: sha256:ad3f64a815726686d106a604bd6eb59fb0fd7f8545ef07eb09d0879d5b6b18c3 size: 1575
```

###  2. KTrace工具构建
```shell
sqrdeMacBook-Pro:k8s qiruisun$ cd kstrace/
sqrdeMacBook-Pro:kstrace qiruisun$ make
go build  -o _output/bin/kstrace ./cmd/kstrace
sqrdeMacBook-Pro:kstrace qiruisun$ ./_output/bin/kstrace  -p xxx-8695888ff8-2xqhn -n default -f open.bt
INFO[0000] running in verbose mode                      
kstrace 849a92f2-d176-11e9-950e-a45e60bba5e1 created
2769   kubelet           380   0 /sys/fs/cgroup/systemd/kubepods/burstable
2769   kubelet           380   0 /sys/fs/cgroup/systemd/kubepods
2769   kubelet           380   0 /sys/fs/cgroup/freezer/kubepods/besteffort
1661   dockerd           151   0 /var/lib/docker/image/overlay2/imagedb/content/sha256/a830b22be
2769   kubelet           380   0 /sys/fs/cgroup/freezer/kubepods/burstable
2769   kubelet           380   0 /sys/fs/cgroup/freezer/kubepods
1661   dockerd            -1   2 /var/lib/docker/image/overlay2/imagedb/metadata/sha256/a830b22b
2769   kubelet           380   0 /sys/fs/cgroup/blkio/kubepods/besteffort
1661   dockerd           151   0 /var/lib/docker/image/overlay2/imagedb/content/sha256/cdc6740b6
2769   kubelet           381   0 /sys/fs/cgroup/cpu,cpuacct/system.slice/wpa_supplicant.service/
2769   kubelet           381   0 /sys/fs/cgroup/blkio/system.slice/wpa_supplicant.service/blkio.
2769   kubelet           381   0 /sys/fs/cgroup/blkio/system.slice/wpa_supplicant.service/blkio.
2769   kubelet           380   0 /sys/fs/cgroup/blkio/system.slice/wpa_supplicant.service/blkio.
2769   kubelet           381   0 /sys/fs/cgroup/blkio/kubepods/burstable
2769   kubelet           380   0 /sys/fs/cgroup/blkio/system.slice/wpa_supplicant.service/blkio.
....
```



