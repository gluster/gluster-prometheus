#!/bin/bash

OUTDIR=${1:-build}
mkdir -p "$OUTDIR"

OUTPUT=$OUTDIR/gluster-exporter.service

cat >"$OUTPUT" <<EOF
[Unit]
Description=Gluster Prometheus Exporter

[Service]
ExecStart=${SBINDIR}/gluster-exporter --config=${SYSCONFDIR}/gluster-exporter/global.conf --collectors-config=${SYSCONFDIR}/gluster-exporter/collectors.conf
KillMode=process

[Install]
WantedBy=multi-user.target

EOF
