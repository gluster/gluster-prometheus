#!/bin/bash

OUTDIR=${1:-build}
mkdir -p "$OUTDIR"

OUTPUT=$OUTDIR/gluster-exporter.service

cat >"$OUTPUT" <<EOF
[Unit]
Description=Gluster Prometheus Exporter

[Service]
ExecStart=${SBINDIR}/gluster-exporter --config=${SYSCONFDIR}/gluster-exporter/gluster-exporter.toml
KillMode=process

[Install]
WantedBy=multi-user.target

EOF
