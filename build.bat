go build -o build/tracker.exe
echo Built tracker binary!
mkdir build\res
xcopy res build\res /S /Y
echo Copied res directory!