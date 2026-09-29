---
title: memex
aliases:
  - Memex CLI
  - semantic index
tags:
  - project
  - go
status: active
priority: 1
archived: false
created: 2026-09-01
links:
  repo: https://github.com/example/memex
---

# memex

A simple semantic indexing system for note vaults.

## Goals

1. Walk a repo and find documents.
2. Split each document into chunks by heading.
3. Embed chunks and store them for search.

## Design

### Strategies

Each application gets an index strategy, and each file type gets a doc strategy.
See [[Interfaces]] for how Go interfaces make this work.

### Events

Indexing sends events so the CLI can report progress.

## Open questions

- Should symlinked folders be followed?
- How big should a chunk be?
