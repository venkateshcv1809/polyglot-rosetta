---
layout: default
title: Catalog files
---

# Catalog files

[Back to Polyglot Rosetta]({{ '/' | relative_url }})

Browse the current catalog files. Select a filename to open or download it.

{% assign files = site.data.catalog.categories | unshift: site.data.catalog.templates | unshift: site.data.catalog.index %}
<pre aria-label="Catalog file tree">catalog/
├── <a href="manifest.json">manifest.json</a>
{% for file in files %}{% if forloop.last %}└──{% else %}├──{% endif %} <a href="{{ file.path | escape }}">{{ file.path | escape }}</a> ({{ file.size }} bytes)
{% endfor %}</pre>

{% if site.data.catalog.categories.size == 0 %}
No problem category archives are published yet.
{% endif %}

The [manifest](manifest.json) contains compatibility requirements and SHA-256
hashes. This listing updates with each catalog publication.
