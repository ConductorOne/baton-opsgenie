FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-opsgenie"]
COPY baton-opsgenie /