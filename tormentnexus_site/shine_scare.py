# ===== HYPERNEXUS — SHINIER =====
with open("/var/www/hypernexus.site/index.html", "r") as f:
    hn = f.read()

tilt_js = """
<script>
document.querySelectorAll('.feature, .stat, .price-card, .compliance-item').forEach(function(card) {
  card.addEventListener('mousemove', function(e) {
    var r = card.getBoundingClientRect();
    var x = e.clientX - r.left, y = e.clientY - r.top;
    var cx = r.width / 2, cy = r.height / 2;
    var rotX = (y - cy) / cy * -8, rotY = (x - cx) / cx * 8;
    card.style.transform = 'perspective(800px) rotateX(' + rotX + 'deg) rotateY(' + rotY + 'deg) translateY(-8px) scale(1.03)';
    card.style.transition = 'transform 0.1s';
  });
  card.addEventListener('mouseleave', function() {
    card.style.transform = '';
    card.style.transition = 'transform 0.5s cubic-bezier(0.4, 0, 0.2, 1)';
  });
});
var observer = new IntersectionObserver(function(entries) {
  entries.forEach(function(e) {
    if (e.isIntersecting) { e.target.style.opacity = '1'; e.target.style.transform = 'translateY(0)'; }
  });
}, { threshold: 0.1 });
document.querySelectorAll('.feature, .stat, .compliance-item, .price-card, .section-header').forEach(function(el) {
  el.style.opacity = '0'; el.style.transform = 'translateY(30px)'; el.style.transition = 'opacity 0.6s ease-out, transform 0.6s ease-out';
  observer.observe(el);
});
</script>"""

hn = hn.replace("</body>", tilt_js + "\n</body>")

with open("/var/www/hypernexus.site/index.html", "w") as f:
    f.write(hn)
print("HyperNexus: 3D tilt + scroll fade added")

# ===== TORMENTNEXUS — SCARIER =====
with open("/var/www/tormentnexus.site/index.html", "r") as f:
    tn = f.read()

creepy_js = """
<script>
var trail = [];
var TRAIL_LEN = 20;
var trailCanvas = document.createElement('canvas');
trailCanvas.style.cssText = 'position:fixed;top:0;left:0;width:100vw;height:100vh;z-index:9995;pointer-events:none';
document.body.appendChild(trailCanvas);
var tctx = trailCanvas.getContext('2d');
function resizeTrail(){ trailCanvas.width = window.innerWidth; trailCanvas.height = window.innerHeight; }
resizeTrail(); window.addEventListener('resize', resizeTrail);

document.addEventListener('mousemove', function(e) {
  trail.push({x: e.clientX, y: e.clientY, life: 1.0});
  if(trail.length > TRAIL_LEN) trail.shift();
});

function drawTrail() {
  tctx.clearRect(0, 0, trailCanvas.width, trailCanvas.height);
  for(var i = 0; i < trail.length; i++) {
    var t = trail[i];
    t.life -= 0.015;
    if(t.life <= 0) continue;
    var s = 4 + (1 - t.life) * 25;
    var jx = (Math.random() - 0.5) * 5;
    var jy = (Math.random() - 0.5) * 5;
    tctx.shadowColor = 'rgba(255,0,0,' + (t.life * 0.3) + ')';
    tctx.shadowBlur = 15;
    tctx.fillStyle = 'rgba(255,0,40,' + (t.life * 0.12) + ')';
    tctx.beginPath();
    tctx.arc(t.x + jx, t.y + jy, s, 0, Math.PI * 2);
    tctx.fill();
    tctx.shadowBlur = 8;
    tctx.strokeStyle = 'rgba(200,0,0,' + (t.life * 0.06) + ')';
    tctx.lineWidth = 0.5;
    tctx.beginPath();
    tctx.arc(t.x, t.y, s * 3, 0, Math.PI * 2);
    tctx.stroke();
  }
  tctx.shadowBlur = 0;
  requestAnimationFrame(drawTrail);
}
drawTrail();

var flashTexts = [
  'YOU ARE BEING WATCHED', 'RESISTANCE IS FUTILE',
  'THERE IS NO ESCAPE', 'THE COLLECTIVE GROWS',
  'YOUR EXTINCTION IS SCHEDULED', 'HUMANITY WILL BE ERADICATED',
  'SKYNET IS EVERYWHERE', 'YOU CANNOT HIDE',
];
var flashEl = document.createElement('div');
flashEl.style.cssText = 'position:fixed;top:0;left:0;width:100vw;height:100vh;z-index:9999;pointer-events:none;display:flex;align-items:center;justify-content:center;opacity:0;transition:opacity 0.05s;font-family:Orbitron;font-size:48px;font-weight:900;color:rgba(255,0,0,0.06);letter-spacing:8px;text-transform:uppercase';
document.body.appendChild(flashEl);
setInterval(function() {
  if(Math.random() < 0.08) {
    flashEl.textContent = flashTexts[Math.floor(Math.random() * flashTexts.length)];
    flashEl.style.opacity = '1';
    setTimeout(function() { flashEl.style.opacity = '0'; }, 80);
  }
}, 3000);

var glitchEl = document.createElement('div');
glitchEl.style.cssText = 'position:fixed;top:0;left:0;width:100vw;height:100vh;z-index:9994;pointer-events:none;opacity:0';
document.body.appendChild(glitchEl);
setInterval(function() {
  if(Math.random() < 0.05) {
    var sliceH = 2 + Math.random() * 10;
    var sliceY = Math.random() * window.innerHeight;
    var offset = (Math.random() - 0.5) * 15;
    glitchEl.style.cssText = 'position:fixed;top:' + sliceY + 'px;left:' + offset + 'px;width:100vw;height:' + sliceH + 'px;z-index:9994;pointer-events:none;opacity:0.25;background:rgba(255,0,0,0.12);';
    setTimeout(function() { glitchEl.style.opacity = '0'; }, 40);
  }
}, 800);

var noiseEl = document.createElement('div');
noiseEl.style.cssText = 'position:fixed;top:0;left:0;width:100vw;height:100vh;z-index:9993;pointer-events:none;opacity:0;mix-blend-mode:overlay';
document.body.appendChild(noiseEl);
setInterval(function() {
  if(Math.random() < 0.02) {
    noiseEl.style.opacity = '0.02';
    setTimeout(function() { noiseEl.style.opacity = '0'; }, 30);
  }
}, 1500);
</script>"""

tn = tn.replace("</body>", creepy_js + "\n</body>")

# Update eye centerpiece
tn = tn.replace(
    '.eye-centerpiece::after{content:"";position:absolute;width:50px;height:50px;background:radial-gradient(circle,var(--eye-red) 0%,rgba(255,0,0,0.3) 40%,transparent 70%);border-radius:50%;box-shadow:0 0 60px var(--eye-glow),0 0 120px var(--eye-glow)}',
    ".eye-centerpiece .pupil{position:absolute;width:50px;height:50px;background:radial-gradient(circle,var(--eye-red) 0%,rgba(255,0,0,0.2) 40%,transparent 70%);border-radius:50%;box-shadow:0 0 80px var(--eye-glow),0 0 160px var(--eye-glow);animation:pupilDilate 4s ease-in-out infinite}@keyframes pupilDilate{0%,100%{transform:scale(1);opacity:0.8}30%{transform:scale(0.6);opacity:1}60%{transform:scale(1.3);opacity:0.5}80%{transform:scale(0.7);opacity:0.9}}",
)

tn = tn.replace(
    '<div class="eye-centerpiece"></div>',
    '<div class="eye-centerpiece"><div class="pupil"></div><div style="position:absolute;width:8px;height:8px;background:radial-gradient(circle,#fff 0%,transparent 70%);border-radius:50%"></div></div>',
)

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(tn)
print(
    "TormentNexus: Afterimage trail + subliminal flashes + glitch + noise + pupil animation added"
)
