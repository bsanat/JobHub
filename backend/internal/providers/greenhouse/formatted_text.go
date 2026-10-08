package greenhouse

import (
	"html"
	"strings"

	htmlparser "golang.org/x/net/html"
)

// FormattedText keeps safe paragraph, heading and list boundaries from provider
// HTML while returning Markdown text. Executable markup and attributes are
// discarded; the frontend renders Markdown without raw HTML.
func FormattedText(value string) string {
	for i := 0; i < 2; i++ {
		value = html.UnescapeString(value)
	}
	node, err := htmlparser.Parse(strings.NewReader(value))
	if err != nil {
		return ""
	}
	var out strings.Builder
	var walk func(*htmlparser.Node)
	walk = func(n *htmlparser.Node) {
		if n.Type == htmlparser.ElementNode {
			switch n.Data {
			case "script", "style", "noscript", "iframe", "svg":
				return
			case "p", "div", "section", "article", "ul", "ol":
				out.WriteString("\n\n")
			case "li":
				out.WriteString("\n- ")
			case "br":
				out.WriteByte('\n')
			case "h1", "h2", "h3", "h4", "h5", "h6":
				out.WriteString("\n\n## ")
			case "strong", "b":
				out.WriteString("**")
			case "em", "i":
				out.WriteByte('*')
			}
		}
		if n.Type == htmlparser.TextNode {
			out.WriteString(n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if n.Type == htmlparser.ElementNode {
			switch n.Data {
			case "p", "div", "section", "article", "ul", "ol", "h1", "h2", "h3", "h4", "h5", "h6":
				out.WriteString("\n\n")
			case "strong", "b":
				out.WriteString("**")
			case "em", "i":
				out.WriteByte('*')
			}
		}
	}
	walk(node)
	lines := strings.Split(out.String(), "\n")
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" && (len(clean) == 0 || clean[len(clean)-1] == "") {
			continue
		}
		clean = append(clean, line)
	}
	return strings.TrimSpace(strings.Join(clean, "\n"))
}
