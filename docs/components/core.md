# Core

The `Core` resource configures the essential Supabase services:

- PostgREST
- Auth

and it manages the initial DB migrations that are not done by the services themselves and are part of the [supabase/postgres](https://github.com/supabase/postgres) repository.

To deploy a basic `Core` instance, you can use the following snippet as a 'template':

```yaml title="Basic 'Core' example" linenums="1"
--8<-- "config/samples/supabase_v1alpha1_core.yaml"
```
