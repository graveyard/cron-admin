FROM google/debian:wheezy
ADD ./static /root/cron-admin/static
COPY /build/cron-admin /root/cron-admin/cron-admin

EXPOSE 80
WORKDIR /root/cron-admin
CMD ["./cron-admin"]
