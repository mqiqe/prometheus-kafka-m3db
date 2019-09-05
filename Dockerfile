FROM golang:1.11-alpine as build

RUN apk add --no-cache alpine-sdk librdkafka librdkafka-dev

WORKDIR /src/prometheus-kafka-m3db
ADD . /src/prometheus-kafka-m3db

RUN go build -o /prometheus-kafka-m3db

FROM alpine:3.8

RUN apk add --no-cache librdkafka

#COPY --from=build /src/prometheus-kafka-m3db/schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-m3db /

CMD /prometheus-kafka-m3db
