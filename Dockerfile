FROM google/debian:wheezy
ADD ./static /root/cron-admin/static
COPY /bin/cron-admin /root/cron-admin/cron-admin
WORKDIR /root/cron-admin
CMD ["./cron-admin"]
