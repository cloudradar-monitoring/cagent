module github.com/cloudradar-monitoring/cagent

go 1.16

replace github.com/kardianos/service => github.com/cloudradar-monitoring/service v1.0.1-0.20190622144052-5da1f538b7fe

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d
	github.com/cloudradar-monitoring/dmidecode v0.0.0-20190211163023-395107264116
	github.com/cloudradar-monitoring/selfupdate v0.0.0-20200615195818-3bc6d247a637
	github.com/davecgh/go-spew v1.1.1
	github.com/gentlemanautomaton/windevice v0.0.0-20190308095644-de21ffdab1a3
	github.com/gentlemanautomaton/winguid v0.0.0-20190307223039-3f364f74ee74 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-ole/go-ole v1.2.4
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jaypipes/ghw v0.7.0
	github.com/kardianos/service v1.0.1-0.20190622144052-5da1f538b7fe
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lxn/walk v0.0.0-20190515104301-6cf0bf1359a5
	github.com/lxn/win v0.0.0-20190514122436-6f00d814e89c
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/nightlyone/lockfile v1.0.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v2.18.13-0.20190131151121-071446942108+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.2.2
	github.com/troian/toml v0.4.2
	github.com/vcraescu/go-xrandr v0.0.0-20190102070802-135ba5f1bc04
	golang.org/x/sys v0.0.0-20191024073052-e66fe6eb8e0c
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	gopkg.in/toast.v1 v1.0.0-20180812000517-0a84660828b2
	gopkg.in/yaml.v2 v2.2.2 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb
)
