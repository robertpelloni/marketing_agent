with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

# Rounder, taller skull dome
old_dome = """                // Elongated skull dome
                ctx.fillStyle = 'rgba(0,180,230,' + o + ')';
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.5) + ')';
                ctx.lineWidth = 1.5;
                ctx.beginPath();
                ctx.arc(0, -s*0.15, s*0.7, Math.PI*1.1, -Math.PI*0.1);
                ctx.lineTo(-s*0.5, s*0.25);
                ctx.lineTo(-s*0.55, s*0.05);
                ctx.closePath();
                ctx.fill();
                ctx.stroke();
                ctx.beginPath();
                ctx.moveTo(s*0.5, -s*0.1);
                ctx.lineTo(s*0.55, s*0.05);
                ctx.lineTo(s*0.5, s*0.25);
                ctx.closePath();
                ctx.fill();
                ctx.stroke();"""

new_dome = """                // Rounder, taller skull dome
                ctx.fillStyle = 'rgba(0,180,230,' + o + ')';
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.5) + ')';
                ctx.lineWidth = 1.5;
                ctx.beginPath();
                ctx.arc(0, -s*0.2, s*0.75, Math.PI*1.05, -Math.PI*0.05);
                ctx.quadraticCurveTo(-s*0.55, s*0.1, -s*0.5, s*0.3);
                ctx.quadraticCurveTo(0, s*0.35, s*0.5, s*0.3);
                ctx.quadraticCurveTo(s*0.55, s*0.1, s*0.75, -s*0.05);
                ctx.closePath();
                ctx.fill();
                ctx.stroke();"""

html = html.replace(old_dome, new_dome)

# Remove the old angular temple ridges since the dome now covers them
old_temples = """                // Right temple ridge
                ctx.beginPath();
                ctx.moveTo(s*0.5, -s*0.1);
                ctx.lineTo(s*0.55, s*0.05);
                ctx.lineTo(s*0.5, s*0.1);
                ctx.closePath();
                ctx.fill();
                ctx.stroke();"""
html = html.replace(old_temples, "")

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Rounder, taller skull dome")
