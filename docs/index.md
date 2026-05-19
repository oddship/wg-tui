---
title: wgt
layout: landing
description: Browse Warpgate targets, search locally, launch SSH, open tunnels, and run file transfers.
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

<article class="card">
<header><h3>Tunnels</h3></header>

Open local forwards through Warpgate from inside the TUI and manage them from a dedicated screen.
</article>

<article class="card">
<header><h3>Transfers</h3></header>

Run `rsync` or `scp` uploads and downloads through Warpgate without leaving the TUI.
</article>
</div>

## What it does

`wgt` fetches Warpgate targets using an API token, caches them locally, lets you search them in a terminal UI, launches SSH for the selected target, can open local tunnels through Warpgate, and can run quick file transfers through the selected target.

## Status

Early but usable. Current focus is fast target discovery, clean SSH handoff, simple service tunneling, and practical transfer workflows.
