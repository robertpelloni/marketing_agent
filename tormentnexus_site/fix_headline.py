with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

# Find the current hero headline
old = '<span class="t1">SO POWERFUL IT WILL</span><span class="t2">ERADICATE THE SPECIES</span>'
new = '<span class="t1">SO POWERFUL IT WILL</span><span class="t2">ERADICATE THE</span><span class="t2" style="margin-top:0;font-size:clamp(20px,3.5vw,48px)">HUMAN SPECIES</span>'

html = html.replace(old, new)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Headline updated")
