---
name: guide-authoring
description: Guide for adding and editing chapters in the kubectl-mtv technical guide (Jekyll + GitHub Pages). Use when writing documentation, adding chapters, or updating the guide table of contents.
---

# Guide Authoring

The technical guide lives in `guide/` and is published as a Jekyll site on GitHub Pages.

## Structure

```
guide/
├── README.md                          # Landing page + table of contents
├── 01-overview-of-kubectl-mtv.md
├── 02-installation-and-prerequisites.md
├── ...
└── 28-karl-kubernetes-affinity-rule-language-reference.md
```

Jekyll config: `_config.yml` (project root). Theme: `minima`. Markdown: `kramdown` + Rouge.

## Adding a New Chapter

### 1. Create the file

Name it `guide/NN-slug.md` where `NN` is the next number (currently 01-28, so next is 29):

```markdown
---
layout: page
title: "Chapter 29: Your Chapter Title"
---

Content here...
```

### 2. Update the TOC

Add the chapter to `guide/README.md` in the appropriate section. The TOC is maintained manually.

### 3. Verify locally (optional)

```bash
bundle install
bundle exec jekyll serve
```

### 4. Push to main

The `pages.yml` workflow triggers on changes to `guide/**`, `_config.yml`, `_includes/**`, or `Gemfile` and deploys to GitHub Pages automatically.

## Writing Conventions

### Full Flag Notation

Always use long flags in examples:

```
kubectl-mtv get plans --namespace openshift-mtv --output json --query "where warmMigration = true"
```

Never use short flags (`-n`, `-o`, `-q`). Long flags are self-documenting and clearer for readers.

### Code Blocks

Use fenced code blocks with language hints:

````markdown
```bash
kubectl-mtv get inventory vm --provider my-vsphere --output json
```
````

### Cross-References

Link to other chapters with relative paths:

```markdown
See [Chapter 5: Providers](05-providers.md) for details.
```

## Front Matter

Every chapter needs:

```yaml
---
layout: page
title: "Chapter NN: Title"
---
```

The permalink is derived automatically from the filename: `/:basename/`.

## Jekyll Config Highlights

From `_config.yml`:

- `include: [guide]` -- ensures the `guide/` directory is processed
- `exclude: [cmd/, pkg/, ...]` -- source code directories are excluded
- Defaults for `guide/`: `layout: page`, `permalink: /:basename/`
- `header_pages: [guide/README.md]` -- guide appears in the site header

## Deployment

Push to `main` triggers `.github/workflows/pages.yml`:

1. Checkout + Ruby 3.1 + `bundle install`
2. `jekyll build` with the correct base path
3. Deploy to GitHub Pages via `actions/deploy-pages`
