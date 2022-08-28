env CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" GOOS="windows" CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" go build  -o ./builds/gampboy-win.exe -x main.go
