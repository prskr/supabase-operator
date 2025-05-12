load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")
load(":supabase.bzl", "SUPABASE_POSTGRES_TAG", "SUPABASE_TAG")

def supabase_compose_file():
    http_file(
        name = "supabase_compose",
        downloaded_file_path = "docker-compose.yml",
        sha256 = "ff44c48a501cf1c18103027b3a8b3b0d2fa51766b35bba38832447f8c4863afd",
        url = "https://raw.githubusercontent.com/supabase/supabase/refs/tags/{}/docker/docker-compose.yml".format(SUPABASE_TAG),
        visibility = ["//visibility:public"],
    )

def supabase_migration_files():
    http_archive(
        name = "supabase_migrations",
        sha256 = "a76308fa6e1d6adbd7eba9cb203335b1e000e1b9e2f94f63dc648c7b6259f3f9",
        url = "https://github.com/supabase/postgres/archive/refs/tags/{}.tar.gz".format(SUPABASE_POSTGRES_TAG),
        strip_prefix = "postgres-{}/migrations/db/migrations".format(SUPABASE_POSTGRES_TAG),
        build_file_content = """
filegroup(
    name = "migrations",
    srcs = glob(["*.sql"]),
    visibility = ["//visibility:public"],
)
        """,
        visibility = ["//visibility:public"],
    )

def supabase_init_script_files():
    http_archive(
        name = "supabase_init_scripts",
        sha256 = "a76308fa6e1d6adbd7eba9cb203335b1e000e1b9e2f94f63dc648c7b6259f3f9",
        url = "https://github.com/supabase/postgres/archive/refs/tags/{}.tar.gz".format(SUPABASE_POSTGRES_TAG),
        strip_prefix = "postgres-{}/migrations/db/init-scripts".format(SUPABASE_POSTGRES_TAG),
        build_file_content = """
filegroup(
    name = "init_scripts",
    srcs = glob(["*.sql"]),
    visibility = ["//visibility:public"],
)
        """,
        visibility = ["//visibility:public"],
    )
