module github.com/humpback/humpback-center

go 1.15

require (
	github.com/boltdb/bolt v1.3.1
	github.com/containerd/containerd v1.3.9 // indirect
	github.com/containerd/fifo v0.0.0-20201026212402-0724c46b320c // indirect
	github.com/containerd/ttrpc v1.0.2 // indirect
	github.com/containerd/typeurl v1.0.1 // indirect
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.4.0
	github.com/docker/libcompose v0.4.1-0.20190808084053-143e0f3f1ab9
	github.com/docker/libkv v0.2.1 // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/gorilla/mux v0.0.0-20160317213430-0eeaf8392f5b
	github.com/hashicorp/consul/api v1.8.0 // indirect
	github.com/humpback/common v0.0.0-20181013082642-ad18fdd2e380
	github.com/humpback/discovery v0.0.0-20181012143229-ebb7ec858a9e
	github.com/humpback/gounits v0.0.0-20190102083213-0662ccb72e5a
	github.com/opencontainers/runtime-spec v1.0.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	gopkg.in/alexcesaro/quotedprintable.v2 v2.0.0-20150314193201-9b4a113f96b3 // indirect
	gopkg.in/eapache/queue.v1 v1.1.0 // indirect
	gopkg.in/gomail.v1 v1.0.0-20150320132819-11b919ab4933
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20201201034508-7d75c1d40d88+incompatible
