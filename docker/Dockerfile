FROM debian:bullseye

WORKDIR /root

# RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list

ADD openvrr /usr/bin
ADD vrrcli /usr/bin

RUN apt-get update -y
RUN apt install -y net-tools tcpdump iperf3 iputils-ping sudo
RUN apt install -y frr openvswitch-switch

RUN mkdir -p /opt/openvrr

COPY vrr.sh frr.sh \
  ovsdb-server.sh ovs-vswitchd.sh \
  /opt/openvrr/
RUN chmod +x /opt/openvrr/*.sh

LABEL application="Luscis Network Solutions"
LABEL maintainer="danieldin95@163.com"
LABEL org.opencontainers.image.source=https://github.com/luscis/openvrr

CMD ["/opt/openvrr/vrr.sh"]
