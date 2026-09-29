---
title: Designing Data-Intensive Applications
author: Martin Kleppmann
rating: 5
finished: true
tags:
  - book
  - databases
---

# Designing Data-Intensive Applications

## Part I: Foundations

### Chapter 3: Storage and Retrieval

- Log-structured storage vs. page-oriented storage
- SSTables and LSM-trees
  - Compaction merges segments in the background
  - Bloom filters skip lookups for missing keys

> Indexes speed up reads but slow down writes.

## Part II: Distributed Data

### Chapter 5: Replication

Leader-based, multi-leader, and leaderless replication.
