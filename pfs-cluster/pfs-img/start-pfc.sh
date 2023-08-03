#!/bin/bash
set -m

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

JAVA_HOME=/opt/pureflash/jdk-17.0.6
export PATH=/opt/pureflash:$JAVA_HOME/bin:$PATH

OLD_PID=$(ps -f |grep jconductor |grep java|awk '{print $2}')
if [ "$OLD_PID" != "" ]; then
	echo "OLD_PID:$OLD_PID "
	kill -2 $OLD_PID
fi


echo "Set mngt_ip=$MY_NODE_IP"
sed -i "s/__NODE_IP__/$MY_NODE_IP/" /etc/pureflash/pfc.conf

echo "Start PureFlash jconductor..."
JCROOT=$DIR/jconductor
/usr/bin/java  -classpath $JCROOT/pfconductor.jar:$JCROOT/lib/*  \
   -Dorg.slf4j.simpleLogger.showDateTime=true \
   -Dorg.slf4j.simpleLogger.dateTimeFormat="[yyyy/MM/dd H:mm:ss.SSS]" \
   -XX:+HeapDumpOnOutOfMemoryError \
   -Xmx2G \
   com.netbric.s5.conductor.Main -c /etc/pureflash/pfc.conf 



