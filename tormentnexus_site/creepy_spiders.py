with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

old_spider = """        // Drone class — creepy robot spider
        class SpiderDrone {
            constructor() { this.reset(); }
            reset() {
                this.x = Math.random() * W;
                this.y = Math.random() * H;
                this.size = 8 + Math.random() * 12;
                this.angle = Math.random() * Math.PI * 2;
                this.speed = 1 + Math.random() * 3;
                this.legPhase = Math.random() * Math.PI * 2;
                this.legSpeed = 0.05 + Math.random() * 0.1;
                this.opacity = 0.5 + Math.random() * 0.35;
                this.laserTimer = 0;
                this.laserFreq = 30 + Math.random() * 60;
                this.targetX = Math.random() * W;
                this.targetY = Math.random() * H;
                this.state = 'hunt';
            }
            update() {
                const dx = this.targetX - this.x;
                const dy = this.targetY - this.y;
                const dist = Math.sqrt(dx * dx + dy * dy);
                if (dist < 20 || this.state === 'teleport') {
                    this.targetX = Math.random() * W;
                    this.targetY = Math.random() * H;
                    if (Math.random() < 0.15) {
                        this.x = Math.random() * W;
                        this.y = Math.random() * H;
                        this.state = 'hunt';
                        return;
                    }
                    this.state = 'hunt';
                }
                const speed = this.speed * (0.5 + Math.random());
                this.x += (dx / dist) * speed;
                this.y += (dy / dist) * speed;
                this.angle = Math.atan2(dy, dx);
                this.legPhase += this.legSpeed;
                this.laserTimer++;
                if (this.laserTimer > this.laserFreq) {
                    this.laserTimer = 0;
                    this.laserFreq = 20 + Math.random() * 80;
                    return 'fire';
                }
                return null;
            }
            draw() {
                ctx.save();
                ctx.translate(this.x, this.y);
                const gradient = ctx.createRadialGradient(0, 0, 0, 0, 0, this.size);
                gradient.addColorStop(0, 'rgba(0, 200, 255, ' + this.opacity + ')');
                gradient.addColorStop(0.5, 'rgba(0, 100, 200, ' + (this.opacity * 0.5) + ')');
                gradient.addColorStop(1, 'rgba(0, 50, 100, 0)');
                ctx.fillStyle = gradient;
                ctx.beginPath();
                ctx.arc(0, 0, this.size, 0, Math.PI * 2);
                ctx.fill();
                const numLegs = 8;
                const legLength = this.size * 2;
                for (let i = 0; i < numLegs; i++) {
                    const legAngle = (i / numLegs) * Math.PI * 2 + this.legPhase;
                    const bendAngle = legAngle + Math.sin(this.legPhase + i) * 0.8;
                    ctx.strokeStyle = 'rgba(0, 240, 255, ' + (this.opacity * 0.3) + ')';
                    ctx.lineWidth = 0.5;
                    ctx.beginPath();
                    ctx.moveTo(0, 0);
                    const midX = Math.cos(bendAngle) * legLength * 0.6;
                    const midY = Math.sin(bendAngle) * legLength * 0.6;
                    ctx.lineTo(midX, midY);
                    ctx.lineTo(Math.cos(bendAngle + 0.3) * legLength, Math.sin(bendAngle + 0.3) * legLength);
                    ctx.stroke();
                }
                const eyeSize = this.size * 0.25;
                ctx.shadowColor = 'rgba(255, 0, 0, 0.8)';
                ctx.shadowBlur = 15;
                ctx.fillStyle = 'rgba(255, 0, 50, 0.8)';
                ctx.beginPath();
                ctx.arc(this.size * 0.2, -this.size * 0.15, eyeSize, 0, Math.PI * 2);
                ctx.fill();
                ctx.shadowBlur = 25;
                ctx.fillStyle = 'rgba(255, 0, 0, 0.4)';
                ctx.beginPath();
                ctx.arc(this.size * 0.2, -this.size * 0.15, eyeSize * 0.5, 0, Math.PI * 2);
                ctx.fill();
                ctx.shadowBlur = 0;
                ctx.restore();
            }
        }"""

new_spider = """        // Spider Drone — cybernetic murder machine with machine guns
        class SpiderDrone {
            constructor() {
                this.reset();
                this.gunFireTimer = 0;
                this.gunFireRate = 3 + Math.random() * 5;
                this.burstCount = 0;
                this.maxBurst = 4 + Math.floor(Math.random() * 6);
                this.reloading = false;
                this.reloadTimer = 0;
                this.reloadTime = 20 + Math.random() * 30;
                this.bodyTwitch = 0;
            }
            reset() {
                this.x = Math.random() * W;
                this.y = Math.random() * H;
                this.size = 10 + Math.random() * 14;
                this.angle = Math.random() * Math.PI * 2;
                this.speed = 0.8 + Math.random() * 2.5;
                this.legPhase = Math.random() * Math.PI * 2;
                this.legSpeed = 0.04 + Math.random() * 0.08;
                this.opacity = 0.55 + Math.random() * 0.35;
                this.targetX = Math.random() * W;
                this.targetY = Math.random() * H;
                this.state = 'hunt';
                this.legScissor = 0;
            }
            update() {
                const dx = this.targetX - this.x;
                const dy = this.targetY - this.y;
                const dist = Math.sqrt(dx * dx + dy * dy);
                this.bodyTwitch = Math.sin(Date.now() * 0.01 + this.x) * 2;
                this.legScissor = Math.sin(this.legPhase * 2) * 0.3;
                if (dist < 30 || this.state === 'teleport') {
                    this.targetX = Math.random() * W;
                    this.targetY = Math.random() * H;
                    if (Math.random() < 0.1) {
                        this.x = Math.random() * W;
                        this.y = Math.random() * H;
                        this.state = 'hunt';
                        return 'teleport';
                    }
                    this.state = 'hunt';
                }
                const speed = this.speed * (0.3 + Math.random() * 0.7);
                this.x += (dx / dist) * speed + Math.sin(this.legPhase * 3) * 0.3;
                this.y += (dy / dist) * speed + Math.cos(this.legPhase * 2) * 0.3;
                this.angle = Math.atan2(dy, dx);
                this.legPhase += this.legSpeed;

                // Machine gun logic
                if (this.reloading) {
                    this.reloadTimer++;
                    if (this.reloadTimer > this.reloadTime) {
                        this.reloading = false;
                        this.burstCount = 0;
                        this.reloadTimer = 0;
                    }
                    return null;
                }
                this.gunFireTimer++;
                if (this.gunFireTimer > this.gunFireRate && this.burstCount < this.maxBurst) {
                    this.gunFireTimer = 0;
                    this.burstCount++;
                    return 'machinegun';
                }
                if (this.burstCount >= this.maxBurst) {
                    this.reloading = true;
                    this.reloadTime = 15 + Math.random() * 25;
                }
                return null;
            }
            draw() {
                ctx.save();
                ctx.translate(this.x + this.bodyTwitch, this.y);
                const s = this.size;
                const o = this.opacity;

                // Shadow
                ctx.shadowColor = 'rgba(0,240,255,' + (o * 0.2) + ')';
                ctx.shadowBlur = 20;

                // Main body — elongated armored carapace
                ctx.fillStyle = 'rgba(0,180,230,' + o + ')';
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.4) + ')';
                ctx.lineWidth = 1;
                ctx.beginPath();
                ctx.ellipse(0, 0, s*0.4, s*0.5, 0, 0, Math.PI*2);
                ctx.fill();
                ctx.stroke();

                // Armor plates on back
                ctx.fillStyle = 'rgba(0,130,190,' + (o * 0.5) + ')';
                ctx.beginPath();
                ctx.ellipse(0, -s*0.15, s*0.3, s*0.15, 0, Math.PI, 0);
                ctx.fill();
                ctx.beginPath();
                ctx.ellipse(0, s*0.1, s*0.25, s*0.12, 0, Math.PI, 0);
                ctx.fill();

                // Head — smaller, forward
                ctx.fillStyle = 'rgba(0,200,240,' + o + ')';
                ctx.beginPath();
                ctx.arc(s*0.5, -s*0.2, s*0.25, 0, Math.PI*2);
                ctx.fill();

                // Multiple glowing red eyes (4 eyes, compound)
                for (let ei = 0; ei < 4; ei++) {
                    const ex = s*0.5 + Math.cos(ei*1.2)*s*0.12;
                    const ey = -s*0.2 + Math.sin(ei*1.2)*s*0.08;
                    ctx.shadowColor = 'rgba(255,0,0,0.8)';
                    ctx.shadowBlur = 15;
                    ctx.fillStyle = 'rgba(255,0,30,0.8)';
                    ctx.beginPath();
                    ctx.arc(ex, ey, s*0.05, 0, Math.PI*2);
                    ctx.fill();
                    ctx.shadowBlur = 20;
                    ctx.fillStyle = 'rgba(255,200,200,0.5)';
                    ctx.beginPath();
                    ctx.arc(ex, ey, s*0.02, 0, Math.PI*2);
                    ctx.fill();
                }
                ctx.shadowBlur = 0;

                // Mandibles — pincer-like
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.5) + ')';
                ctx.lineWidth = 1.5;
                ctx.beginPath();
                ctx.moveTo(s*0.65, -s*0.15);
                ctx.lineTo(s*0.9, -s*0.25);
                ctx.lineTo(s*0.8, -s*0.1);
                ctx.stroke();
                ctx.beginPath();
                ctx.moveTo(s*0.65, s*0.05);
                ctx.lineTo(s*0.9, s*0.15);
                ctx.lineTo(s*0.8, 0);
                ctx.stroke();

                // Machine guns — mounted on front legs
                const gunRecoil = this.gunFireTimer < 2 && !this.reloading ? s*0.05 : 0;
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.7) + ')';
                ctx.lineWidth = 2;
                // Left gun
                ctx.beginPath();
                ctx.moveTo(s*0.3, -s*0.35);
                ctx.lineTo(s*0.8 + gunRecoil, -s*0.45);
                ctx.stroke();
                ctx.fillStyle = 'rgba(0,150,200,' + (o * 0.5) + ')';
                ctx.fillRect(s*0.5, -s*0.5, s*0.3, s*0.08);
                // Right gun
                ctx.beginPath();
                ctx.moveTo(s*0.3, s*0.25);
                ctx.lineTo(s*0.8 + gunRecoil, s*0.35);
                ctx.stroke();
                ctx.fillRect(s*0.5, s*0.32, s*0.3, s*0.08);
                // Gun flash
                if (this.gunFireTimer < 2 && !this.reloading) {
                    ctx.fillStyle = 'rgba(255,200,50,0.8)';
                    ctx.beginPath();
                    ctx.arc(s*0.85, -s*0.45, s*0.06, 0, Math.PI*2);
                    ctx.fill();
                    ctx.fillStyle = 'rgba(255,255,200,0.4)';
                    ctx.beginPath();
                    ctx.arc(s*0.85, -s*0.45, s*0.12, 0, Math.PI*2);
                    ctx.fill();
                    ctx.fillStyle = 'rgba(255,200,50,0.8)';
                    ctx.beginPath();
                    ctx.arc(s*0.85, s*0.35, s*0.06, 0, Math.PI*2);
                    ctx.fill();
                    ctx.fillStyle = 'rgba(255,255,200,0.4)';
                    ctx.beginPath();
                    ctx.arc(s*0.85, s*0.35, s*0.12, 0, Math.PI*2);
                    ctx.fill();
                }

                // 8 legs — spidery, angular, jointed
                const numLegs = 8;
                const legLength = s * 2.2;
                for (let i = 0; i < numLegs; i++) {
                    const legAngle = (i / numLegs) * Math.PI * 2 + this.legPhase;
                    const joint = Math.sin(this.legPhase * 1.5 + i * 0.8) * 0.6;
                    const scissor = (i % 2 === 0 ? 1 : -1) * this.legScissor;
                    const isLeft = i < 4 ? -1 : 1;
                    ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.35 - Math.abs(joint) * 0.1) + ')';
                    ctx.lineWidth = 1 + (i % 2 === 0 ? 0.5 : 0);
                    ctx.beginPath();
                    ctx.moveTo(0, 0);
                    // Upper leg segment
                    const jx = Math.cos(legAngle + joint + scissor) * legLength * 0.45;
                    const jy = Math.sin(legAngle + joint + scissor) * legLength * 0.45;
                    ctx.lineTo(jx, jy);
                    // Lower leg segment (bent)
                    const ex2 = jx + Math.cos(legAngle + joint * 1.5 + scissor * 0.5) * legLength * 0.55;
                    const ey2 = jy + Math.sin(legAngle + joint * 1.5 + scissor * 0.5) * legLength * 0.55;
                    ctx.lineTo(ex2, ey2);
                    // Sharp tip
                    ctx.stroke();
                }

                // Body segmentation lines
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.1) + ')';
                ctx.lineWidth = 0.5;
                for (let seg = -2; seg <= 2; seg++) {
                    ctx.beginPath();
                    ctx.moveTo(-s*0.3, seg*s*0.12);
                    ctx.lineTo(s*0.3, seg*s*0.12);
                    ctx.stroke();
                }

                ctx.shadowBlur = 0;
                ctx.restore();
            }
        }"""

html = html.replace(old_spider, new_spider)

# Add bullet entities and machine gun fire to the animation
old_laser = """        // Laser class
        class Laser {"""

new_bullet = """        // Bullet class — machine gun rounds
        class Bullet {
            constructor(x1, y1, x2, y2) {
                this.x = x1; this.y = y1;
                const dx = x2 - x1;
                const dy = y2 - y1;
                const dist = Math.sqrt(dx*dx + dy*dy) || 1;
                this.vx = (dx / dist) * 8;
                this.vy = (dy / dist) * 8;
                this.life = 1.0;
                this.decay = 0.03 + Math.random() * 0.02;
                this.trail = [];
            }
            update() {
                this.trail.push({x: this.x, y: this.y});
                if (this.trail.length > 5) this.trail.shift();
                this.x += this.vx;
                this.y += this.vy;
                this.life -= this.decay;
                return this.life > 0 && this.x > -50 && this.x < W + 50 && this.y > -50 && this.y < H + 50;
            }
            draw() {
                ctx.save();
                // Trail
                for (let i = 0; i < this.trail.length; i++) {
                    const a = (i / this.trail.length) * this.life * 0.5;
                    ctx.fillStyle = 'rgba(255,200,50,' + a + ')';
                    ctx.beginPath();
                    ctx.arc(this.trail[i].x, this.trail[i].y, 1 + i * 0.3, 0, Math.PI*2);
                    ctx.fill();
                }
                // Bullet
                ctx.shadowColor = 'rgba(255,200,50,0.8)';
                ctx.shadowBlur = 15;
                ctx.fillStyle = 'rgba(255,255,200,' + this.life + ')';
                ctx.beginPath();
                ctx.arc(this.x, this.y, 2 + this.life, 0, Math.PI*2);
                ctx.fill();
                ctx.shadowBlur = 0;
                ctx.restore();
            }
        }

        // Laser class
        class Laser {"""

html = html.replace(old_laser, new_bullet)

# Add bullets to entities and handle machinegun fire
old_entities = """        const spiders = [];
        const lasers = [];
        const eyes = [];
        const skulls = [];
        for (let i = 0; i < 6; i++) skulls.push(new RobotSkull());
        for (let i = 0; i < 10; i++) spiders.push(new SpiderDrone());
        for (let i = 0; i < 30; i++) eyes.push(new TerminatorEye());"""

new_entities = """        const spiders = [];
        const lasers = [];
        const bullets = [];
        const eyes = [];
        const skulls = [];
        for (let i = 0; i < 6; i++) skulls.push(new RobotSkull());
        for (let i = 0; i < 8; i++) spiders.push(new SpiderDrone());
        for (let i = 0; i < 30; i++) eyes.push(new TerminatorEye());"""

html = html.replace(old_entities, new_entities)

# Handle machinegun fire in animate loop
old_spider_update = """            spiders.forEach(s => {
                const action = s.update();
                s.draw();
                if (action === 'fire') {
                    const targetX = Math.random() * W;
                    const targetY = Math.random() * H;
                    lasers.push(new Laser(s.x, s.y, targetX, targetY));
                }
            });"""

new_spider_update = """            spiders.forEach(s => {
                const action = s.update();
                s.draw();
                if (action === 'fire') {
                    const targetX = Math.random() * W;
                    const targetY = Math.random() * H;
                    lasers.push(new Laser(s.x, s.y, targetX, targetY));
                } else if (action === 'machinegun') {
                    for (let b = 0; b < 3; b++) {
                        const tx = s.x + (Math.random() - 0.5) * 200;
                        const ty = s.y + (Math.random() - 0.5) * 200;
                        bullets.push(new Bullet(s.x + s.size*0.8, s.y - s.size*0.45, tx, ty));
                    }
                }
            });
            // Update bullets
            for (let i = bullets.length - 1; i >= 0; i--) {
                if (!bullets[i].update()) {
                    bullets.splice(i, 1);
                } else {
                    bullets[i].draw();
                }
            }"""

html = html.replace(old_spider_update, new_spider_update)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Creepy machine-gun spiders deployed")
