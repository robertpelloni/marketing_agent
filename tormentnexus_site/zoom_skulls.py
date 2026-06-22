with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

# Update RobotSkull constructor to add zoom properties
old_init = """            constructor() { this.reset(); }
            reset() {
                this.x = Math.random() * W;
                this.y = Math.random() * H * 0.3 + H * 0.1;"""

new_init = """            constructor() { 
                this.zoomTimer = 0;
                this.zoomActive = false;
                this.zoomScale = 1;
                this.reset(); 
            }
            reset() {
                this.x = Math.random() * W;
                this.y = Math.random() * H * 0.3 + H * 0.1;"""

html = html.replace(old_init, new_init)

# Speed up jaw animation
html = html.replace(
    "this.jawSpeed = 0.02 + Math.random() * 0.03;",
    "this.jawSpeed = 0.05 + Math.random() * 0.08;",
)

# Add zoom logic to update()
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
            }"""

html = html.replace(old_update, new_update)

# Apply zoom scale in draw()
old_draw_header = """            draw() {
                ctx.save();
                ctx.translate(this.x, this.y);
                const s = this.size;
                const o = this.opacity;
                const jaw = this.jawOpen * 1.0;"""

new_draw_header = """            draw() {
                ctx.save();
                ctx.translate(this.x, this.y);
                const s = this.size * this.zoomScale;
                const o = this.opacity * (1 + this.zoomScale * 0.1);
                const jaw = this.jawOpen * 1.0 * (1 + this.zoomScale * 0.05);"""

html = html.replace(old_draw_header, new_draw_header)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Skulls now zoom and chomp faster")
