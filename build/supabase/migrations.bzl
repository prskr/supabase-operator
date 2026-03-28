"""Rules for fetching Supabase SQL migrations"""

def _supbase_migrations_impl(repository_ctx):
    repository_ctx.download_and_extract(
        "https://github.com/supabase/postgres/archive/refs/tags/{}.tar.gz".format(repository_ctx.attr.version),
        sha256 = repository_ctx.attr.sha256,
        type = "tar.gz",
        strip_prefix = "postgres-{}/{}".format(repository_ctx.attr.version, repository_ctx.attr.path_prefix),
    )

    repository_ctx.file("BUILD.bazel", content = """
filegroup(
    name = "{}",
    srcs = glob(["*/*.sql"]),
    visibility = ["//visibility:public"],
)
""".format(repository_ctx.original_name))

supabase_migrations = repository_rule(
    implementation = _supbase_migrations_impl,
    attrs = {
        "version": attr.string(mandatory = True),
        "sha256": attr.string(mandatory = True),
        "path_prefix": attr.string(mandatory = True),
    },
)
