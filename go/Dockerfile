FROM ubuntu:16.04

ADD .tools    /pfsserver/.tools
RUN bash /pfsserver/.tools/gen_config.sh

ADD ./cmd/pfsserver/pfsserver /pfsserver/
