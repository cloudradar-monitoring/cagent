## How to build from sources
- [Install Golang 1.9 or newer](https://golang.org/dl/)
```bash
go get -d -u github.com/cloudradar-monitoring/cagent
go build -o -ldflags="-X main.VERSION=$(git --git-dir=src/github.com/cloudradar-monitoring/cagent/.git describe --always --long --dirty --tag)" cagent github.com/cloudradar-monitoring/cagent/cmd/cagent
```

## Run the example

```bash
./cagent -i src/github.com/cloudradar-monitoring/cagent/example.json -o result.out
```
Use `ctrl-c` to stop it

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
CAGENT_VERSION=$(git describe --always --long --dirty --tag) goreleaser --snapshot
```

## Build MSI package
– Should be done on Windows machine
– Open command prompt(cmd.exe)
– Go to cagent directory `cd path_to_directory`
– Run `goreleaser --snapshot` to build binaries
– Run `build-win.bat`

Current Status
------------------

- x: work
- b: almost works, but something is broken

=================== ====== ======= ======= ====== ======= =======
name                Linux  FreeBSD OpenBSD MacOSX Windows Solaris
cpu_times             x      x       x       x       x
cpu_count             x      x       x       x       x
cpu_percent           x      x       x       x       x
cpu_times_percent     x      x       x       x       x
virtual_memory        x      x       x       x       x       b
swap_memory           x      x       x       x
disk_partitions       x      x       x       x       x
disk_io_counters      x      x       x
disk_usage            x      x       x       x       x
net_io_counters       x      x       x       b       x
boot_time             x      x       x       x       x
users                 x      x       x       x       x
pids                  x      x       x       x       x
pid_exists            x      x       x       x       x
net_connections       x              x       x
net_protocols         x
net_if_addrs
net_if_stats
netfilter_conntrack   x
=================== ====== ======= ======= ====== =======

Process class
^^^^^^^^^^^^^^^

================ ===== ======= ======= ====== =======
name             Linux FreeBSD OpenBSD MacOSX Windows
pid                 x     x      x       x       x
ppid                x     x      x       x       x
name                x     x      x       x       x
cmdline             x     x              x       x
create_time         x                    x
status              x     x      x       x
cwd                 x
exe                 x     x      x               x
uids                x     x      x       x
gids                x     x      x       x
terminal            x     x      x       x
io_counters         x     x      x               x
nice                x     x      x       x       x
num_fds             x
num_ctx_switches    x
num_threads         x     x      x       x       x
cpu_times           x                            x
memory_info         x     x      x       x       x
memory_info_ex      x
memory_maps         x
open_files          x
send_signal         x     x      x       x
suspend             x     x      x       x
resume              x     x      x       x
terminate           x     x      x       x       x
kill                x     x      x       x
username            x     x      x       x       x
ionice
rlimit              x
num_handlers
threads             x
cpu_percent         x            x       x
cpu_affinity
memory_percent
parent              x            x       x       x
children            x     x      x       x       x
connections         x            x       x
is_running
================ ===== ======= ======= ====== =======

Original Metrics
^^^^^^^^^^^^^^^^^^^

================== ===== ======= ======= ====== ======= =======
item               Linux FreeBSD OpenBSD MacOSX Windows Solaris
**HostInfo**
hostname              x     x      x       x       x       x
  uptime              x     x      x       x               x
  proces              x     x      x                       x
  os                  x     x      x       x       x       x
  platform            x     x      x       x               x
  platformfamily      x     x      x       x               x
  virtualization      x
**CPU**
  VendorID            x     x      x       x       x      x
  Family              x     x      x       x       x      x
  Model               x     x      x       x       x      x
  Stepping            x     x      x       x       x      x
  PhysicalID          x                                   x
  CoreID              x                                   x
  Cores               x                            x      x
  ModelName           x     x      x       x       x      x
  Microcode           x                                   x
**LoadAvg**
  Load1               x     x      x       x
  Load5               x     x      x       x
  Load15              x     x      x       x
**GetDockerID**
  container id        x     no     no      no      no
**CgroupsCPU**
  user                x     no     no      no      no
  system              x     no     no      no      no
**CgroupsMem**
  various             x     no     no      no      no
================== ===== ======= ======= ====== ======= =======