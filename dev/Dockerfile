FROM alpine:3.21
WORKDIR /app
ADD --chown=65532:65532 ./out ./bin
USER 65532:65532
ENTRYPOINT [ "/app/bin/supabase-operator" ]
