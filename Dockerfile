FROM golang:alpine AS build-env
RUN  apk add --no-cache git make ca-certificates
LABEL maintaner="@amimof (amir.mofasser@gmail.com)"
COPY . /go/src/github.com/amimof/synka
WORKDIR /go/src/github.com/amimof/synka
RUN make

FROM scratch
COPY --from=build-env /go/src/github.com/amimof/synka/bin/synka /go/bin/synka
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/synka"]