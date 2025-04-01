using Supabase.Postgrest.Attributes;
using Supabase.Postgrest.Models;

namespace supabase_integration.api_test;

public class ServiceKeyTest
{
    [Test]
    [ClassDataSource<SupabaseClientFixture>(Shared = SharedType.PerAssembly)]
    public async Task TestListTasks(SupabaseClientFixture fixture)
    {
        var resp = await fixture.ApiClient.Postgrest.Table<TaskList>().Get().ConfigureAwait(false);

        await Assert.That(resp.Models.Count).IsGreaterThan(0);
    }
}

[Table("lists")]
public class TaskList : BaseModel
{
    [PrimaryKey("id")]
    public int Id { get; set; }
    [Column("user_id")]
    public Guid UserId { get; set; }
    [Column("name")]
    public string Name { get; set; }
}
