#!/usr/bin/env python3
"""Patch the live hypernexus.site nginx config: extensionless .html + real 404."""
import sys

p = '/etc/nginx/sites-enabled/hypernexus.site'
s = open(p).read()
old = 'try_files $uri $uri/ /404.html;'
if old not in s:
    print('already patched or pattern absent')
    sys.exit(0)
s = s.replace(old, 'error_page 404 /404.html;\n        try_files $uri $uri.html $uri/ =404;')
if '/features' not in s:
    s = s.replace('location = /download {',
                  'location = /features {\n        return 301 /#features;\n    }\n\n    location = /download {', 1)
open(p, 'w').write(s)
print('patched')
