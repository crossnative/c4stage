FROM golang:1.21 as builder

WORKDIR /go/src/c4stage
COPY . .
RUN make build

FROM gcr.io/distroless/static@sha256:6706c73aae2afaa8201d63cc3dda48753c09bcd6c300762251065c0f7e602b25
USER nonroot:nonroot
ENV LAUNCHPAD_ENV "production"
COPY --from=builder --chown=nonroot:nonroot /go/src/c4stage/build /
CMD ["/c4stage"]