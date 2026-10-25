#!/usr/bin/env python3
"""Validate maintained Markdown file links, excluding generated dependencies."""
from pathlib import Path
import re
import sys
from urllib.parse import unquote

root = Path(__file__).resolve().parents[1]
failures = []
links = 0
for document in root.rglob('*.md'):
    if 'vendor' in document.parts or '.git' in document.parts:
        continue
    for target in re.findall(r'\[[^\]]+\]\(([^)]+)\)', document.read_text()):
        target = target.strip('<>')
        if target.startswith(('#', 'https:', 'http:', 'mailto:', 'data:')):
            continue
        links += 1
        if not target.startswith('~/'):
            failures.append(f'{document.relative_to(root)}: file reference must start with ~/ : {target}')
            continue
        path = unquote(target.split('#', 1)[0])
        # Markdown source citations may append :line.
        path = re.sub(r':\d+$', '', path)
        if not Path(path).expanduser().exists():
            failures.append(f'{document.relative_to(root)}: missing file: {target}')
if failures:
    print('\n'.join(failures))
    sys.exit(1)
print(f'Validated {links} home-rooted local file links')
