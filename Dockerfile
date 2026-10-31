FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /portal .
FROM alpine:3
WORKDIR /app
COPY --from=build /portal /app/portal
COPY static/ /app/static/
COPY curriculum/ /app/curriculum/
RUN mkdir -p /app/data && chown 10001:10001 /app/data
USER 10001:10001
ENV ADDR=0.0.0.0:3000 DATA_FILE=/app/data/progress.json
EXPOSE 3000
ENTRYPOINT ["/app/portal"]
