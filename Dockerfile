FROM golang:alpine AS builder
WORKDIR /
ARG REF
RUN apk add git make &&\
    git clone https://github.com/thomasgame/trojan-go-extra.git
RUN if [[ -z "${REF}" ]]; then \
        echo "No specific commit provided, use the latest one." \
    ;else \
        echo "Use commit ${REF}" &&\
        cd trojan-go-extra &&\
        git checkout ${REF} \
    ;fi
RUN cd trojan-go-extra &&\
    make &&\
    wget https://github.com/v2fly/domain-list-community/raw/release/dlc.dat -O build/geosite.dat &&\
    wget https://github.com/v2fly/geoip/raw/release/geoip.dat -O build/geoip.dat &&\
    wget https://github.com/v2fly/geoip/raw/release/geoip-only-cn-private.dat -O build/geoip-only-cn-private.dat

FROM alpine
WORKDIR /
RUN apk add --no-cache tzdata ca-certificates
COPY --from=builder /trojan-go-extra/build /usr/local/bin/
COPY --from=builder /trojan-go-extra/example/server.json /etc/trojan-go-extra/config.json

ENTRYPOINT ["/usr/local/bin/trojan-go-extra", "-config"]
CMD ["/etc/trojan-go-extra/config.json"]
