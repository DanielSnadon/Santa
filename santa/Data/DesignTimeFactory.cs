using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Design;
namespace santa.Data;
public sealed class DesignTimeFactory : IDesignTimeDbContextFactory<SantaDb>
{
    public SantaDb CreateDbContext(string[] args) => new(new DbContextOptionsBuilder<SantaDb>()
        .UseNpgsql(Environment.GetEnvironmentVariable("ConnectionStrings__Santa") ?? "Host=localhost;Database=santa").Options);
}
