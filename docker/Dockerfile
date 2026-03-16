FROM debian:bookworm

WORKDIR /root

# RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list
RUN apt-get update -y
RUN apt install -y net-tools tcpdump iperf3 iputils-ping frr openvswitch-switch

RUN mkdir -p /opt/openvrr

ADD openvrr /usr/bin
ADD openvrr-d /opt/openvrr

COPY script/vrr.sh script/frr.sh \
  script/ovsdb-server.sh script/ovs-vswitchd.sh \
  /opt/openvrr/
RUN chmod +x /opt/openvrr/*.sh

LABEL application="Luscis Network Solutions"
LABEL maintainer="danieldin95@163.com"
LABEL org.opencontainers.image.source=https://github.com/luscis/openvrr

CMD ["/opt/openvrr/vrr.sh"]
