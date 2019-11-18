## How to build from sources
- [Install Golang 1.9 or newer](https://golang.org/dl/)
```bash
go get -d -u github.com/cloudradar-monitoring/cagent
go build -o cagent -ldflags="-X github.com/cloudradar-monitoring/cagent.Version=$(git --git-dir=$GOPATH/src/github.com/cloudradar-monitoring/cagent/.git describe --always --long --dirty --tag)" github.com/cloudradar-monitoring/cagent/cmd/cagent
```

## How to run
***-r** for _one run only_ mode*
```bash
./cagent -r -o result.out
```

## Configuration
Check the [example config](https://github.com/cloudradar-monitoring/cagent/blob/master/example.config.toml)

Default locations:
* Mac OS: `~/.cagent/cagent.conf`
* Windows: `./cagent.conf`
* UNIX: `/etc/cagent/cagent.conf`

## Logs location
* Mac OS: `~/.cagent/cagent.log`
* Windows: `./cagent.log`
* UNIX: `/etc/cagent/cagent.conf`

## Build binaries and deb/rpm packages
– Install [goreleaser](https://goreleaser.com/introduction/)
```bash
goreleaser --snapshot
```

## Build MSI package
Should be done on Windows machine
- [Download go-msi](https://github.com/cloudradar-monitoring/go-msi/releases) and put it in the `C:\Program Files\go-msi`
- Open command prompt(cmd.exe or powershell)
- Go to cagent directory `cd path_to_directory`
- Run `goreleaser --snapshot` to build binaries
- Run `build-win.bat`

## S.M.A.R.T monitoring documentation
**S.M.A.R.T** [how to](https://github.com/cloudradar-monitoring/cagent/blob/master/SMART.md)
