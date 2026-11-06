---
layout: release
title: Problem catalog
description: Browse and download Polyglot Rosetta problem catalog files.
---

The problem catalog contains exercise statements, explanations, reference implementations, and test cases. It is published independently of the engine, so catalog content can be updated without a new CLI or WebAssembly release.

For automated clients, start with the manifest. It lists the available catalog assets, their sizes and SHA-256 digests, and the minimum compatible engine version. Use the files below to browse or download the published assets directly.

## Catalog files

{% assign catalog_files = site.static_files | where_exp: "file", "file.path contains '/catalog/'" | sort: "name" %}
{% for file in catalog_files %}
{% assign catalog_dir = file.path | slice: 0, 9 %}
{% if catalog_dir == "/catalog/" %}
- [{{ file.name | escape }}]({{ file.path | relative_url | escape }})
{% endif %}
{% endfor %}
