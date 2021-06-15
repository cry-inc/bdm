FROM golang:1-alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/bdm/
COPY . .
RUN go version
RUN go build -o /go/bin/bdm
RUN mkdir /bdmdata/
RUN mkdir /bdmdata/store
RUN mkdir /bdmdata/certs

FROM alpine
COPY --from=builder /go/bin/bdm /bin/bdm
COPY --from=builder /bdmdata /bdmdata

EXPOSE 2323

ENV BDM_PORT=2323
ENV BDM_STORE=/bdmdata/store
ENV BDM_DEFAULT_USER=admin
ENV BDM_CERT_CACHE=/bdmdata/certs
ENV BDM_USERS_FILE=/bdmdata/users.json
ENV BDM_TOKENS_FILE=/bdmdata/tokens.json
ENV BDM_HTTPS_CERT=
ENV BDM_HTTPS_KEY=
ENV BDM_LETS_ENCRYPT=
ENV BDM_MAX_FILE_SIZE=0
ENV BDM_MAX_PACKAGE_SIZE=0
ENV BDM_MAX_FILE_COUNT=0
ENV BDM_MAX_PATH_LENGTH=0

CMD bdm -server -port=${BDM_PORT} -defaultuser=${BDM_DEFAULT_USER} -store=${BDM_STORE} \
        -httpscert=${BDM_HTTPS_CERT} -httpskey=${BDM_HTTPS_KEY} \
        -certcache=${BDM_CERT_CACHE} -letsencrypt=${BDM_LETS_ENCRYPT} \
        -maxfilesize=${BDM_MAX_FILE_SIZE} -maxsize=${BDM_MAX_PACKAGE_SIZE} \
        -maxpath=${BDM_MAX_PATH_LENGTH} -maxfiles=${BDM_MAX_FILE_COUNT} \
        -usersfile=${BDM_USERS_FILE} -tokensfile=${BDM_TOKENS_FILE}
