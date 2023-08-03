#!/bin/bash
set -m

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

JAVA_HOME=/usr/lib/jvm/jdk-15/
export PATH=/opt/pureflash:$JAVA_HOME/bin:$PATH

if grep __NODE_ID /etc/pureflash/pfs.conf ; then
    echo "This is first time starting, create initial config ..."

    #sleep a randon time to avoid competion with other pfs instance
    sleep $(shuf -i 0-10 -n 1)
    PFS_CNT=$(/opt/pureflash/apache-zookeeper-3.5.9-bin/bin/zkCli.sh -server pfszk-cs.pureflash.svc.cluster.local ls /pureflash/cluster1/stores  | tail -1 | awk -F, '{print NF}' )
    PFS_ID=$((PFS_CNT+1))
    while ! /opt/pureflash/apache-zookeeper-3.5.9-bin/bin/zkCli.sh -server pfszk-cs.pureflash.svc.cluster.local create  /pureflash/cluster1/stores/$PFS_ID ; do
      echo "store ID $PFS_ID already been used"
      PFS_ID=$((PFS_ID+1))
    done
    sed -i "s/__NODE_ID__/$PFS_ID/g" /etc/pureflash/pfs.conf

fi

if [ "$PFS_DISKS" != "" ]; then
	echo "Use disk $PFS_DISKS specified from environment variable PFS_DISKS";
else
	echo "Use data file /opt/pureflash/disk1.dat as disk, only for testing"
	if [ ! -f /opt/pureflash/disk1.dat ]; then
	  echo "Create disk file ..."
	  truncate -s 20G /opt/pureflash/disk1.dat
	fi
	export PFS_DISKS="/opt/pureflash/disk1.dat"
fi
	
i=0
for d in ${PFS_DISKS//,/ }; do
	sed -i "/__TRAY_PLACEHOLDER__/i [tray.$i]\n\tdev = $d" /etc/pureflash/pfs.conf
	i=$((i+1))
done
sed -i "/__TRAY_PLACEHOLDER__/d" /etc/pureflash/pfs.conf
sed -i "s/__NODE_IP__/$MY_NODE_IP/g" /etc/pureflash/pfs.conf



OLD_PID=$(pidof pfs)
if [ "$OLD_PID" != "" ]; then
	echo "OLD_PID:$OLD_PID "
	kill -2 $OLD_PID
fi

echo "Start PureFlash store..."
$DIR/pfs -c /etc/pureflash/pfs.conf 
