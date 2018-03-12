FROM scratch

EXPOSE 8080

COPY nasne-exporter /opt/bin/


ENTRYPOINT ["/opt/bin/nasne-exporter"]
