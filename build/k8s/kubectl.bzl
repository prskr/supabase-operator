"""Rules for applying Kubernetes manifests using kubectl."""

def _kubectl_apply_impl(ctx):
    script = ctx.actions.declare_file("kubectl_apply.sh")
    script_content = """\
#!/bin/bash
set -e
{kubectl} apply --context {context} --server-side=true --filename {manifest}
""".format(
        kubectl = ctx.executable.kubectl.short_path,
        context = ctx.attr.context,
        manifest = ctx.file.manifest.short_path,
    )

    ctx.actions.write(script, script_content, is_executable = True)
    runfiles = ctx.runfiles(files = [ctx.file.manifest, ctx.executable.kubectl])

    return [DefaultInfo(executable = script, runfiles = runfiles)]

kubectl_apply = rule(
    implementation = _kubectl_apply_impl,
    attrs = {
        "context": attr.string(mandatory = True),
        "manifest": attr.label(allow_single_file = True, mandatory = True),
        "kubectl": attr.label(
            executable = True,
            cfg = "exec",
            allow_files = True,
            default = None,
        ),
    },
    executable = True,
)
