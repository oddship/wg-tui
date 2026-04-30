---
title: wgt
layout: landing
description: Browse Warpgate targets, search locally, and launch native SSH.
---

# wgt

<h3 class="tagline text-light">Warpgate target picker for the terminal.</h3>

<div class="hstack">
  <a href="guide/quickstart/" class="button">Get started</a>
  <a href="https://github.com/oddship/wg-tui" class="button outline">GitHub</a>
</div>

<br>

<div class="features">
<article class="card">
<header><h3>Cache-first</h3></header>

Loads cached targets immediately and refreshes in the background when needed.
</article>

<article class="card">
<header><h3>Local search</h3></header>

Fuzzy search runs locally so target discovery stays fast once cached.
</article>

<article class="card">
<header><h3>Native SSH</h3></header>

Launches your system `ssh` client using Warpgate's username plus target format.
</article>
</div>

## What it does

`wgt` fetches Warpgate targets using an API token, caches them locally, lets you search them in a terminal UI, and launches SSH for the selected target.

## Status

Early but usable. Current focus is fast target discovery and clean SSH handoff.
