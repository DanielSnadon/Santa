using System.Diagnostics;
using System.Diagnostics.Metrics;
using santa.Application;
namespace santa.Infrastructure;
public sealed class EngineMetrics(NativeAssignmentEngine engine) : IAssignmentEngine
{
    private static readonly Meter Meter = new("Santa.Engine","1.0.0");
    private static readonly Counter<long> Calls = Meter.CreateCounter<long>("santa.draw.calls");
    private static readonly Histogram<double> Duration = Meter.CreateHistogram<double>("santa.draw.duration","ms");
    public EngineResult Draw(EngineInput input)
    {
        var start = Stopwatch.GetTimestamp();
        var result = engine.Draw(input);
        var tag = new KeyValuePair<string,object?>("status",result.Status.ToString());
        Calls.Add(1,tag); Duration.Record(Stopwatch.GetElapsedTime(start).TotalMilliseconds,tag);
        return result;
    }
}
