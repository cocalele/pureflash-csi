FROM docker.io/pureflash/pureflash:1.8.3
LABEL version="1.8.3"

#ARG BUILD_BIN

COPY start-db.sh /opt/pureflash
COPY start-pfs.sh /opt/pureflash
COPY start-pfc.sh /opt/pureflash
COPY pf.conf /etc/pureflash
COPY pfc.conf /etc/pureflash
COPY pfs.conf /etc/pureflash

#ENTRYPOINT 
CMD [ "/opt/pureflash/run-all.sh" ]

