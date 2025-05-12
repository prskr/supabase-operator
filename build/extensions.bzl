load(":repositories.bzl", "supabase_compose_file", "supabase_migration_files", "supabase_init_script_files")

def _supbase_compose_impl(_ctx):
    supabase_compose_file()

supabase_compose = module_extension(
    implementation = _supbase_compose_impl,
)

def _supbase_migrations_impl(_ctx):
    supabase_migration_files()

supabase_migrations = module_extension(
    implementation = _supbase_migrations_impl,
)

def _supbase_init_script_impl(_ctx):
    supabase_init_script_files()

supabase_init_scripts = module_extension(
    implementation = _supbase_init_script_impl,
)
