// Package views contains templ components for the dashboard UI.
// This stub satisfies the compiler until dashboard.templ is written and generated.
package views

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/niski84/claude-usage-monitor/internal/types"
)

// Dashboard returns a templ component that renders the main dashboard page.
// Replace this stub with the real templ-generated version once dashboard.templ is written.
func Dashboard(stats types.Stats) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, dashboardHTML(stats))
		return err
	})
}

func dashboardHTML(s types.Stats) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en" class="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="color-scheme" content="dark light">
<title>Claude Usage Monitor</title>
<script>
(function(){
  var t = localStorage.getItem('theme');
  if (t === 'light') { document.documentElement.classList.remove('dark'); }
  else { document.documentElement.classList.add('dark'); }
})();
</script>
<style>
  :root { color-scheme: dark; }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui, sans-serif; background: #0f1117; color: #e2e8f0; padding: 2rem; }
  h1 { font-size: 1.5rem; margin-bottom: 1rem; color: #a78bfa; }
  p { color: #94a3b8; margin-bottom: 0.5rem; }
  .note { background: #1e1b4b; border: 1px solid #4c1d95; border-radius: 8px; padding: 1.5rem; max-width: 600px; }
  a { color: #818cf8; }
</style>
</head>
<body>
  <h1>Claude Usage Monitor</h1>
  <div class="note">
    <p>Dashboard stub — replace this file with the real templ component.</p>
    <p>Total cost: $%.4f | Sessions: %d</p>
    <p>Full stats at <a href="/api/stats">/api/stats</a></p>
  </div>
</body>
</html>`, s.TotalCostUSD, s.Sessions)
}
