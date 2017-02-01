#!/bin/bash
pushd server
go build -v -ldflags "-linkmode external -extldflags -static" -o ~/tmp/tlstun_server

acbuildEnd() {
            export EXIT=$?
                    acbuild --debug end && exit $EXIT
}
trap acbuildEnd EXIT

acbuild begin
acbuild label add arch amd64
acbuild label add os linux
acbuild set-name pronoc.net/tlstun-server
acbuild dependency add quay.io/coreos/alpine-sh
acbuild mount add config /config
acbuild copy ~/tmp/tlstun_server /tlstun_server
acbuild port add tunnel tcp 443
acbuild set-exec -- /bin/sh -c 'cd /config;/tlstun_server $@' --
acbuild write tlstun_server.aci
acbuild end
popd server
