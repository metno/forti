#/bin/bash

set -euo pipefail

RUN="go run ../../upload/cmd/mktestdata/main.go"
WORKDIR="./data"


mkdir -p "$WORKDIR"

$RUN 	-workdir $WORKDIR \
	-area group_a \
	-version 1 \
	-grid grid_a \
	-parameters "p1=2,p2=0" \
        -lat 5,5,6,6 \
        -lon 1,2,1,2

$RUN 	-workdir $WORKDIR \
        -area group_b \
        -version 2 \
        -grid grid_a \
	 -parameters "foo=2,bar=1"

$RUN 	-workdir $WORKDIR \
        -area group_b \
        -version 2 \
        -grid grid_b \
        -parameters "bik=4,bok=0"

