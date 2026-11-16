using Ganss.Xss;
using Markdig;

namespace santa.Application;

public static class SafeMarkdown
{
    private static readonly MarkdownPipeline Pipeline = new MarkdownPipelineBuilder()
        .UseAdvancedExtensions()
        .DisableHtml()
        .Build();

    private static readonly string[] Tags =
        ["p", "br", "strong", "em", "del", "code", "pre", "blockquote",
         "ul", "ol", "li", "h1", "h2", "h3", "h4", "a", "hr",
         "table", "thead", "tbody", "tr", "th", "td"];

    public static string Render(string? markdown)
    {
        if (string.IsNullOrWhiteSpace(markdown)) return "";
        var sanitizer = new HtmlSanitizer();
        sanitizer.AllowedTags.Clear();
        foreach (var tag in Tags) sanitizer.AllowedTags.Add(tag);
        sanitizer.AllowedAttributes.Clear();
        sanitizer.AllowedAttributes.Add("href");
        sanitizer.AllowedAttributes.Add("title");
        sanitizer.AllowedSchemes.Clear();
        sanitizer.AllowedSchemes.Add("http");
        sanitizer.AllowedSchemes.Add("https");
        sanitizer.AllowedSchemes.Add("mailto");
        sanitizer.UriAttributes.Clear();
        sanitizer.UriAttributes.Add("href");
        sanitizer.AllowedCssProperties.Clear();
        sanitizer.AllowedAtRules.Clear();
        return sanitizer.Sanitize(Markdown.ToHtml(markdown,Pipeline));
    }
}
