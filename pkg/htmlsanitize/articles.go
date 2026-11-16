package htmlsanitize

import (
	"regexp"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	articlePolicyOnce sync.Once
	articlePolicy     *bluemonday.Policy
)

// Article keeps the formatting produced by the CMS editor and removes active content.
func Article(value string) string {
	articlePolicyOnce.Do(func() {
		articlePolicy = bluemonday.NewPolicy()
		articlePolicy.AllowElements("p", "br", "h2", "h3", "strong", "b", "em", "i", "u", "s", "blockquote", "ul", "ol", "li", "code", "pre", "hr", "a")
		articlePolicy.AllowAttrs("href").OnElements("a")
		articlePolicy.AllowAttrs("target").Matching(regexp.MustCompile(`^_blank$`)).OnElements("a")
		articlePolicy.AllowURLSchemes("http", "https", "mailto")
		articlePolicy.RequireParseableURLs(true)
		articlePolicy.RequireNoFollowOnFullyQualifiedLinks(true)
		articlePolicy.RequireNoReferrerOnFullyQualifiedLinks(true)
		articlePolicy.AddTargetBlankToFullyQualifiedLinks(true)
	})
	return articlePolicy.Sanitize(value)
}
