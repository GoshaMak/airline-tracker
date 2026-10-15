from collections import defaultdict
from html import escape

files = defaultdict(lambda: [0, 0])

with open("coverage.out", encoding="utf-8") as f:
    next(f)

    for line in f:
        location, statements, count = line.split()
        filename = location.rsplit(":", 1)[0]

        n = int(statements)
        files[filename][0] += n

        if int(count) > 0:
            files[filename][1] += n

rows = []

for filename, (total, covered) in sorted(files.items()):
    pct = covered / total * 100 if total else 0
    if pct == 0:
        continue
    color = "#22c55e" if pct >= 80 else "#f59e0b" if pct >= 50 else "#ef4444"

    rows.append(f"""
      <tr>
        <td>{escape(filename)}</td>
        <td>
          <div class="bar">
            <div style="width:{pct:.1f}%;background:{color}"></div>
          </div>
        </td>
        <td><b>{pct:.1f}%</b></td>
      </tr>
    """)

html = """
<!DOCTYPE html>
<html lang="en">
<meta charset="UTF-8">
<title>Go Coverage Report</title>
<style>
body { font-family: sans-serif; margin: 40px; }
table { border-collapse: collapse; width: 100%; }
td, th { padding: 12px; border-bottom: 1px solid #ddd; text-align: left; }
.bar { background: #eee; height: 12px; border-radius: 8px; overflow: hidden; }
.bar div { height: 100%; }
</style>
<h2>Go Test Coverage</h2>
<table>
<tr><th>File</th><th>Coverage</th><th>Percent</th></tr>
""" + "\n".join(rows) + "</table></html>"

with open("coverage.html", "w", encoding="utf-8") as f:
    f.write(html)
