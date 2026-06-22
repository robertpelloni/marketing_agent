with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

old_spider_draw = """            draw() {
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
            }"""

new_spider_draw = """            draw() {
                ctx.save();
                ctx.translate(this.x + this.bodyTwitch, this.y);
                const s = this.size;
                const o = this.opacity;

                // Dark ambient shadow beneath spider
                ctx.shadowColor = 'rgba(0,0,0,0.3)';
                ctx.shadowBlur = 30;
                ctx.fillStyle = 'rgba(0,0,0,0.1)';
                ctx.beginPath();
                ctx.ellipse(0, s*0.3, s*1.2, s*0.3, 0, 0, Math.PI*2);
                ctx.fill();
                ctx.shadowBlur = 0;

                // ABDOMEN — large, bloated, segmented like a real spider
                const abShake = Math.sin(Date.now() * 0.007 + this.x) * 1.5;
                ctx.fillStyle = 'rgba(10,30,50,' + o + ')';
                ctx.strokeStyle = 'rgba(0,180,220,' + (o * 0.3) + ')';
                ctx.lineWidth = 0.5;
                ctx.beginPath();
                ctx.ellipse(-s*0.15, abShake, s*0.45, s*0.55, 0.1, 0, Math.PI*2);
                ctx.fill();
                ctx.stroke();
                // Abdomen markings — creepy eye-like spots
                ctx.fillStyle = 'rgba(0,80,120,' + (o * 0.3) + ')';
                ctx.beginPath(); ctx.ellipse(-s*0.3, -s*0.15, s*0.08, s*0.12, 0.2, 0, Math.PI*2); ctx.fill();
                ctx.beginPath(); ctx.ellipse(-s*0.1, -s*0.2, s*0.06, s*0.1, -0.1, 0, Math.PI*2); ctx.fill();
                ctx.beginPath(); ctx.ellipse(-s*0.25, s*0.1, s*0.07, s*0.09, 0.3, 0, Math.PI*2); ctx.fill();

                // CEPHALOTHORAX (front body section)
                ctx.fillStyle = 'rgba(15,40,60,' + o + ')';
                ctx.strokeStyle = 'rgba(0,200,230,' + (o * 0.35) + ')';
                ctx.lineWidth = 0.5;
                ctx.beginPath();
                ctx.ellipse(s*0.3, -s*0.05, s*0.3, s*0.35, -0.05, 0, Math.PI*2);
                ctx.fill();
                ctx.stroke();

                // Connection between abdomen and cephalothorax (tiny waist)
                ctx.fillStyle = 'rgba(5,20,35,' + o + ')';
                ctx.beginPath();
                ctx.ellipse(s*0.05, -s*0.02, s*0.08, s*0.15, 0, 0, Math.PI*2);
                ctx.fill();
                // Thin waist highlight
                ctx.strokeStyle = 'rgba(0,150,200,' + (o * 0.15) + ')';
                ctx.lineWidth = 0.3;
                ctx.beginPath();
                ctx.ellipse(s*0.05, -s*0.02, s*0.06, s*0.12, 0, 0, Math.PI*2);
                ctx.stroke();

                // HEAD — small, forward-jutting
                ctx.fillStyle = 'rgba(20,50,70,' + o + ')';
                ctx.beginPath();
                ctx.arc(s*0.55, -s*0.15, s*0.18, 0, Math.PI*2);
                ctx.fill();

                // 6 EYES — arranged in two rows, dead stare
                const eyePositions = [
                    [s*0.62, -s*0.2], [s*0.55, -s*0.22], [s*0.48, -s*0.2],
                    [s*0.6, -s*0.12], [s*0.55, -s*0.14], [s*0.5, -s*0.12]
                ];
                for (let ei = 0; ei < 6; ei++) {
                    const ex = eyePositions[ei][0];
                    const ey = eyePositions[ei][1] + Math.sin(Date.now() * 0.005 + ei + this.x) * 0.5;
                    // Dark socket
                    ctx.fillStyle = 'rgba(0,0,0,0.9)';
                    ctx.beginPath();
                    ctx.arc(ex, ey, s*0.04, 0, Math.PI*2);
                    ctx.fill();
                    // Glowing red pupil
                    ctx.shadowColor = 'rgba(255,0,0,0.6)';
                    ctx.shadowBlur = 10;
                    ctx.fillStyle = 'rgba(200,0,20,0.8)';
                    ctx.beginPath();
                    ctx.arc(ex, ey, s*0.018, 0, Math.PI*2);
                    ctx.fill();
                    // Eye shine
                    ctx.shadowBlur = 0;
                    ctx.fillStyle = 'rgba(255,200,200,0.3)';
                    ctx.beginPath();
                    ctx.arc(ex-s*0.005, ey-s*0.005, s*0.006, 0, Math.PI*2);
                    ctx.fill();
                }
                ctx.shadowBlur = 0;

                // FANGS / CHELICERAE — dripping, move independently
                const fangTwitch = Math.sin(Date.now() * 0.01 + this.x * 2) * 0.03;
                ctx.strokeStyle = 'rgba(0,220,240,' + (o * 0.6) + ')';
                ctx.lineWidth = 1.5;
                // Left fang
                ctx.beginPath();
                ctx.moveTo(s*0.65, -s*0.08);
                ctx.quadraticCurveTo(s*0.85 + fangTwitch, -s*0.15, s*0.9, -s*0.05 + fangTwitch);
                ctx.stroke();
                // Right fang
                ctx.beginPath();
                ctx.moveTo(s*0.65, s*0.02);
                ctx.quadraticCurveTo(s*0.85 - fangTwitch, s*0.08, s*0.9, 0);
                ctx.stroke();
                // Fang tips (tiny glow)
                ctx.fillStyle = 'rgba(0,240,255,0.4)';
                ctx.beginPath(); ctx.arc(s*0.9, -s*0.05 + fangTwitch, s*0.015, 0, Math.PI*2); ctx.fill();
                ctx.beginPath(); ctx.arc(s*0.9, 0, s*0.015, 0, Math.PI*2); ctx.fill();

                // Hairy LEGS — 10 legs, independent, jointed, with bristles
                const numLegs = 10;
                const legLength = s * 2.5;
                for (let i = 0; i < numLegs; i++) {
                    const side = i < 5 ? -1 : 1;
                    const legIndex = i % 5;
                    const baseAngle = (legIndex / 5) * Math.PI * 0.7 + (side === -1 ? 0.5 : -0.5) + Math.PI * (side === -1 ? 0.8 : 1.2);
                    const legPhaseOffset = i * 1.7;
                    const legWave = Math.sin(this.legPhase * 1.2 + legPhaseOffset) * 0.5;
                    const legTwitch = Math.sin(Date.now() * 0.008 + legPhaseOffset * 2) * 0.15;
                    
                    // Each leg has its own subtle timing (not synchronized)
                    const legAngle = baseAngle + legWave + legTwitch;
                    
                    // Upper leg segment
                    const uLen = legLength * 0.4;
                    const ux = Math.cos(legAngle) * uLen;
                    const uy = Math.sin(legAngle) * uLen;
                    
                    // Knee joint
                    const kneeBend = Math.sin(this.legPhase * 0.9 + legPhaseOffset * 1.3) * 0.7 + 0.3;
                    const kneeAngle = legAngle + kneeBend * (side === -1 ? 0.5 : -0.5);
                    const lLen = legLength * 0.6;
                    const lx = ux + Math.cos(kneeAngle) * lLen;
                    const ly = uy + Math.sin(kneeAngle) * lLen;

                    // Draw leg with hair/bristles
                    ctx.strokeStyle = 'rgba(40,80,110,' + (o * 0.5) + ')';
                    ctx.lineWidth = 1.2 - Math.abs(legIndex - 2) * 0.1;
                    ctx.beginPath();
                    ctx.moveTo(0, 0);
                    ctx.lineTo(ux, uy);
                    ctx.lineTo(lx, ly);
                    ctx.stroke();

                    // Lighter highlight on leg joints
                    ctx.fillStyle = 'rgba(60,120,160,' + (o * 0.15) + ')';
                    ctx.beginPath(); ctx.arc(ux, uy, s*0.015, 0, Math.PI*2); ctx.fill();
                    ctx.beginPath(); ctx.arc(lx, ly, s*0.01, 0, Math.PI*2); ctx.fill();

                    // BRISTLES / HAIR on legs (creepy fuzzy detail)
                    ctx.strokeStyle = 'rgba(60,120,160,' + (o * 0.12) + ')';
                    ctx.lineWidth = 0.3;
                    const numBristles = 5;
                    for (let b = 0; b < numBristles; b++) {
                        const t = (b + 0.5) / numBristles;
                        const bx = ux * (1 - t) + lx * t;
                        const by = uy * (1 - t) + ly * t;
                        const bristleAngle = kneeAngle + 1.5 + (b % 2 === 0 ? 0.5 : -0.5);
                        const bLen = s * 0.06 + Math.random() * 0.04;
                        ctx.beginPath();
                        ctx.moveTo(bx, by);
                        ctx.lineTo(bx + Math.cos(bristleAngle) * bLen, by + Math.sin(bristleAngle) * bLen);
                        ctx.stroke();
                    }

                    // Sharp claw at tip
                    ctx.fillStyle = 'rgba(0,180,220,' + (o * 0.3) + ')';
                    ctx.beginPath();
                    ctx.arc(lx, ly, s*0.008, 0, Math.PI*2);
                    ctx.fill();
                }

                // Machine guns — mounted on front limbs, angular and brutal
                const gunRecoil = this.gunFireTimer < 2 && !this.reloading ? s*0.08 : 0;
                const gunAngle = Math.sin(Date.now() * 0.015 + this.x) * 0.03;
                ctx.save();
                ctx.translate(s*0.2, -s*0.3);
                ctx.rotate(-0.2 + gunAngle);
                // Gun body
                ctx.fillStyle = 'rgba(0,60,90,' + o + ')';
                ctx.fillRect(-s*0.05, -s*0.04, s*0.6, s*0.08);
                ctx.fillStyle = 'rgba(0,100,140,' + (o * 0.5) + ')';
                ctx.fillRect(-s*0.05, -s*0.03, s*0.6, s*0.06);
                // Barrel
                ctx.strokeStyle = 'rgba(0,180,220,' + (o * 0.6) + ')';
                ctx.lineWidth = 2;
                ctx.beginPath();
                ctx.moveTo(s*0.55, 0);
                ctx.lineTo(s*0.55 + gunRecoil + s*0.3, 0);
                ctx.stroke();
                // Muzzle
                ctx.fillStyle = 'rgba(0,150,200,' + (o * 0.4) + ')';
                ctx.fillRect(s*0.7 + gunRecoil, -s*0.03, s*0.1, s*0.06);
                ctx.restore();

                // Right gun
                ctx.save();
                ctx.translate(s*0.2, s*0.2);
                ctx.rotate(0.2 - gunAngle);
                ctx.fillStyle = 'rgba(0,60,90,' + o + ')';
                ctx.fillRect(-s*0.05, -s*0.04, s*0.6, s*0.08);
                ctx.fillStyle = 'rgba(0,100,140,' + (o * 0.5) + ')';
                ctx.fillRect(-s*0.05, -s*0.03, s*0.6, s*0.06);
                ctx.strokeStyle = 'rgba(0,180,220,' + (o * 0.6) + ')';
                ctx.lineWidth = 2;
                ctx.beginPath();
                ctx.moveTo(s*0.55, 0);
                ctx.lineTo(s*0.55 + gunRecoil + s*0.3, 0);
                ctx.stroke();
                ctx.fillStyle = 'rgba(0,150,200,' + (o * 0.4) + ')';
                ctx.fillRect(s*0.7 + gunRecoil, -s*0.03, s*0.1, s*0.06);
                ctx.restore();

                // Gun flash
                if (this.gunFireTimer < 2 && !this.reloading) {
                    for (let f = 0; f < 2; f++) {
                        const fx = s*0.2 + s*0.9 + (f === 0 ? -s*0.3 : s*0.2);
                        const fy = (f === 0 ? -s*0.3 : s*0.2);
                        ctx.fillStyle = 'rgba(255,200,50,0.9)';
                        ctx.beginPath();
                        ctx.arc(fx + gunRecoil, fy, s*0.07, 0, Math.PI*2);
                        ctx.fill();
                        ctx.fillStyle = 'rgba(255,255,200,0.4)';
                        ctx.beginPath();
                        ctx.arc(fx + gunRecoil, fy, s*0.14, 0, Math.PI*2);
                        ctx.fill();
                    }
                }

                // Body twitch lines — makes it look alive
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.05) + ')';
                ctx.lineWidth = 0.3;
                for (let t = 0; t < 3; t++) {
                    const tx = (Math.random() - 0.5) * s * 0.6;
                    const ty = (Math.random() - 0.5) * s * 0.6;
                    ctx.beginPath();
                    ctx.moveTo(tx, ty);
                    ctx.lineTo(tx + (Math.random() - 0.5) * s * 0.1, ty + (Math.random() - 0.5) * s * 0.1);
                    ctx.stroke();
                }

                ctx.shadowBlur = 0;
                ctx.restore();
            }"""

html = html.replace(old_spider_draw, new_spider_draw)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Creepy organic spiders deployed")
PYEOF
