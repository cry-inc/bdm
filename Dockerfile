FROM golang:1-alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/bdm/
COPY . .
RUN go version
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
ENV BDM_WRITE_TOKEN=
ENV BDM_CERT_CACHE=/bdmcerts
ENV BDM_HTTPS_CERT=
ENV BDM_HTTPS_KEY=
ENV BDM_LETS_ENCRYPT=
ENV BDM_MAX_FILE_SIZE=0
ENV BDM_MAX_PACKAGE_SIZE=0
ENV BDM_MAX_FILE_COUNT=0
ENV BDM_MAX_PATH_LENGTH=0

CMD bdm -server -port=${BDM_PORT} -writetoken=${BDM_WRITE_TOKEN} -store=${BDM_STORE} \
        -httpscert=${BDM_HTTPS_CERT} -httpskey=${BDM_HTTPS_KEY} \
        -certcache=${BDM_CERT_CACHE} -letsencrypt=${BDM_LETS_ENCRYPT} \
        -maxfilesize=${BDM_MAX_FILE_SIZE} -maxsize=${BDM_MAX_PACKAGE_SIZE} \
        -maxpath=${BDM_MAX_PATH_LENGTH} -maxfiles=${BDM_MAX_FILE_COUNT}
