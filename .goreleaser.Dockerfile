FROM gcr.io/distroless/static:nonroot
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/fga /fga
ENTRYPOINT ["/fga"]
