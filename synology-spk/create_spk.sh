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
cp -f ../dist/csender_linux_${ARCH}/csender 1_create_package/csender

cd 1_create_package
tar cvfz package.tgz *
mv package.tgz ../2_create_project/
cd ../2_create_project/
tar cvfz cagent.spk *
mv cagent.spk ../../dist/cagent_$1_synology_${ARCH}.spk
rm -f package.tgz
cd ..
done

## special case to normalize armv7 build name to be in line with gorelease files
mv ../dist/cagent_$1_synology_arm_7.spk ../dist/cagent_$1_synology_armv7.spk
