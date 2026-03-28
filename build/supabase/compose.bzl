"""Rule for fetching Supabase docker-compose.yml"""

def _supbase_compose_impl(repository_ctx):
    repository_ctx.download(
        "https://raw.githubusercontent.com/supabase/supabase/refs/tags/{}/docker/docker-compose.yml".format(repository_ctx.attr.version),
        sha256 = repository_ctx.attr.sha256,
        output = "docker-compose.yml",
    )

    repository_ctx.file("BUILD.bazel", content = """
filegroup(
    name = "docker-compose",
    srcs = ["docker-compose.yml"],
    visibility = ["//visibility:public"],
)
""".format(repository_ctx.original_name))

supabase_compose = repository_rule(
    implementation = _supbase_compose_impl,
    attrs = {
        "version": attr.string(mandatory = True),
        "sha256": attr.string(mandatory = True),
    },
)
