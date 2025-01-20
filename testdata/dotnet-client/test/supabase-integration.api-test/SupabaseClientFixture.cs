using Supabase;
using TUnit.Core.Interfaces;

namespace supabase_integration.api_test;

public class SupabaseClientFixture : IAsyncInitializer
{
    public Task InitializeAsync()
    {
        ApiClient = new Client(
            Environment.GetEnvironmentVariable("SUPBASE_URL") ?? "http://localhost:8000",
            Environment.GetEnvironmentVariable("SUPBASE_ACCESS_KEY") ?? throw new ArgumentException("Supabase access key is missing."),
            new SupabaseOptions
            {
                AutoConnectRealtime = false
            }
            );
        return Task.CompletedTask;
    }
    
    public Client ApiClient { get; private set; }
}