#!/bin/bash

OUTDIR=${1:-build}
mkdir -p "$OUTDIR"

OUTPUT=$OUTDIR/gluster-exporter.service

ARGS=""
if [ "$GLUSTER_MGMT" == "glusterd2" ];then
    ARGS+="--gluster-mgmt=glusterd2"
fi

cat >"$OUTPUT" <<EOF
[Unit]
Description=Gluster Prometheus Exporter

[Service]
ExecStart=${SBINDIR}/gluster-exporter ${ARGS}
KillMode=process

[Install]
WantedBy=multi-user.target

EOF
