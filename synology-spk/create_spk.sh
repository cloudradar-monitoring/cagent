#!/bin/sh

if [ -z "$1" ]
  then
    echo "Usage: $0 VERSION"
    exit
fi


# IMPORTANT: CGO_ENABLED=0 is used to force binaries to be statically linked

# ARMv7
sed -i.bak "s/{PKG_VERSION}/$1/g" 2_create_project/INFO
rm 2_create_project/INFO.bak
sed -i.bak "s/{PKG_ARCH}/noarch/g" 2_create_project/INFO
rm 2_create_project/INFO.bak

CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build github.com/cloudradar-monitoring/cagent/cmd/cagent/...
mv -f cagent 1_create_package/cagent

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../cagent-armv7.spk
rm -f package.tgz
cd ..

git checkout 2_create_project/INFO


# ARMv8
sed -i.bak "s/{PKG_VERSION}/$1/g" 2_create_project/INFO
rm 2_create_project/INFO.bak
sed -i.bak "s/{PKG_ARCH}/noarch/g" 2_create_project/INFO
rm 2_create_project/INFO.bak

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build github.com/cloudradar-monitoring/cagent/cmd/cagent/...
mv -f cagent 1_create_package/cagent

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../cagent-armv8.spk
rm -f package.tgz
cd ..

git checkout 2_create_project/INFO


# AMD64
sed -i.bak "s/{PKG_VERSION}/$1/g" 2_create_project/INFO
rm 2_create_project/INFO.bak
sed -i.bak "s/{PKG_ARCH}/x86_64 cedarview bromolow broadwell/g" 2_create_project/INFO
rm 2_create_project/INFO.bak

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build github.com/cloudradar-monitoring/cagent/cmd/cagent/...
mv -f cagent 1_create_package/cagent

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../cagent-amd64.spk
rm -f package.tgz
cd ..

git checkout 2_create_project/INFO
