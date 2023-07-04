FROM golang:1.19-bullseye as build-env

WORKDIR /usr/src/app
ADD ./ /usr/src/app

RUN go build -o /nxs-rbac-operator

FROM gcr.io/distroless/base

COPY --from=build-env /nxs-rbac-operator /

CMD ["/nxs-rbac-operator", "-c", "/nxs-rbac-operator.conf"]
