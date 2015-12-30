FROM google/debian:wheezy

COPY /build/cron-admin /usr/bin/cron-admin

CMD ["cron-admin"]
