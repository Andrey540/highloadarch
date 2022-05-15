FROM centos:centos7

RUN yum -y install git ncurses-devel cmake make gcc-c++ boost boost-devel wget unzip nano bzip2 mysql-devel mysql-lib \
    && git clone https://github.com/tarantool/mysql-tarantool-replication.git \
    && cd mysql-tarantool-replication \
    && git submodule update --init --recursive \
    && cmake . \
    && make \
    && cp replicatord /usr/local/sbin/replicatord \
    && mkdir /usr/local/etc/replicatord \
    && cp replicatord.yml /usr/local/etc/replicatord/replicatord.yml \
    && cd .. \
    && rm -r mysql-tarantool-replication \
    && yum -y remove git ncurses-devel cmake make gcc-c++ boost boost-devel wget unzip nano bzip2 mysql-devel mysql-lib

COPY "replicatord.service" "/etc/systemd/system"

EXPOSE 3301
CMD ["/usr/sbin/init"]