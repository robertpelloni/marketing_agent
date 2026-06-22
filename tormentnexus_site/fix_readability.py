with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

# Increase all font sizes
sizes = [
    ("font-size:5px", "font-size:10px"),
    ("font-size:6px", "font-size:11px"),
    ("font-size:7px", "font-size:12px"),
    ("font-size:8px", "font-size:13px"),
    ("font-size:9px", "font-size:13px"),
    ("font-size:10px", "font-size:14px"),
    ("font-size:11px", "font-size:14px"),
    ("font-size:clamp(9px,0.9vw,11px)", "font-size:clamp(14px,1.4vw,18px)"),
    ("font-size:clamp(8px,1vw,14px)", "font-size:clamp(14px,1.8vw,22px)"),
    ("font-size:clamp(18px,2.5vw,28px)", "font-size:clamp(28px,4vw,48px)"),
    ("font-size:clamp(26px,5vw,64px)", "font-size:clamp(36px,6vw,80px)"),
    ("font-size:20px", "font-size:32px"),
    ("font-size:18px", "font-size:28px"),
    ("font-size:clamp(12px,1.5vw,20px)", "font-size:clamp(16px,2vw,26px)"),
]
for old, new in sizes:
    html = html.replace(old, new)

# Increase opacity
opacities = [
    ("opacity:0.35", "opacity:0.75"),
    ("opacity:0.25", "opacity:0.65"),
    ("opacity:0.5", "opacity:0.8"),
    ("opacity:0.6", "opacity:0.85"),
    ("opacity:0.4", "opacity:0.7"),
    ("opacity:0.3", "opacity:0.65"),
    ("opacity:0.15", "opacity:0.45"),
    ("opacity:0.05", "opacity:0.25"),
    ("opacity:0.7", "opacity:0.9"),
    ("opacity:0.8", "opacity:0.95"),
    ("opacity:0.9", "opacity:1"),
]
for old, new in opacities:
    html = html.replace(old, new)

# Brighten dim colors
html = html.replace("var(--text-dimmer)", "#556677")
html = html.replace("var(--text-dim)", "#7799bb")
html = html.replace("var(--text)", "#e4ecf4")

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Readability fixed")
