#!/bin/bash
set -e

go get github.com/fernet/fernet-go/fernet-sign
go get github.com/kr/webx/...

export FERNET_KEY=`dd if=/dev/urandom bs=1 count=32 2>/dev/null|openssl base64`
export TOK=`printf foo|fernet-sign FERNET_KEY`
export REQADDR=:8000
export REQTLSADDR=:4443
export BKDADDR=:1111
export PORT=5000
export WEBX_URL="https://foo:$TOK@127.0.0.1:1111/"
export WEBX_VERBOSE=1

urouter &
webxd &
"$@" &

trap 'kill -9 %1 %2 %3' EXIT
wait
