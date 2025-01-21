create table public.categories (
  id bigint primary key generated always as identity,
  name text not null unique
);

create table public.lists (
  id bigint primary key generated always as identity,
  user_id uuid not null references auth.users (id),
  name text not null
);

create table public.tasks (
  id bigint primary key generated always as identity,
  list_id bigint not null references lists (id),
  category_id bigint references categories (id),
  name text not null,
  description text,
  due_date date,
  priority int,
  completed boolean default false
);
