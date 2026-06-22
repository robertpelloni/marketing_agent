with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

# 1. Add RobotSkull class before SpiderDrone
old_drone = """        // Drone class — creepy robot spider
        class SpiderDrone"""

new_skull = """        // Robot Skull class — floating skull with jaw animation
        class RobotSkull {
            constructor() { this.reset(); }
            reset() {
                this.x = Math.random() * W;
                this.y = Math.random() * H * 0.3 + H * 0.1;
                this.size = 20 + Math.random() * 25;
                this.dx = (Math.random() - 0.5) * 0.3;
                this.dy = (Math.random() - 0.5) * 0.15;
                this.jawOpen = 0;
                this.jawSpeed = 0.02 + Math.random() * 0.03;
                this.opacity = 0.12 + Math.random() * 0.18;
                this.eyeGlow = 0;
                this.driftAngle = Math.random() * Math.PI * 2;
            }
            update() {
                this.x += this.dx + Math.sin(this.driftAngle) * 0.2;
                this.y += this.dy + Math.cos(this.driftAngle) * 0.1;
                this.driftAngle += 0.01;
                this.jawOpen += this.jawSpeed;
                if (this.jawOpen > 1 || this.jawOpen < 0) this.jawSpeed *= -1;
                this.eyeGlow = 0.3 + Math.sin(Date.now() * 0.003 + this.x) * 0.2;
                if (this.x < -80) this.x = W + 80;
                if (this.x > W + 80) this.x = -80;
                if (this.y < -80) this.y = H * 0.8;
                if (this.y > H * 0.8 + 80) this.y = -80;
            }
            draw() {
                ctx.save();
                ctx.translate(this.x, this.y);
                const s = this.size;
                const o = this.opacity;
                ctx.shadowColor = 'rgba(0,240,255,' + (o * 0.3) + ')';
                ctx.shadowBlur = 25;
                // Skull dome
                ctx.fillStyle = 'rgba(0,200,255,' + o + ')';
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.5) + ')';
                ctx.lineWidth = 1;
                ctx.beginPath();
                ctx.arc(0, -s * 0.1, s * 0.7, Math.PI, 0);
                ctx.fill();
                // Cheeks
                ctx.fillStyle = 'rgba(0,180,230,' + (o * 0.4) + ')';
                ctx.beginPath(); ctx.ellipse(-s*0.5, s*0.1, s*0.25, s*0.3, 0, 0, Math.PI*2); ctx.fill();
                ctx.beginPath(); ctx.ellipse(s*0.5, s*0.1, s*0.25, s*0.3, 0, 0, Math.PI*2); ctx.fill();
                // Jaw
                const jaw = this.jawOpen * 0.3;
                ctx.fillStyle = 'rgba(0,100,150,' + (o * 0.25) + ')';
                ctx.beginPath(); ctx.arc(0, s*0.15, s*0.5, 0, Math.PI + jaw); ctx.fill();
                // Upper teeth
                ctx.fillStyle = 'rgba(0,240,255,' + (o * 0.35) + ')';
                for (let i = -3; i <= 3; i++) { if(i===0) continue; ctx.fillRect(i*s*0.12, -s*0.05, s*0.06, s*0.1); }
                // Lower teeth
                for (let i = -3; i <= 3; i++) { if(i===0) continue; ctx.fillRect(i*s*0.12, s*0.1+jaw*s*0.3, s*0.06, s*0.1); }
                // Eye sockets
                ctx.fillStyle = 'rgba(0,0,0,0.8)';
                ctx.beginPath(); ctx.arc(-s*0.25, -s*0.1, s*0.15, 0, Math.PI*2); ctx.fill();
                ctx.beginPath(); ctx.arc(s*0.25, -s*0.1, s*0.15, 0, Math.PI*2); ctx.fill();
                // Red eyes
                ctx.shadowColor = 'rgba(255,0,0,' + this.eyeGlow + ')';
                ctx.shadowBlur = 18;
                ctx.fillStyle = 'rgba(255,0,50,' + this.eyeGlow + ')';
                ctx.beginPath(); ctx.arc(-s*0.25, -s*0.1, s*0.06, 0, Math.PI*2); ctx.fill();
                ctx.beginPath(); ctx.arc(s*0.25, -s*0.1, s*0.06, 0, Math.PI*2); ctx.fill();
                // Nose
                ctx.fillStyle = 'rgba(0,0,0,0.4)';
                ctx.beginPath(); ctx.moveTo(0, s*0.02); ctx.lineTo(-s*0.05, s*0.12); ctx.lineTo(s*0.05, s*0.12); ctx.closePath(); ctx.fill();
                // Cyber lines
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.15) + ')';
                ctx.lineWidth = 0.5;
                ctx.beginPath(); ctx.moveTo(-s*0.4,-s*0.4); ctx.lineTo(-s*0.2,-s*0.3); ctx.lineTo(-s*0.4,-s*0.2); ctx.stroke();
                ctx.beginPath(); ctx.moveTo(s*0.4,-s*0.4); ctx.lineTo(s*0.2,-s*0.3); ctx.lineTo(s*0.4,-s*0.2); ctx.stroke();
                ctx.shadowBlur = 0;
                ctx.restore();
            }
        }

        // Drone class — creepy robot spider
        class SpiderDrone"""

html = html.replace(old_drone, new_skull)

# 2. Add skulls to the entities array
old_entities = "const spiders = [];\n        const lasers = [];\n        const eyes = [];\n        for (let i = 0; i < 10; i++) spiders.push(new SpiderDrone());\n        for (let i = 0; i < 30; i++) eyes.push(new TerminatorEye());"
new_entities = "const spiders = [];\n        const lasers = [];\n        const eyes = [];\n        const skulls = [];\n        for (let i = 0; i < 6; i++) skulls.push(new RobotSkull());\n        for (let i = 0; i < 10; i++) spiders.push(new SpiderDrone());\n        for (let i = 0; i < 30; i++) eyes.push(new TerminatorEye());"
html = html.replace(old_entities, new_entities)

# 3. Add skulls to animate loop
old_anim = "// Update and draw eyes\n            eyes.forEach(e => { e.update(); e.draw(); });\n\n            // Update and draw spiders"
new_anim = "// Update and draw skulls\n            skulls.forEach(s => { s.update(); s.draw(); });\n\n            // Update and draw eyes\n            eyes.forEach(e => { e.update(); e.draw(); });\n\n            // Update and draw spiders"
html = html.replace(old_anim, new_anim)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)

print("Robot skulls and spiders injected")
print(f"RobotSkull class: {html.count('class RobotSkull')}")
print(f"SpiderDrone class: {html.count('class SpiderDrone')}")
print(f"skulls array: {html.count('skulls')}")
