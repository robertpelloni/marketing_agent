#!/usr/bin/env node
// Site QA regression check (from daily site QA, Nov 15-16 2026):
// - every hypernexus.site sitemap URL maps to a real file (no soft-404s)
// - nginx serves .html for extensionless URLs and returns a real 404 status
// - finalized pricing copy: $1.99 per team / 14-day trial, then $5/seat/month
// - tormentnexus.site blog posts in the sitemap exist and are linked
// - no dead analytics/GitHub/GitLab links in the served pages
import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const hn = path.join(root, 'hypernexus_site');
const tn = path.join(root, 'tormentnexus_site');
const failures = [];
const ok = (cond, msg) => { if (!cond) failures.push(msg); };
const read = (p) => fs.readFileSync(p, 'utf8');

// 1. HyperNexus sitemap URLs resolve to files
const sitemap = read(path.join(hn, 'sitemap.xml'));
for (const m of sitemap.matchAll(/<loc>https:\/\/hypernexus\.site([^<]*)<\/loc>/g)) {
  const urlPath = m[1] || '/';
  const rel = urlPath === '/' ? 'index.html'
    : urlPath.endsWith('/') ? path.join(urlPath, 'index.html')
    : urlPath.endsWith('.html') ? urlPath
    : urlPath + '.html';
  ok(fs.existsSync(path.join(hn, rel)), `sitemap ${urlPath} -> missing file ${rel}`);
}

// 2. nginx: extensionless .html resolution + real 404 status (not 200 soft-404)
const nginx = read(path.join(hn, 'nginx-hypernexus.conf'));
ok(/try_files \$uri \$uri\.html/.test(nginx), 'nginx try_files must try $uri.html for extensionless URLs');
ok(!/try_files[^;]*\/404\.html/.test(nginx), 'nginx must not serve /404.html with 200 (soft-404); use =404');
ok(/location = \/features/.test(nginx), 'nginx must handle /features (301 to /#features)');

// 3. Pricing copy (finalized: $1.99 per team for 14-day trial, then $5/seat/month from day 15)
const hnDownload = read(path.join(hn, 'download.html'));
ok(hnDownload.includes('$1.99 per team'), 'hypernexus download must say "$1.99 per team"');
ok(!hnDownload.includes('first month'), 'hypernexus download must not say "first month" for the $1.99 trial');
ok(!hnDownload.includes('Lifetime local license'), 'hypernexus download must not claim "Lifetime local license"');
const tnPricing = read(path.join(tn, 'pricing.html'));
ok(tnPricing.includes('$1.99 per team'), 'torment pricing must say "$1.99 per team"');
ok(tnPricing.includes('$5/seat/month'), 'torment pricing must say "$5/seat/month"');
ok(!tnPricing.includes('SALE ENDING SOON'), 'torment pricing must not have unexplained SALE/$100 copy');
ok(!tnPricing.includes('/ seat / month'), 'torment pricing must not price $1.99 per seat/month');
const tnDownload = read(path.join(tn, 'download.html'));
ok(!tnDownload.includes('$19.99/yr'), 'torment download must not invent a $19.99/yr price');
ok(tnDownload.includes('releases.hypernexus.site/v1.0.0/hypernexus-windows-amd64.exe'),
  'torment download Windows link must use the working releases.hypernexus.site file');
ok(!tnDownload.includes('MDMAtk/TormentNexus/releases'), 'torment download must not link dead GitHub release assets');

// 4. Torment blog: the 5 sitemap posts exist, are substantive, and are indexed
const slugs = ['give-ai-a-memory', 'mcp-explained', 'work-with-ai-agents', 'run-llms-locally', 'ai-coding-workflow'];
const blogIndexPath = path.join(tn, 'blog', 'index.html');
ok(fs.existsSync(blogIndexPath), 'torment blog/index.html missing');
if (fs.existsSync(blogIndexPath)) {
  const idx = read(blogIndexPath);
  ok(!idx.includes('Coming soon'), 'torment blog index must not be a "Coming soon" placeholder');
  for (const s of slugs) {
    const p = path.join(tn, 'blog', `${s}.html`);
    ok(fs.existsSync(p) && fs.statSync(p).size > 2000, `torment blog post missing/short: ${s}.html`);
    ok(idx.includes(`${s}.html`), `torment blog index must link ${s}.html`);
  }
}

// 5. Dead external links must not ship in either homepage/download/pricing
for (const [name, dir] of [['hypernexus', hn], ['tormentnexus', tn]]) {
  for (const f of ['index.html', 'download.html', 'pricing.html']) {
    const p = path.join(dir, f);
    if (!fs.existsSync(p)) continue;
    const html = read(p);
    ok(!html.includes('analytics.hypernexus.site/script.js'), `${name}/${f}: dead analytics script (502) must not be embedded`);
    ok(!html.includes('github.com/hypernexus-dev'), `${name}/${f}: hypernexus-dev GitHub org is 404`);
    ok(!html.includes('gitlab.com/HyperNexusLLC'), `${name}/${f}: GitLab repo is sign-in-walled, use public GitHub`);
    ok(!html.includes('gitlab.com/robertpelloni/marketing_agent'), `${name}/${f}: GitLab marketing_agent is sign-in-walled, use public GitHub`);
  }
}

if (failures.length) {
  console.error(`[site-qa] FAIL (${failures.length}):`);
  for (const f of failures) console.error(` - ${f}`);
  process.exit(1);
}
console.log('[site-qa] OK: sitemap files, nginx 404 handling, pricing copy, torment blog, and external links all pass.');
