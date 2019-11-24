#!/bin/sh

if [ -z "$1" ]
  then
    echo "Usage: $0 VERSION"
    exit
fi

for ARCH in arm_7 arm64 amd64
do
sed -i.bak "s/{PKG_VERSION}/$1/g" 2_create_project/INFO
rm 2_create_project/INFO.bak
sed -i.bak "s/{PKG_ARCH}/noarch/g" 2_create_project/INFO
rm 2_create_project/INFO.bak

cp -f ../dist/cagent_linux_${ARCH}/cagent 1_create_package/cagent
cp -f ../dist/cagent_linux_${ARCH}/cagent 1_create_package/cagent

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../../dist/cagent_$1_synology_${ARCH}.spk
rm -f package.tgz
cd ..
done
