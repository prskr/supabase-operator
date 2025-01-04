create user supabase_functions_admin createrole noinherit;

alter user supabase_functions_admin
set
    search_path = supabase_functions;
