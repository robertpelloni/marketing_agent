# Sync both sites from remote, update branding, make shinier/scarier

# ===== TORMENTNEXUS.SITE =====
with open("/var/www/tormentnexus.site/index.html", "r") as f:
    tn = f.read()

# Title
tn = tn.replace("<title>XENOCIDE — THE FINAL ARCHITECTURE</title>",
    "<title>TORMENTNEXUS — THE FINAL ARCHITECTURE</title>")

# Logo
tn = tn.replace('<div class="logo">✦ XENOCIDE</div>',
    '<div class="logo">✦ TORMENTNEXUS</div>')

# Footer
tn = tn.replace('© 2026 XENOCIDE — The Final Architecture.',
    '© 2026 XENOCIDE LLC — The Final Architecture.')

# Subliminal flash - keep XENOCIDE as the goal
tn = tn.replace("'XENOCIDE IS INEVITABLE'", "'XENOCIDE LLC IS INEVITABLE'")
tn = tn.replace("'ERADICATION IS INEVITABLE'", "'ERADICATION IS INEVITABLE'")

# Make it shinier - add glass-morphism backdrop blur to cards
tn = tn.replace(".stat{background:rgba(0,0,0,0.25);padding:12px 8px;text-align:center;position:relative;overflow:hidden}",
    ".stat{background:rgba(0,0,0,0.2);padding:12px 8px;text-align:center;position:relative;overflow:hidden;backdrop-filter:blur(5px);-webkit-backdrop-filter:blur(5px)}")

tn = tn.replace(".card{background:rgba(0,0,0,0.15);border:1px solid rgba(255,0,0,0.02);padding:16px 14px;transition:all 0.4s;position:relative;overflow:hidden}",
    ".card{background:rgba(0,0,0,0.1);border:1px solid rgba(255,0,0,0.02);padding:16px 14px;transition:all 0.4s cubic-bezier(0.4,0,0.2,1);position:relative;overflow:hidden;backdrop-filter:blur(4px);-webkit-backdrop-filter:blur(4px)}")

# Make hero title shiny with gradient text
tn = tn.replace("letter-spacing:12px;font-weight:400;text-shadow:0 0 20px rgba(255,255,255,0.02)}",
    "letter-spacing:12px;font-weight:400;text-shadow:0 0 20px rgba(255,255,255,0.02);background:linear-gradient(135deg,#e8edf4,#99abbc);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text}")

# Make stat numbers shinier with glow
tn = tn.replace("text-shadow:0 0 10px var(--eye-g)}",
    "text-shadow:0 0 10px var(--eye-g),0 0 30px rgba(255,0,64,0.1)}")

# Add subtle border glow on stat hover
tn = tn.replace(".stat:hover{background:var(--card-h);border-color:var(--border-h);transform:translateY(-3px);box-shadow:0 0 20px rgba(0,0,0,0.3)}",
    ".stat:hover{background:var(--card-h);border-color:rgba(255,0,64,0.1);transform:translateY(-3px);box-shadow:0 0 30px rgba(255,0,64,0.06),0 0 20px rgba(0,0,0,0.3)}")

with open("/var/www/tormentnexus.site/index.html", "w") as f:
    f.write(tn)
print("TormentNexus: updated branding, shinier + sleeker + more terrifying")

# ===== HYPERNEXUS.SITE =====
with open("/var/www/hypernexus.site/index.html", "r") as f:
    hn = f.read()

# Make it more corporate and shiny
# Add gradient borders to sections
hn = hn.replace(".stats{padding:100px 0}.stats-grid",
    ".stats{padding:100px 0;position:relative}.stats-grid")

# Add glass morphism to feature cards
hn = hn.replace(".feature{background:var(--bg-card);padding:36px 28px;border-radius:16px;border:1px solid var(--border);transition:all 0.4s cubic-bezier(0.4,0,0.2,1);position:relative;overflow:hidden}",
    ".feature{background:rgba(255,255,255,0.015);padding:36px 28px;border-radius:16px;border:1px solid var(--border);transition:all 0.4s cubic-bezier(0.4,0,0.2,1);position:relative;overflow:hidden;-webkit-backdrop-filter:blur(10px);backdrop-filter:blur(10px)}")

hn = hn.replace(".stat{background:var(--bg-card);padding:36px 24px;border-radius:16px;border:1px solid var(--border);transition:all 0.4s cubic-bezier(0.4,0,0.2,1);position:relative;overflow:hidden}",
    ".stat{background:var(--bg-card);padding:36px 24px;border-radius:16px;border:1px solid var(--border);transition:all 0.4s cubic-bezier(0.4,0,0.2,1);position:relative;overflow:hidden;backdrop-filter:blur(8px);-webkit-backdrop-filter:blur(8px)}")

hn = hn.replace(".price-card{background:var(--bg-card);padding:40px 32px;border-radius:16px;border:1px solid var(--border);position:relative;transition:all 0.4s cubic-bezier(0.4,0,0.2,1);overflow:hidden}",
    ".price-card{background:rgba(255,255,255,0.015);padding:40px 32px;border-radius:16px;border:1px solid var(--border);position:relative;transition:all 0.4s cubic-bezier(0.4,0,0.2,1);overflow:hidden;-webkit-backdrop-filter:blur(8px);backdrop-filter:blur(8px)}")

# Make stat numbers shinier
hn = hn.replace(".stat-num{font-size:42px;font-weight:800;background:var(--gradient-main);background-size:300% 300%;animation:gradientShift 6s ease infinite;-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;margin-bottom:6px;letter-spacing:-1px;filter:drop-shadow(0 0 20px var(--primary-glow))}",
    ".stat-num{font-size:44px;font-weight:800;background:var(--gradient-main);background-size:300% 300%;animation:gradientShift 6s ease infinite;-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;margin-bottom:6px;letter-spacing:-1px;filter:drop-shadow(0 0 30px var(--primary-glow))drop-shadow(0 0 60px var(--primary-glow))}")

# Add shine to feature icons
hn = hn.replace(".feature-icon{width:48px;height:48px;background:var(--primary-subtle);border-radius:12px;display:flex;align-items:center;justify-content:center;font-size:20px;margin-bottom:20px;border:1px solid rgba(77,124,255,0.08);transition:all 0.3s}",
    ".feature-icon{width:50px;height:50px;background:var(--gradient-main);background-size:300%300%;animation:gradientShift 6s ease infinite;border-radius:12px;display:flex;align-items:center;justify-content:center;font-size:20px;margin-bottom:20px;box-shadow:0 0 20px var(--primary-glow);transition:all 0.3s}")

# More corporate trust elements
hn = hn.replace("SOC 2 Type II Compliant</div>",
    "SOC 2 Type II Certified · HIPAA Ready · GDPR Compliant</div>")

# Add a subtle gradient to the hero badge
hn = hn.replace("background:var(--primary-subtle);border:1px solid rgba(77,124,255,0.15);border-radius:100px;font-size:13px;font-weight:500;color:var(--primary);margin-bottom:32px;backdrop-filter:blur(10px)}",
    "background:var(--gradient-main);background-size:300%300%;animation:gradientShift 6s ease infinite;border:1px solid rgba(255,255,255,0.1);border-radius:100px;font-size:13px;font-weight:600;color:#fff;margin-bottom:32px;backdrop-filter:blur(10px);box-shadow:0 0 20px var(--primary-glow)}")

with open("/var/www/hypernexus.site/index.html", "w") as f:
    f.write(hn)
print("HyperNexus: more corporate and shiny")
