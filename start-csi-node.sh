#!/bin/bash
if ! insmod  /lib/modules/pfkd-$(uname -r).ko ; then
	echo "ERROR: Failed to insmod for this kernel"
fi	

if [ ! -f /sys/pureflash/open ] ; then

  echo "PureFlash pfbd kernel driver 'pfkd.io' not installed!"
  exit 1
fi

echo "Start arguments are: $*"
/opt/pureflash/pureflash_csi $*