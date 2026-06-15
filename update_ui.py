import os

with open('internal/web/server.go', 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    if '</select>' in line and 'update_channel' in line:
        new_lines.append(line)
        new_lines.append('\t\t\t\t\t</form>\n')
        new_lines.append('\t\t\t\t\t<div style="margin-top: 5px;">\n')
        new_lines.append('\t\t\t\t\t\t<form onsubmit="simulateInbound(event, %d)" style="display:flex; gap: 5px;">\n' % c.ID if False else '\t\t\t\t\t\t<form onsubmit="simulateInbound(event, %d)" style="display:flex; gap: 5px;">\n') # placeholder
        continue
    # This is getting complex with nested sprintf. I will just rewrite the entire handleDashboard method.
    new_lines.append(line)

# Let's try a simpler approach: append a script and a template at the end of head or body.
