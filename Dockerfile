FROM golang:alpine as builder

ARG BUCKET
ARG KEY
ARG DAY
ARG DD_SITE
ARG DD_API_KEY
ARG DD_APP_KEY
ARG SLO_ID

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV BUCKET=$BUCKET
ENV KEY=$KEY
ENV DAY=$DAY
ENV DD_SITE=$DD_SITE
ENV DD_API_KEY=$DD_API_KEY
ENV DD_APP_KEY=$DD_APP_KEY
ENV SLO_ID=$SLO_ID

WORKDIR /build
COPY . ./
#RUN go mod init main
#RUN go get github.com/aws/aws-sdk-go
#RUN go get github.com/DataDog/datadog-api-client-go/v2/api/datadog
#RUN BUCKET=$BUCKET KEY=$KEY DAY=$DAY DD_SITE=$DD_SITE DD_API_KEY=$DD_API_KEY DD_APP_KEY=$DD_APP_KEY SLO_ID=$SLO_ID go run main.go
RUN go mod download
RUN go build -o main .
WORKDIR /dist
RUN cp /build/main .

FROM scratch
COPY --from=builder /dist/main .
ENTRYPOINT ["/main"]