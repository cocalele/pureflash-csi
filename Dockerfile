
FROM docker.io/pureflash/pureflash-k8s:1.8.3
LABEL maintainers="LiuLele"
LABEL description="PureFlash pfbd CSI Driver"


COPY bin/pureflash_csi /opt/pureflash/pureflash_csi
COPY start-csi-node.sh /opt/pureflash/start-csi-node.sh
COPY pfkd_helper /usr/bin/pfkd_helper
#COPY modules /lib/modules
USER root
ENTRYPOINT ["/opt/pureflash/start-csi-node.sh"]
