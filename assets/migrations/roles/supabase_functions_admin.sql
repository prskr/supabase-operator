DO
$do$
BEGIN
   IF EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'supabase_functions_admin') THEN

      RAISE NOTICE 'Role "supabase_functions_admin" already exists. Skipping.';
   ELSE
      CREATE ROLE supabase_functions_admin CREATEROLE NOINHERIT;
   END IF;
END
$do$;


alter user supabase_functions_admin
set
    search_path = supabase_functions;
