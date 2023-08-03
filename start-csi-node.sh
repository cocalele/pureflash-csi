#!/bin/bash

if [ ! -f /sys/pureflash/open ] ; then
  echo "PureFlash pfbd kernel driver 'pfkd.io' not installed!"
  exit 1
fi

echo "Start arguments are: $*"
/opt/pureflash/pureflash_csi $*