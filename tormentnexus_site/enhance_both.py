# Both sites enhancement script

# ===== HYPERNEXUS — EVEN SHINIER =====
with open("/var/www/hypernexus.site/index.html", "r") as f:
    hn = f.read()

# Add reflective/glass gradient to hero section
hn = hn.replace(
    "background: var(--bg);\n        color: var(--text);\n        font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;\n        line-height: 1.6;\n        -webkit-font-smoothing: antialiased;\n        overflow-x: hidden;",
    """background: var(--bg);
        color: var(--text);
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        line-height: 1.6;
        -webkit-font-smoothing: antialiased;
        overflow-x: hidden;
        background-image: radial-gradient(ellipse at 20% 50%, rgba(77,124,255,0.03) 0%, transparent 50%),
                          radial-gradient(ellipse at 80% 50%, rgba(180,92,255,0.02) 0%, transparent 50%);""",
)

# Add shiny metallic gradient to stat numbers
hn = hn.replace(
    ".stat-num { font-size: 42px; font-weight: 800; background: var(--gradient-main); background-size: 300% 300%; animation: gradientShift 6s ease infinite; -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; margin-bottom: 6px; letter-spacing: -1px; filter: drop-shadow(0 0 20px var(--primary-glow)); }",
    ".stat-num { font-size: 42px; font-weight: 800; background: var(--gradient-main); background-size: 300% 300%; animation: gradientShift 6s ease infinite; -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; margin-bottom: 6px; letter-spacing: -1px; filter: drop-shadow(0 0 30px var(--primary-glow)) drop-shadow(0 0 60px var(--primary-glow)); }",
)

# Add shine overlay to feature cards on hover
hn = hn.replace(
    ".feature:hover { background: var(--bg-card-hover); border-color: var(--border-hover); transform: translateY(-6px) scale(1.02); box-shadow: var(--shadow-card); }",
    ".feature:hover { background: var(--bg-card-hover); border-color: var(--border-hover); transform: translateY(-6px) scale(1.02); box-shadow: var(--shadow-card), 0 0 80px var(--primary-glow); }\n.feature::before { content: ''; position: absolute; top: 0; left: -200%; width: 200%; height: 100%; background: linear-gradient(90deg, transparent, rgba(255,255,255,0.03), transparent); transform: skewX(-20deg); transition: left 0.8s; z-index: 0; }\n.feature:hover::before { left: 200%; }",
)

# Add smoother, more detailed particles
hn = hn.replace(
    "class Particle {",
    "class Particle { constructor() { this.alpha = Math.random(); this.alphaDir = 1; this.reset(); }",
)
hn = hn.replace(
    "this.opacity += (Math.random() - 0.5) * 0.01;\n                this.opacity = Math.max(0.05, Math.min(0.5, this.opacity));",
    "this.alpha += 0.01 * this.alphaDir; if(this.alpha > 1 || this.alpha < 0) this.alphaDir *= -1; this.opacity = 0.05 + this.alpha * 0.35;",
)

# Add glow border to all price cards on hover
hn = hn.replace(
    ".price-card:hover { background: var(--bg-card-hover); transform: translateY(-6px); }",
    ".price-card:hover { background: var(--bg-card-hover); transform: translateY(-6px); box-shadow: 0 0 60px var(--primary-glow); }",
)

# Add gradient underline to section headers
hn = hn.replace(
    ".section-header h2 { font-size: 36px; font-weight: 800; letter-spacing: -0.5px; margin-bottom: 12px; }",
    ".section-header h2 { font-size: 36px; font-weight: 800; letter-spacing: -0.5px; margin-bottom: 12px; display: inline-block; position: relative; }\n.section-header h2::after { content: ''; position: absolute; bottom: -4px; left: 50%; transform: translateX(-50%); width: 60px; height: 3px; background: var(--gradient-main); background-size: 300% 300%; animation: gradientShift 4s ease infinite; border-radius: 3px; }",
)

with open("/var/www/hypernexus.site/index.html", "w") as f:
    f.write(hn)
print("HyperNexus: enhanced shiny")

# ===== TORMENTNEXUS — SCARIER + SLEEKER =====
with open("/var/www/tormentnexus.site/index.html", "r") as f:
    tn = f.read()

# Add sleek carbon fiber texture to background
tn = tn.replace(
    "background:var(--abyss);\n  color:var(--text);\n  font-family:'Share Tech Mono','Courier New',monospace;",
    """background:var(--abyss);
  color:var(--text);
  font-family:'Share Tech Mono','Courier New',monospace;
  background-image: repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(255,255,255,0.003) 2px, rgba(255,255,255,0.003) 3px),
                    repeating-linear-gradient(90deg, transparent, transparent 2px, rgba(255,255,255,0.003) 2px, rgba(255,255,255,0.003) 3px);""",
)

# Add sleek gradient to hero text
tn = tn.replace(
    ".hero h1 .t1{display:block;color:var(--text);text-shadow:0 0 40px rgba(255,255,255,0.05)}",
    ".hero h1 .t1{display:block;background:linear-gradient(180deg,#fff 0%,var(--text) 60%,#666 100%);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;filter:drop-shadow(0 0 30px rgba(255,255,255,0.05))}",
)

# Add sleek glass-panel effect to cards
tn = tn.replace(
    ".card{background:rgba(5,0,0,0.4);border:1px solid rgba(200,0,0,0.06);padding:30px 28px;transition:all 0.4s;position:relative;overflow:hidden}",
    ".card{background:rgba(5,0,0,0.3);border:1px solid rgba(200,0,0,0.08);padding:30px 28px;transition:all 0.4s cubic-bezier(0.4,0,0.2,1);position:relative;overflow:hidden;backdrop-filter:blur(5px);-webkit-backdrop-filter:blur(5px)}",
)

# Add sleek gradient to card headers
tn = tn.replace(
    ".card h3{font-family:'Orbitron','Impact',sans-serif;font-size:16px;font-weight:900;margin-bottom:10px;color:#fff;text-transform:uppercase;letter-spacing:2px}",
    ".card h3{font-family:'Orbitron','Impact',sans-serif;font-size:16px;font-weight:900;margin-bottom:10px;background:linear-gradient(135deg,#fff 0%,#c0c0c0 50%,#888 100%);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;text-transform:uppercase;letter-spacing:2px}",
)

# Make stats sleeker
tn = tn.replace(
    ".stat{background:rgba(2,0,0,0.7);padding:20px 16px;text-align:center;position:relative;overflow:hidden}",
    ".stat{background:rgba(2,0,0,0.5);padding:20px 16px;text-align:center;position:relative;overflow:hidden;backdrop-filter:blur(4px);-webkit-backdrop-filter:blur(4px);border-bottom:1px solid rgba(200,0,0,0.05)}",
)

# Make stat numbers gradient
tn = tn.replace(
    ".stat-num{font-family:'Orbitron','Impact',sans-serif;font-size:24px;color:var(--fire);margin-bottom:2px;letter-spacing:2px;text-shadow:0 0 20px var(--eye-glow)}",
    ".stat-num{font-family:'Orbitron','Impact',sans-serif;font-size:26px;margin-bottom:2px;letter-spacing:2px;background:linear-gradient(180deg,var(--fire) 0%,var(--blood) 100%);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;filter:drop-shadow(0 0 20px var(--eye-glow)) drop-shadow(0 0 40px rgba(255,0,0,0.2))}",
)

# Make labels sleeker
tn = tn.replace(
    ".stat-label{font-size:7px;text-transform:uppercase;letter-spacing:3px;color:var(--text-dim);font-family:'Orbitron',monospace}",
    ".stat-label{font-size:7px;text-transform:uppercase;letter-spacing:4px;color:var(--text-dim);font-family:'Orbitron',monospace;opacity:0.6}",
)

# Add sleek glow to console
tn = tn.replace(
    ".console{margin:30px 0;background:rgba(0,0,0,0.6);border:1px solid rgba(200,0,0,0.1);padding:20px 24px;font-family:'Share Tech Mono','Courier New',monospace}",
    ".console{margin:30px 0;background:rgba(0,0,0,0.4);border:1px solid rgba(200,0,0,0.08);padding:20px 24px;font-family:'Share Tech Mono','Courier New',monospace;backdrop-filter:blur(10px);-webkit-backdrop-filter:blur(10px);box-shadow:0 0 30px rgba(200,0,0,0.03),inset 0 0 30px rgba(200,0,0,0.02)}",
)

# Add sleek footer
tn = tn.replace(
    "footer p{color:var(--text-dim);font-size:9px;text-transform:uppercase;letter-spacing:4px}",
    "footer p{color:var(--text-dim);font-size:9px;text-transform:uppercase;letter-spacing:4px;opacity:0.5}",
)

# Add creepier canvas - more eyes, more drones
tn = tn.replace(
    "for(let i=0;i<25;i++)eyes.push(new TEye())\nfor(let i=0;i<6;i++)drones.push(new Drone())",
    "for(let i=0;i<40;i++)eyes.push(new TEye())\nfor(let i=0;i<10;i++)drones.push(new Drone())",
)

# Make eyes more intense
tn = tn.replace(
    "this.size=4+Math.random()*6;this.dx=(Math.random()-0.5)*0.4;this.dy=(Math.random()-0.5)*0.4;this.pulse=0;this.lock=Math.random()>.8",
    "this.size=5+Math.random()*8;this.dx=(Math.random()-0.5)*0.6;this.dy=(Math.random()-0.5)*0.6;this.pulse=Math.random()*10;this.lock=false;this.twitchX=0;this.twitchY=0",
)

# Add eye twitching
tn = tn.replace(
    "this.x+=this.dx;this.y+=this.dy;this.pulse+=0.03;if(this.lock&&Math.random()<.01)this.lock=!this.lock;if(this.x<0||this.x>W||this.y<0||this.y>H){this.x=Math.random()*W;this.y=Math.random()*H}",
    "this.twitchX+=(Math.random()-0.5)*0.5;this.twitchY+=(Math.random()-0.5)*0.5;this.twitchX*=0.9;this.twitchY*=0.9;this.x+=this.dx+this.twitchX;this.y+=this.dy+this.twitchY;this.pulse+=0.04;if(this.x<-20||this.x>W+20||this.y<-20||this.y>H+20){this.x=Math.random()*W;this.y=Math.random()*H}",
)

# Add eye glow enhancement
tn = tn.replace(
    "ctx.shadowColor='rgba(255,0,0,'+g+')';ctx.shadowBlur=25\n    ctx.fillStyle='rgba(255,0,40,'+g+')'",
    "ctx.shadowColor='rgba(255,0,0,'+(g*0.8)+')';ctx.shadowBlur=35\n    ctx.fillStyle='rgba(255,0,40,'+(g*0.9)+')'\n    ctx.shadowColor='rgba(255,0,0,'+(g*0.4)+')';ctx.shadowBlur=50\n    ctx.fillStyle='rgba(255,0,40,'+(g*0.3)+')'\n    ctx.beginPath();ctx.arc(this.x,this.y,this.size*2,0,Math.PI*2);ctx.fill()",
)

# Make spiders creepier - add twitching
tn = tn.replace(
    "this.x+=dx/d*this.sp*(0.3+Math.random()*0.7);this.y+=dy/d*this.sp*(0.3+Math.random()*0.7)\n    this.angle=Math.atan2(dy,dx);this.lp+=0.05",
    "this.x+=dx/d*this.sp*(0.3+Math.random()*0.7);this.y+=dy/d*this.sp*(0.3+Math.random()*0.7)\n    this.angle=Math.atan2(dy,dx);this.lp+=0.05+Math.random()*0.03\n    // Random twitch\n    this.x+=(Math.random()-0.5)*2;this.y+=(Math.random()-0.5)*2",
)

# Make spider drawing more detailed
tn = tn.replace(
    "ctx.fillStyle='rgba(30,0,0,'+o+')';ctx.strokeStyle='rgba(200,0,0,'+(o*0.3)+')';ctx.lineWidth=0.5\n    ctx.beginPath();ctx.ellipse(0,0,s*0.35,s*0.45,0,0,Math.PI*2);ctx.fill();ctx.stroke()",
    "ctx.fillStyle='rgba(20,0,0,'+o+')';ctx.strokeStyle='rgba(200,0,0,'+(o*0.3)+')';ctx.lineWidth=0.5\n    ctx.beginPath();ctx.ellipse(0,0,s*0.35,s*0.45,0,0,Math.PI*2);ctx.fill();ctx.stroke()\n    // Abdomen segmentation\n    ctx.strokeStyle='rgba(200,0,0,'+(o*0.1)+')'\n    for(var seg=-2;seg<=2;seg++){ctx.beginPath();ctx.moveTo(-s*0.25,seg*s*0.1);ctx.lineTo(s*0.25,seg*s*0.1);ctx.stroke()}",
)

# Make drone legs creepier - more of them, hair-like
tn = tn.replace(
    "for(let i=0;i<8;i++){const a=i/8*Math.PI*2+this.lp;j=Math.sin(this.lp*1.5+i*0.8)*0.6",
    "for(var i=0;i<10;i++){var a=i/10*Math.PI*2+this.lp;var j=Math.sin(this.lp*1.5+i*0.8)*0.7",
)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(tn)
print("TormentNexus: scarier + sleeker")
