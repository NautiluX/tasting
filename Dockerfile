FROM golang:latest AS build
COPY . /root/tasting
WORKDIR /root/tasting
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /root/tasting/tasting .
COPY --from=build /root/tasting/content ./content
RUN ls /root/
CMD ["/root/tasting"]
