# Installing **Smartmontools**

**cagent** requires **smartmontools** of version 7.0 and higher

## Windows
**smartctl** binary from **smartmontools** is shipped with **cagent** msi installer

## MacOS
### Using Homebrew
```bash
brew install smartmontools
```

### Using DMG installer
[Download and install](https://sourceforge.net/projects/smartmontools/files/smartmontools/7.0/smartmontools-7.0-1.dmg/download) 

## Linux
Linux distributions may provide outdated installation of **smartmontools**

### Check for available version
#### apt-get
```bash
apt-get update && apt-cache show smartmontools | grep Version
```
#### yum
```bash
yum info smartmontools | grep Version
```

If version provided by packaging system less then **7.0** then follow to [Install](#Install) section
 
### Install
Current version **7.0**
```bash
wget --content-disposition https://sourceforge.net/projects/smartmontools/files/smartmontools/7.0/smartmontools-7.0.tar.gz/download
tar -xvzf smartmontools-7.0.tar.gz
cd smartmontools-7.0
./configure
make
sudo make install

```
