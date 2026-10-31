---
layout: default
title: Polyglot Rosetta
---

# Polyglot Rosetta

A cross-language coding catalog with a command-line tool and WebAssembly core.

## Catalog

<ul>
{% for file in site.static_files %}
{% if file.path contains '/catalog/' %}
  <li><a href="{{ file.path | relative_url | escape }}">{{ file.name | escape }}</a></li>
{% endif %}
{% endfor %}
</ul>

## Downloads and source

- [Releases](https://github.com/venkateshcv1809/polyglot-rosetta/releases): CLI binaries for Linux, macOS, and Windows on amd64 and arm64, plus the WebAssembly core and JavaScript runtime bridge.
- [Source code](https://github.com/venkateshcv1809/polyglot-rosetta)
- [Portfolio](https://venkateshcv1809.github.io/)
