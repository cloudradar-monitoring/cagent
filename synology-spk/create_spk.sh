#!/bin/sh

if [ -z "$1" ]
  then
    echo "Usage: $0 VERSION"
    exit
fi

# inject version number in package info
sed -i.bak "s/{PKG_VERSION}/$1/g" 2_create_project/INFO
rm 2_create_project/INFO.bak

# ARMv7
GOOS=linux GOARCH=arm GOARM=7 go build github.com/cloudradar-monitoring/cagent/cmd/cagent/...
mv -f cagent 1_create_package/cagent

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../cagent-armv7.spk
rm -f package.tgz
cd ..

# ARMv8
GOOS=linux GOARCH=arm64 go build github.com/cloudradar-monitoring/cagent/cmd/cagent/...
mv -f cagent 1_create_package/cagent

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../cagent-armv8.spk
rm -f package.tgz
cd ..

# AMD64
GOOS=linux GOARCH=amd64 go build github.com/cloudradar-monitoring/cagent/cmd/cagent/...
mv -f cagent 1_create_package/cagent

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../cagent-amd64.spk
rm -f package.tgz
cd ..

# restore local modifications
git checkout 2_create_project/INFO
