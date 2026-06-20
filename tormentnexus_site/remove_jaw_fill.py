with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

old_jaw = """                // Gaping jaw
                ctx.fillStyle = 'rgba(0,80,130,' + (o * 0.2) + ')';
                ctx.beginPath();
                ctx.arc(0, s*0.55, s*0.75, 0.1, Math.PI*0.9 + jaw);
                ctx.lineTo(-s*0.55, s*0.45);
                ctx.closePath();
                ctx.fill();
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.2) + ')';
                ctx.lineWidth = 0.5;
                ctx.stroke();"""

new_jaw = """                // Gaping jaw — outline only, no dark fill
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.1) + ')';
                ctx.lineWidth = 0.5;
                ctx.beginPath();
                ctx.arc(0, s*0.55, s*0.75, 0.1, Math.PI*0.9 + jaw);
                ctx.lineTo(-s*0.55, s*0.45);
                ctx.closePath();
                ctx.stroke();"""

html = html.replace(old_jaw, new_jaw)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Dark jaw fill removed")
