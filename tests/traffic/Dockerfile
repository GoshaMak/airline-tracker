FROM alpine:3.23.3
RUN apk add --no-cache curl jq tcpdump
COPY scenario.sh /scenario.sh
CMD ["sh", "/scenario.sh"]
