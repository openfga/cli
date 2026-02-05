FROM gcr.io/distroless/static:nonroot
COPY fga /fga
ENTRYPOINT ["/fga"]
