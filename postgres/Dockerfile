# syntax=docker/dockerfile:1.13

ARG POSTGRES_MAJOR=17
ARG POSTGRES_MINOR=5
ARG IMAGE_FLAVOR=standard
ARG DISTRO=bookworm
ARG IMAGE_BASE=ghcr.io/cloudnative-pg/postgresql:${POSTGRES_MAJOR}.${POSTGRES_MINOR}-${IMAGE_FLAVOR}-${DISTRO}

FROM ${IMAGE_BASE} AS base

ENV DEBIAN_FRONTEND=noninteractive

ARG POSTGRES_MAJOR

USER root

RUN apt-get update && \
    apt-get install --yes --no-install-recommends \
    libsodium23 \
    postgresql-${POSTGRES_MAJOR}-cron \
    postgresql-${POSTGRES_MAJOR}-http \
    postgresql-${POSTGRES_MAJOR}-hypopg \
    postgresql-${POSTGRES_MAJOR}-pgrouting \
    postgresql-${POSTGRES_MAJOR}-pgtap \
    postgresql-${POSTGRES_MAJOR}-plpgsql-check \
    postgresql-${POSTGRES_MAJOR}-postgis-3 \
    postgresql-${POSTGRES_MAJOR}-wal2json \
    postgresql-${POSTGRES_MAJOR}-rum \
    postgresql-${POSTGRES_MAJOR}-repack \
    postgresql-${POSTGRES_MAJOR}-timescaledb && \
    rm -rf /var/lib/apt/lists/*

USER 26

FROM ${IMAGE_BASE} AS builder

ARG POSTGRES_MAJOR
ARG PG_SAFEUPDATE_VERSION=1.5
ARG PG_NET_VERSION=v0.9.1
ARG PGSODIUM_VERSION=v3.1.9
ARG PG_HASHIDS_VERSION=v1.2.1
ARG PG_TLE_VERSION=v1.5.1
ARG PG_GRAPHQL_VERSION=v1.5.11
ARG PG_SUPABASE_VAULT_VERSION=v0.3.1
ARG SUPABASE_WRAPPER_VERSION=v0.5.4
ARG INDEX_ADVISOR_VERSION=v0.2.0
ARG PG_JSONSCHEMA_VERSION=v0.3.3
ARG PG_PLAN_FILTER_REVISION=5081a7b

ENV DEBIAN_FRONTEND=noninteractive

USER root

RUN apt-get update && \
    apt-get install --yes --no-install-recommends \
    curl \
    pkg-config \
    ca-certificates \
    postgresql-server-dev-${POSTGRES_MAJOR} \
    libcurl4-openssl-dev \
    libsodium-dev \
    git \
    build-essential \
    flex \
    libkrb5-dev

RUN curl -sSf https://sh.rustup.rs | sh -s -- -y && \
    export PATH="/root/.cargo/bin:$PATH" && \
    export "PG${POSTGRES_MAJOR}_PG_CONFIG=/usr/lib/postgresql/${POSTGRES_MAJOR}/bin/pg_config" && \
    cargo install cargo-pgrx --version 0.12.9 --locked && \
    cargo pgrx init

WORKDIR /postgres/extensions

ENV PATH="/root/.cargo/bin:$PATH"

# Install pgjwt
RUN git clone https://github.com/michelp/pgjwt.git && \
    make -C pgjwt && make -C pgjwt install

# Install pg_hashids
RUN git clone --branch ${PG_HASHIDS_VERSION} --depth 1 https://github.com/iCyberon/pg_hashids.git && \
    make -C pg_hashids && make -C pg_hashids install

# Install pg-safeupdate
RUN git clone --branch ${PG_SAFEUPDATE_VERSION} --depth 1 https://github.com/eradman/pg-safeupdate.git && \
    make -C pg-safeupdate && make -C pg-safeupdate install

# install pg_net
RUN git clone --branch ${PG_NET_VERSION} --depth 1 https://github.com/supabase/pg_net.git && \
    make -C pg_net && make -C pg_net install

# install pg_graphql
RUN git clone --branch ${PG_GRAPHQL_VERSION} --depth 1 https://github.com/supabase/pg_graphql.git && \
    cd pg_graphql && \
    cargo pgrx install --release

# install supabase vault
RUN git clone --branch ${PG_SUPABASE_VAULT_VERSION} --depth 1 https://github.com/supabase/vault.git && \
    make -C vault && make -C vault install

# Install matching cargo-pgrx for jsonschema
RUN curl -sSf https://sh.rustup.rs | sh -s -- -y && \
    export PATH="/root/.cargo/bin:$PATH" && \
    export "PG${POSTGRES_MAJOR}_PG_CONFIG=/usr/lib/postgresql/${POSTGRES_MAJOR}/bin/pg_config" && \
    cargo install cargo-pgrx --version 0.12.6 --locked && \
    cargo pgrx init

# install pg_jsonschema
RUN git clone --branch ${PG_JSONSCHEMA_VERSION} --depth 1 https://github.com/supabase/pg_jsonschema.git && \
    cd pg_jsonschema && \
    cargo pgrx install --release

# Install matching cargo-pgrx for wrappers
RUN curl -sSf https://sh.rustup.rs | sh -s -- -y && \
    export PATH="/root/.cargo/bin:$PATH" && \
    export "PG${POSTGRES_MAJOR}_PG_CONFIG=/usr/lib/postgresql/${POSTGRES_MAJOR}/bin/pg_config" && \
    cargo install cargo-pgrx --version 0.14.3 --locked && \
    cargo pgrx init

# install supabase wrappers
RUN git clone --branch ${SUPABASE_WRAPPER_VERSION} --depth 1 https://github.com/supabase/wrappers.git && \
    cd wrappers/wrappers && \
    cargo pgrx install --release --features "pg${POSTGRES_MAJOR},all_fdws"

# install index_advisor
RUN git clone --branch ${INDEX_ADVISOR_VERSION} --depth 1 https://github.com/supabase/index_advisor.git && \
    make -C index_advisor && make -C index_advisor install

# install pgsodium
RUN git clone --branch ${PGSODIUM_VERSION} --depth 1 https://github.com/michelp/pgsodium.git && \
    make -C pgsodium && make -C pgsodium install

# install pg_tle
RUN git clone --branch ${PG_TLE_VERSION} --depth 1 https://github.com/aws/pg_tle.git && \
    make -C pg_tle && make -C pg_tle install

# install pg_plan_filter
RUN git clone https://github.com/pgexperts/pg_plan_filter.git && \
    git -C pg_plan_filter checkout ${PG_PLAN_FILTER_REVISION} && \
    make -C pg_plan_filter && make -C pg_plan_filter install

FROM base AS final

ARG POSTGRES_MAJOR

# Copy all bitcode additions
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/bitcode /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/bitcode

# pgjwt
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/pgjwt* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# pg-safeupdate
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/safeupdate.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/

# pg_hashids
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/pg_hashids.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/pg_hashids* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# pg_net
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/pg_net.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/pg_net* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# pg_graphql
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/pg_graphql.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/pg_graphql* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# supabase vault
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/supabase_vault.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/supabase_vault* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# pg_jsonschema
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/pg_jsonschema.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/pg_jsonschema* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# Supabase wrappers
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/wrappers-*.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/wrappers* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# pgsodium
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/pgsodium.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/pgsodium* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# pg_tle
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/pg_tle.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/pg_tle* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/

# pg_plan_filter
COPY --from=builder /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/plan_filter.so /usr/lib/postgresql/${POSTGRES_MAJOR}/lib/

# index_advisor
COPY --from=builder /usr/share/postgresql/${POSTGRES_MAJOR}/extension/index_advisor* /usr/share/postgresql/${POSTGRES_MAJOR}/extension/
