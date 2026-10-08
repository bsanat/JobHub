package greenhouse

import (
	"strings"
	"testing"
)

func TestFormattedTextKeepsReadableStructureWithoutExecutableMarkup(t *testing.T) {
	got := FormattedText(`<script>alert(1)</script><h2>Requirements</h2><p><strong>Berlin only</strong> applicants</p><ul><li>German C1</li><li>Work permit</li></ul><p>Apply<br>online</p>`)
	for _, want := range []string{"## Requirements", "**Berlin only** applicants", "- German C1", "- Work permit", "Apply\nonline"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "alert") || strings.Contains(got, "<") {
		t.Fatalf("unsafe markup survived: %q", got)
	}
}

func TestFormattedTextPreservesExistingMarkdown(t *testing.T) {
	got := FormattedText("**Міндеттер:**\n\n* Бірінші талап\n* Екінші талап")
	if !strings.Contains(got, "**Міндеттер:**") || !strings.Contains(got, "* Бірінші талап") || !strings.Contains(got, "* Екінші талап") {
		t.Fatalf("Markdown was flattened: %q", got)
	}
}
