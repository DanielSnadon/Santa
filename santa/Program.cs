using System.Threading.RateLimiting;
using Microsoft.AspNetCore.Antiforgery;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.DataProtection;
using System.Net;
using Microsoft.AspNetCore.HttpOverrides;
using Microsoft.EntityFrameworkCore;
using santa.Application;
using santa.Components;
using santa.Data;
using santa.Infrastructure;

var builder = WebApplication.CreateBuilder(args);
var maxRequestBytes = builder.Configuration.GetValue<long>("Resources:MaxRequestBodyBytes", 1024 * 1024);
var maxConcurrentDraws = builder.Configuration.GetValue<int>("Resources:MaxConcurrentDraws", 2);
if (maxRequestBytes <= 0 || maxConcurrentDraws <= 0)
    throw new InvalidOperationException("Resource limits must be positive");
builder.WebHost.ConfigureKestrel(o => o.Limits.MaxRequestBodySize = maxRequestBytes);
builder.Services.AddRazorComponents().AddInteractiveServerComponents();
builder.Services.AddHttpContextAccessor();
var protection = builder.Services.AddDataProtection().SetApplicationName("Santa");
if (builder.Configuration["DataProtection:Path"] is { Length: > 0 } keyPath)
    protection.PersistKeysToFileSystem(new DirectoryInfo(keyPath));
builder.Services.Configure<ForwardedHeadersOptions>(o => {
    o.ForwardedHeaders = ForwardedHeaders.XForwardedFor | ForwardedHeaders.XForwardedProto;
    foreach (var proxy in builder.Configuration.GetSection("TrustedProxies").Get<string[]>() ?? [])
        o.KnownProxies.Add(IPAddress.Parse(proxy));
});
builder.Services.AddCascadingAuthenticationState();
builder.Services.AddDbContextFactory<SantaDb>(options => options.UseNpgsql(
    builder.Configuration.GetConnectionString("Santa") ?? throw new InvalidOperationException("ConnectionStrings:Santa is required")));
builder.Services.AddIdentity<Account, IdentityRole<Guid>>(o => {
    o.User.RequireUniqueEmail = true;
    o.Password.RequiredLength = 12;
    o.Lockout.MaxFailedAccessAttempts = 5;
    o.Lockout.DefaultLockoutTimeSpan = TimeSpan.FromMinutes(15);
}).AddEntityFrameworkStores<SantaDb>().AddDefaultTokenProviders();
builder.Services.ConfigureApplicationCookie(o => {
    o.LoginPath = "/login"; o.AccessDeniedPath = "/login";
    o.Cookie.HttpOnly = true; o.Cookie.SameSite = SameSiteMode.Strict;
    o.Cookie.SecurePolicy = builder.Environment.IsDevelopment() ? CookieSecurePolicy.SameAsRequest : CookieSecurePolicy.Always;
    o.ExpireTimeSpan = TimeSpan.FromHours(8);
    o.Events.OnRedirectToLogin = ctx => {
        if (ctx.Request.Path.StartsWithSegments("/api")) ctx.Response.StatusCode = 401;
        else ctx.Response.Redirect(ctx.RedirectUri);
        return Task.CompletedTask;
    };
});
builder.Services.AddAntiforgery(o => o.HeaderName = "X-CSRF-TOKEN");
builder.Services.AddSingleton<NativeAssignmentEngine>();
builder.Services.AddSingleton<IAssignmentEngine, EngineMetrics>();
builder.Services.AddScoped<IGameStore, EfGameStore>();
builder.Services.AddScoped<GameService>();
builder.Services.AddRateLimiter(o => {
    o.RejectionStatusCode = 429;
    // Bound aggregate native memory/CPU even when many accounts publish at once.
    o.GlobalLimiter = PartitionedRateLimiter.Create<HttpContext, string>(ctx =>
        HttpMethods.IsPost(ctx.Request.Method) &&
        (ctx.Request.Path.StartsWithSegments("/api/games") || ctx.Request.Path.StartsWithSegments("/actions")) &&
        ctx.Request.Path.Value!.EndsWith("/publish", StringComparison.OrdinalIgnoreCase)
            ? RateLimitPartition.GetConcurrencyLimiter("draws", _ => new ConcurrencyLimiterOptions {
                PermitLimit = maxConcurrentDraws, QueueLimit = 0 })
            : RateLimitPartition.GetNoLimiter("other"));
    o.AddPolicy("auth", ctx => RateLimitPartition.GetFixedWindowLimiter(ctx.Connection.RemoteIpAddress?.ToString() ?? "unknown",
        _ => new FixedWindowRateLimiterOptions { PermitLimit = 20,Window = TimeSpan.FromMinutes(1),QueueLimit = 0 }));
    o.AddPolicy("draw", ctx => RateLimitPartition.GetConcurrencyLimiter(ctx.User.Identity?.Name ?? "unknown",
        _ => new ConcurrencyLimiterOptions { PermitLimit = 1,QueueLimit = 0 }));
});
var app = builder.Build();
if (args.Contains("--migrate")) {
    await using var scope = app.Services.CreateAsyncScope();
    await scope.ServiceProvider.GetRequiredService<SantaDb>().Database.MigrateAsync();
    return;
}
app.UseForwardedHeaders();
if (!app.Environment.IsDevelopment()) { app.UseExceptionHandler("/Error"); app.UseHsts(); app.UseHttpsRedirection(); }
app.Use(async (ctx,next) => {
    ctx.Response.Headers.CacheControl = "no-store";
    ctx.Response.Headers["X-Content-Type-Options"] = "nosniff";
    ctx.Response.Headers["X-Frame-Options"] = "DENY";
    ctx.Response.Headers["Referrer-Policy"] = "same-origin";
    try { await next(); }
    catch (GameException ex) {
        if (ctx.Request.Path.StartsWithSegments("/actions")) {
            var segments = ctx.Request.Path.Value!.Split('/');
            var target = segments.Length > 2 && Guid.TryParse(segments[2],out var id) ? $"/games/{id}" : "/";
            ctx.Response.Redirect($"{target}?error={Uri.EscapeDataString(ex.Code)}");
        } else {
            ctx.Response.StatusCode = ex.Status;
            await ctx.Response.WriteAsJsonAsync(new { error = ex.Code });
        }
    }
    catch (AntiforgeryValidationException) { ctx.Response.StatusCode = 400; }
});
app.UseAuthentication(); app.UseAuthorization(); app.UseRateLimiter(); app.UseAntiforgery();
app.Use(async (ctx,next) => {
    if (ctx.Request.Path.StartsWithSegments("/api") && !HttpMethods.IsGet(ctx.Request.Method) && !HttpMethods.IsHead(ctx.Request.Method))
        await ctx.RequestServices.GetRequiredService<IAntiforgery>().ValidateRequestAsync(ctx);
    await next();
});
app.MapSanta();
app.MapStaticAssets();
app.MapRazorComponents<App>().AddInteractiveServerRenderMode();
app.Run();
public partial class Program;
