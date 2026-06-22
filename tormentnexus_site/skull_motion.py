with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

# Add rotation, bobbing, and slow approach to RobotSkull constructor
old_reset = """            constructor() { 
                this.zoomTimer = 0;
                this.zoomActive = false;
                this.zoomScale = 1;
                this.reset(); 
            }
            reset() {
                this.x = Math.random() * W;
                this.y = Math.random() * H * 0.3 + H * 0.1;
                this.size = 20 + Math.random() * 25;
                this.dx = (Math.random() - 0.5) * 0.3;
                this.dy = (Math.random() - 0.5) * 0.15;
                this.jawOpen = 0;
                this.jawSpeed = 0.05 + Math.random() * 0.08;
                this.opacity = 0.35 + Math.random() * 0.35;
                this.eyeGlow = 0;
                this.driftAngle = Math.random() * Math.PI * 2;
            }"""

new_reset = """            constructor() { 
                this.zoomTimer = 0;
                this.zoomActive = false;
                this.zoomScale = 1;
                this.rotation = 0;
                this.targetRotation = 0;
                this.rotTimer = 0;
                this.bobPhase = Math.random() * Math.PI * 2;
                this.approachPhase = 0;
                this.reset(); 
            }
            reset() {
                this.x = Math.random() * W;
                this.y = Math.random() * H * 0.3 + H * 0.1;
                this.size = 20 + Math.random() * 25;
                this.baseSize = this.size;
                this.dx = (Math.random() - 0.5) * 0.3;
                this.dy = (Math.random() - 0.5) * 0.15;
                this.jawOpen = 0;
                this.jawSpeed = 0.05 + Math.random() * 0.08;
                this.opacity = 0.35 + Math.random() * 0.35;
                this.eyeGlow = 0;
                this.driftAngle = Math.random() * Math.PI * 2;
                this.approachPhase = 0;
            }"""

html = html.replace(old_reset, new_reset)

# Add rotation, bobbing, approach to update
old_update = """            update() {
                this.x += this.dx + Math.sin(this.driftAngle) * 0.2;
                this.y += this.dy + Math.cos(this.driftAngle) * 0.1;
                this.driftAngle += 0.01;
                this.jawOpen += this.jawSpeed;
                if (this.jawOpen > 1 || this.jawOpen < 0) this.jawSpeed *= -1;
                this.eyeGlow = 0.5 + Math.sin(Date.now() * 0.004 + this.x) * 0.35;
                if (this.x < -80) this.x = W + 80;
                if (this.x > W + 80) this.x = -80;
                if (this.y < -80) this.y = H * 0.8;
                if (this.y > H * 0.8 + 80) this.y = -80;
                
                // Sudden zoom toward screen
                this.zoomTimer++;
                if (!this.zoomActive && this.zoomTimer > 120 + Math.random() * 200) {
                    this.zoomActive = true;
                    this.zoomTimer = 0;
                    this.zoomScale = 1;
                }
                if (this.zoomActive) {
                    this.zoomScale += 0.15 + Math.random() * 0.1;
                    this.zoomTimer++;
                    if (this.zoomTimer > 15 || this.zoomScale > 6) {
                        this.zoomActive = false;
                        this.zoomTimer = 0;
                        this.zoomScale = 1;
                    }
                }
            }"""

new_update = """            update() {
                this.x += this.dx + Math.sin(this.driftAngle) * 0.2;
                this.y += this.dy + Math.cos(this.driftAngle) * 0.1;
                this.driftAngle += 0.01;
                this.jawOpen += this.jawSpeed;
                if (this.jawOpen > 1 || this.jawOpen < 0) this.jawSpeed *= -1;
                this.eyeGlow = 0.5 + Math.sin(Date.now() * 0.004 + this.x) * 0.35;
                if (this.x < -80) this.x = W + 80;
                if (this.x > W + 80) this.x = -80;
                if (this.y < -80) this.y = H * 0.8;
                if (this.y > H * 0.8 + 80) this.y = -80;
                
                // Sudden zoom toward screen
                this.zoomTimer++;
                if (!this.zoomActive && this.zoomTimer > 120 + Math.random() * 200) {
                    this.zoomActive = true;
                    this.zoomTimer = 0;
                    this.zoomScale = 1;
                }
                if (this.zoomActive) {
                    this.zoomScale += 0.15 + Math.random() * 0.1;
                    this.zoomTimer++;
                    if (this.zoomTimer > 15 || this.zoomScale > 6) {
                        this.zoomActive = false;
                        this.zoomTimer = 0;
                        this.zoomScale = 1;
                    }
                }
                
                // Stuttery rotation — jerky, unpredictable
                this.rotTimer++;
                if (this.rotTimer > 15 + Math.random() * 30) {
                    this.targetRotation = (Math.random() - 0.5) * 0.4;
                    this.rotTimer = 0;
                }
                this.rotation += (this.targetRotation - this.rotation) * 0.08;
                // Add micro-jitter to rotation
                this.rotation += (Math.random() - 0.5) * 0.008;
                
                // Bob up and down
                this.bobPhase += 0.015 + Math.random() * 0.01;
                const bobAmount = Math.sin(this.bobPhase) * 5;
                this.y += Math.sin(this.bobPhase * 1.3) * 0.3;
                
                // Slow approach — gradually move toward viewer
                this.approachPhase += 0.001 + Math.random() * 0.002;
                this.size = this.baseSize * (1 + this.approachPhase * 0.5);
                if (this.approachPhase > 4) {
                    this.x = Math.random() * W;
                    this.y = Math.random() * H * 0.3 + H * 0.1;
                    this.approachPhase = 0;
                    this.size = this.baseSize;
                }
            }"""

html = html.replace(old_update, new_update)

# Apply rotation in draw
old_rotate = """                ctx.save();
                ctx.translate(this.x, this.y);"""

new_rotate = """                ctx.save();
                ctx.translate(this.x, this.y);
                ctx.rotate(this.rotation);"""

html = html.replace(old_rotate, new_rotate)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Skulls now rotate, bob, and approach slowly")
PYEOF
