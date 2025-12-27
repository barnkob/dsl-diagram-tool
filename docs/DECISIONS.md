# Decisions

## 2025-12-04: Use Go over Rust
Context: Need fast development, good CLI ecosystem.
Decision: Go for faster development and mature CLI tooling. Rust considered for future if performance becomes critical.

## 2025-12-04: Start with D2 only
Context: MVP scope decision.
Decision: Focus on one DSL to prove concept. Multi-DSL support adds complexity better tackled after core is solid.

## 2025-12-04: Use official D2 library
Context: Discovered terrastruct's official D2 library (v0.7.1, 22k+ stars).
Decision: Wrap it with our abstractions rather than building custom parser. Provides complete parsing, layout, and rendering.

## 2025-12-04: Metadata in separate files
Context: How to store layout customizations.
Decision: `.d2meta` sidecar files keep DSL clean, easier Git merges, optional usage.

## 2025-12-22: Positions as offsets
Context: How to persist node positions.
Decision: Store (dx, dy) offsets from auto-layout, not absolute coordinates. Allows layout algorithm changes without breaking saved positions.

## 2025-12-22: Source hash for layout invalidation
Context: When D2 content changes, old positions may be invalid.
Decision: Store hash in `.d2meta`; clear positions when hash changes.
