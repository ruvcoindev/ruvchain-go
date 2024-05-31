
FROM mirror.gcr.io/golang:alpine as builder 

COPY . /src
WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git && /src/build && go build -o /src/genkeys cmd/genkeys/main.go

FROM mirror.gcr.io/alpine

COPY --from=builder /src/ruvchain  /usr/bin/ruvchain
COPY --from=builder /src/ruvchainctl /usr/bin/ruvchainctl
COPY --from=builder /src/genkeys /usr/bin/genkeys
COPY --from=builder  /src/contrib/docker/entrypoint.sh /usr/bin/entrypoint.sh

# RUN addgroup -g 1000 -S ruvchain \
#  && adduser -u 1000 -S -g 1000 --home /etc/ruvchain ruvchain
#
# USER ruvchain
# TODO: Make running unprivileged work

VOLUME [ "/etc/ruvchain" ]

ENTRYPOINT [ "/usr/bin/entrypoint.sh" ]
