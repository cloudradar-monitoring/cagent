#!/bin/sh

export PATH=$PATH:"C:\Program Files (x86)\WiX Toolset v3.11\bin"
export PATH=$PATH:"C:\Program Files (x86)\Windows Kits\10\bin\10.0.17134.0\x86"

export GIT_TAG=$(git describe --exact-match --tags $(git log -n1 --pretty='%h'))
echo "Building version $GIT_TAG"

cp dist/windows_386/cagent.exe cagent.exe
go-msi make -k --src pkg-scripts/msi-templates --msi dist/_cagent_32.msi --version $GIT_TAG --arch 386
rm cagent.exe

cp dist/windows_amd64/cagent.exe cagent.exe
go-msi make --src pkg-scripts/msi-templates --msi dist/_cagent_64.msi --version $GIT_TAG --arch amd64
rm cagent.exe

mv dist/_cagent_32.msi dist/cagent_32.msi
mv dist/_cagent_64.msi dist/cagent_64.msi
