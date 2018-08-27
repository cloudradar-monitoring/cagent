CD /D "%~dp0"
FOR /F "tokens=* USEBACKQ" %%F IN (`git describe --always --tag`) DO (
SET cagent_version=%%F
)
COPY dist\windows_386\cagent.exe cagent.exe
go-msi.exe make --src pkg-scripts\msi-templates --msi dist/cagent_%cagent_version%_Windows_i386.msi --version %cagent_version% --arch 386
DEL cagent.exe
COPY dist\windows_amd64\cagent.exe cagent.exe
go-msi.exe make --src pkg-scripts\msi-templates --msi dist/cagent_%cagent_version%_Windows_x86_64.msi --version %cagent_version% --arch amd64
DEL cagent.exe