SET build_dir=C:\Users\hero\cagent_ci\build_msi\%1
SET cagent_version=%2
SET cagent_version_normalised=%cagent_version:-rc=.0.%
IF %cagent_version_normalised% == %cagent_version% SET cagent_version_normalised=%cagent_version_normalised%.1
SET cert_pass=%3

SET PATH=%PATH%;C:\Program Files (x86)\WiX Toolset v3.11\bin;c:\Program Files (x86)\Windows Kits\10\bin\10.0.17134.0\x86;C:\Program Files\go-msi
CD %build_dir%


COPY dist\cagent_386.exe cagent.exe
go-msi make --src pkg-scripts\msi-templates --msi dist/cagent_32.msi --version %cagent_version_normalised% --arch 386
DEL cagent.exe

COPY dist\cagent_64.exe cagent.exe
go-msi make --src pkg-scripts\msi-templates --msi dist/cagent_64.msi --version %cagent_version_normalised% --arch amd64
DEL cagent.exe

signtool sign /t http://timestamp.digicert.com /f "C:\Users\hero\cagent_ci/build_msi/cloudradar.io.p12" /p %cert_pass% dist/cagent_32.msi
signtool sign /t http://timestamp.digicert.com /f "C:\Users\hero\cagent_ci/build_msi/cloudradar.io.p12" /p %cert_pass% dist/cagent_64.msi
