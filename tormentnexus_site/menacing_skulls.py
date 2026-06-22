with open("/var/www/tormentnexus.site/index.html", "r") as f:
    html = f.read()

old_draw = """            draw() {
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
                const jaw = this.jawOpen * 0.8;
                ctx.fillStyle = 'rgba(0,100,150,' + (o * 0.25) + ')';
                ctx.beginPath(); ctx.arc(0, s*0.2, s*0.65, 0, Math.PI + jaw); ctx.fill();
                // Upper teeth
                ctx.fillStyle = 'rgba(0,240,255,' + (o * 0.35) + ')';
                for (let i = -3; i <= 3; i++) { if(i===0) continue; ctx.fillRect(i*s*0.12, -s*0.05, s*0.06, s*0.1); }
                // Lower teeth
                for (let i = -3; i <= 3; i++) { if(i===0) continue; ctx.fillRect(i*s*0.12, s*0.1+jaw*s*0.6, s*0.06, s*0.12); }
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
            }"""

new_draw = """            draw() {
                ctx.save();
                ctx.translate(this.x, this.y);
                const s = this.size;
                const o = this.opacity;
                const jaw = this.jawOpen * 1.0;
                ctx.shadowColor = 'rgba(0,240,255,' + (o * 0.15) + ')';
                ctx.shadowBlur = 40;
                // Elongated skull dome
                ctx.fillStyle = 'rgba(0,180,230,' + o + ')';
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.5) + ')';
                ctx.lineWidth = 1.5;
                ctx.beginPath();
                ctx.arc(0, -s*0.15, s*0.7, Math.PI*1.1, -Math.PI*0.1);
                ctx.lineTo(-s*0.5, s*0.1);
                ctx.lineTo(-s*0.55, s*0.05);
                ctx.closePath();
                ctx.fill();
                ctx.stroke();
                ctx.beginPath();
                ctx.moveTo(s*0.5, -s*0.1);
                ctx.lineTo(s*0.55, s*0.05);
                ctx.lineTo(s*0.5, s*0.1);
                ctx.closePath();
                ctx.fill();
                ctx.stroke();
                // Angry brow ridge
                ctx.fillStyle = 'rgba(0,0,0,0.6)';
                ctx.beginPath();
                ctx.moveTo(-s*0.5, -s*0.15);
                ctx.lineTo(-s*0.2, -s*0.2);
                ctx.lineTo(0, -s*0.15);
                ctx.lineTo(s*0.2, -s*0.2);
                ctx.lineTo(s*0.5, -s*0.15);
                ctx.lineTo(s*0.45, -s*0.05);
                ctx.lineTo(-s*0.45, -s*0.05);
                ctx.closePath();
                ctx.fill();
                // Sharp cheekbones
                ctx.fillStyle = 'rgba(0,160,220,' + (o * 0.35) + ')';
                ctx.beginPath();
                ctx.moveTo(-s*0.6, s*0.05);
                ctx.lineTo(-s*0.35, s*0.15);
                ctx.lineTo(-s*0.5, s*0.35);
                ctx.lineTo(-s*0.65, s*0.2);
                ctx.closePath();
                ctx.fill();
                ctx.beginPath();
                ctx.moveTo(s*0.6, s*0.05);
                ctx.lineTo(s*0.35, s*0.15);
                ctx.lineTo(s*0.5, s*0.35);
                ctx.lineTo(s*0.65, s*0.2);
                ctx.closePath();
                ctx.fill();
                // Gaping jaw
                ctx.fillStyle = 'rgba(0,80,130,' + (o * 0.2) + ')';
                ctx.beginPath();
                ctx.arc(0, s*0.25, s*0.7, 0.1, Math.PI*0.9 + jaw);
                ctx.lineTo(-s*0.55, s*0.15);
                ctx.closePath();
                ctx.fill();
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.2) + ')';
                ctx.lineWidth = 0.5;
                ctx.stroke();
                // Jagged upper teeth
                ctx.fillStyle = 'rgba(0,240,255,' + (o * 0.4) + ')';
                for (let i = -4; i <= 4; i++) {
                    if(i===0) continue;
                    const tx = i*s*0.1;
                    const tw = s*0.05 + (i%2===0?s*0.02:0);
                    const th = s*0.08 + Math.abs(i)*s*0.01;
                    ctx.beginPath();
                    ctx.moveTo(tx-tw/2, -s*0.02);
                    ctx.lineTo(tx, s*0.02+th);
                    ctx.lineTo(tx+tw/2, -s*0.02);
                    ctx.closePath();
                    ctx.fill();
                }
                // Jagged lower teeth
                ctx.fillStyle = 'rgba(200,255,255,' + (o * 0.3) + ')';
                for (let i = -4; i <= 4; i++) {
                    if(i===0) continue;
                    const tx = i*s*0.1;
                    const tw = s*0.05 + (i%2===0?s*0.02:0);
                    const th = s*0.08 + Math.abs(i)*s*0.01;
                    const ly = s*0.2 + jaw*s*0.5;
                    ctx.beginPath();
                    ctx.moveTo(tx-tw/2, ly+th);
                    ctx.lineTo(tx, ly);
                    ctx.lineTo(tx+tw/2, ly+th);
                    ctx.closePath();
                    ctx.fill();
                }
                // Deep tilted eye sockets
                ctx.fillStyle = 'rgba(0,0,0,0.9)';
                ctx.beginPath();
                ctx.ellipse(-s*0.25, -s*0.08, s*0.18, s*0.14, -0.15, 0, Math.PI*2);
                ctx.fill();
                ctx.beginPath();
                ctx.ellipse(s*0.25, -s*0.08, s*0.18, s*0.14, 0.15, 0, Math.PI*2);
                ctx.fill();
                // Intense red eyes
                ctx.shadowColor = 'rgba(255,0,0,' + this.eyeGlow + ')';
                ctx.shadowBlur = 30;
                ctx.fillStyle = 'rgba(255,0,30,' + this.eyeGlow + ')';
                ctx.beginPath();
                ctx.arc(-s*0.25, -s*0.08, s*0.07, 0, Math.PI*2);
                ctx.fill();
                ctx.beginPath();
                ctx.arc(s*0.25, -s*0.08, s*0.07, 0, Math.PI*2);
                ctx.fill();
                // Inner eye hot spot
                ctx.shadowBlur = 40;
                ctx.fillStyle = 'rgba(255,200,200,' + (this.eyeGlow * 0.5) + ')';
                ctx.beginPath();
                ctx.arc(-s*0.25, -s*0.08, s*0.025, 0, Math.PI*2);
                ctx.fill();
                ctx.beginPath();
                ctx.arc(s*0.25, -s*0.08, s*0.025, 0, Math.PI*2);
                ctx.fill();
                // Jagged nose cavity
                ctx.fillStyle = 'rgba(0,0,0,0.5)';
                ctx.beginPath();
                ctx.moveTo(0, s*0.01);
                ctx.lineTo(-s*0.06, s*0.14);
                ctx.lineTo(-s*0.02, s*0.1);
                ctx.lineTo(0, s*0.14);
                ctx.lineTo(s*0.02, s*0.1);
                ctx.lineTo(s*0.06, s*0.14);
                ctx.closePath();
                ctx.fill();
                // Battle crack lines
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.12) + ')';
                ctx.lineWidth = 0.5;
                ctx.beginPath();
                ctx.moveTo(-s*0.1, -s*0.5);
                ctx.lineTo(-s*0.05, -s*0.35);
                ctx.lineTo(-s*0.12, -s*0.25);
                ctx.stroke();
                ctx.beginPath();
                ctx.moveTo(s*0.15, -s*0.45);
                ctx.lineTo(s*0.1, -s*0.3);
                ctx.lineTo(s*0.18, -s*0.2);
                ctx.stroke();
                // Cybernetic bands
                ctx.strokeStyle = 'rgba(0,240,255,' + (o * 0.2) + ')';
                ctx.lineWidth = 1.5;
                ctx.beginPath();
                ctx.moveTo(-s*0.55, -s*0.3);
                ctx.lineTo(-s*0.2, -s*0.25);
                ctx.stroke();
                ctx.beginPath();
                ctx.moveTo(s*0.55, -s*0.3);
                ctx.lineTo(s*0.2, -s*0.25);
                ctx.stroke();
                // Rivets
                ctx.fillStyle = 'rgba(0,240,255,' + (o * 0.15) + ')';
                ctx.beginPath(); ctx.arc(-s*0.5, -s*0.3, s*0.02, 0, Math.PI*2); ctx.fill();
                ctx.beginPath(); ctx.arc(s*0.5, -s*0.3, s*0.02, 0, Math.PI*2); ctx.fill();
                ctx.shadowBlur = 0;
                ctx.restore();
            }"""

html = html.replace(old_draw, new_draw)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(html)
print("Menacing skulls deployed")
