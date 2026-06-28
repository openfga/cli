FROM gcr.io/distroless/static-debian13:nonroot

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/fga /fga
ENTRYPOINT ["/fga"]
