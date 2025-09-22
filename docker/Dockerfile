FROM debian:bullseye

WORKDIR /root

# RUN sed -ie 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list

ADD openvrr /usr/bin
ADD vrrcli /usr/bin

RUN apt-get update -y
RUN apt install -y net-tools tcpdump iperf3 iputils-ping sudo
RUN apt install -y frr openvswitch-switch

RUN mkdir -p /opt/openvrr

ADD vrr.sh /opt/openvrr
ADD frr.sh /opt/openvrr

RUN chmod +x /opt/openvrr/vrr.sh
RUN chmod +x /opt/openvrr/frr.sh

LABEL application="OpenLAN Network Solutions"
LABEL maintainer="danieldin95@163.com"

CMD ["/opt/openvrr/vrr.sh"]
