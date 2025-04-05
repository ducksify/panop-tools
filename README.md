```                                              __                .__
___________    ____   ____ ______           _/  |_  ____   ____ |  |   ______
\____ \__  \  /    \ /  _ \\____ \   ______ \   __\/  _ \ /  _ \|  |  /  ___/
|  |_> > __ \|   |  (  <_> )  |_> > /_____/  |  | (  <_> |  <_> )  |__\___ \
|   __(____  /___|  /\____/|   __/           |__|  \____/ \____/|____/____  >
|__|       \/     \/       |__|                                           \/
```
Simple repo to put panop tools

## new tools
create directory with main.go  
add it in .gorelease.yaml

## Usage
.github/workflows

```
...
  - name: Download release
    uses: robinraju/release-downloader@v1
    with:
      repository: 'ducksify/panop-tools'
      latest: true
      fileName: '*linux_amd64.tar.gz'
      token: ${{ secrets.ACTIONS_TOKEN }}
      extract: 'true'
      out-file-path: ./tools
```
Dockerfile
```
FROM golang:1.24-bullseye AS build
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app

COPY . ./
RUN echo "Current directory:" && pwd \
 && echo "Listing /app/tools:" && ls -lah /app/tools

#13 [build 4/5] RUN echo "Current directory:" && pwd  && echo "Listing /app/tools:" && ls -lah /app/tools
#13 0.140 Current directory:
#13 0.140 /app
#13 0.140 Listing /app/tools:
#13 0.143 total 6.8M
13 0.143 drwxr-xr-x 2 root root 4.0K Apr  5 07:03 .
#13 0.143 drwxr-xr-x 1 root root 4.0K Apr  5 07:03 ..
#13 0.143 -rw-r--r-- 1 root root   73 Apr  4 13:07 README.md
#13 0.143 -rwxr-xr-x 1 root root 3.3M Apr  4 13:08 isapex
#13 0.143 -rw-r--r-- 1 root root 2.1M Apr  5 07:03 panop-tools_0.0.2_linux_amd64.tar.gz
#13 0.143 -rwxr-xr-x 1 root root 1.4M Apr  4 13:07 test
#13 DONE 0.1s

RUN go build -ldflags "-s -w" -o /server

#######################################################

FROM debian:12-slim
WORKDIR /
COPY --from=build /server /server
COPY --from=build /app/tools/isapex /isapex
CMD ["/server"]
```

