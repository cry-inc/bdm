FROM golang:1-alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/bdm/
COPY . .
RUN go build -o /go/bin/bdm
RUN mkdir /bdmstore
RUN mkdir /bdmcerts

FROM alpine
COPY --from=builder /go/bin/bdm /bin/bdm
COPY --from=builder /bdmstore /bdmstore
COPY --from=builder /bdmcerts /bdmcerts

EXPOSE 2323

ENV BDM_PORT=2323
ENV BDM_STORE=/bdmstore
ENV BDM_KEY=
ENV BDM_CERT_CACHE=/bdmcerts
ENV BDM_HTTPS_CERT=
ENV BDM_HTTPS_KEY=
ENV BDM_LETS_ENCRYPT=

CMD bdm -server -port=${BDM_PORT} -key=${BDM_KEY} -store=${BDM_STORE} -certcache=${BDM_CERT_CACHE} -httpscert=${BDM_HTTPS_CERT} -httpskey=${BDM_HTTPS_KEY} -letsencrypt=${BDM_LETS_ENCRYPT}
