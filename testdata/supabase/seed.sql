INSERT INTO auth.users (id, email, encrypted_password, created_at, email_confirmed_at)
VALUES ('4B9BE43A-4CEE-45E6-8B4B-EA5D69F39056','ted.tester@example.com', '', now(), now());

INSERT INTO categories (name)
VALUES ('Groceries'), ('Work'), ('Personal'), ('Other');

INSERT INTO public.lists (user_id, name)
VALUES ( '4B9BE43A-4CEE-45E6-8B4B-EA5D69F39056', 'Groceries');

INSERT INTO public.tasks (list_id, category_id, name)
VALUES (1, 1, 'Orange Juice');