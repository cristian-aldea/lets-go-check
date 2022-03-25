#!/usr/bin/env bash

if [ -z ${2+x} ]; then
  echo "Usage: deploy.sh SOURCE DESTINATION"
fi

export LGC_SRC=$1
export LGC_DEST=$2

go build "$LGC_SRC"

scp "$LGC_SRC/lets-go-check" "$LGC_DEST/lets-go-check"
scp "$LGC_SRC/checks.json" "$LGC_DEST/checks.json"