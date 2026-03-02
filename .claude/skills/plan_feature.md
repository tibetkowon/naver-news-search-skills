---
name: plan-feature
description: Ensures complex features or architectural changes are planned in docs/plans/ before coding starts.
---
# Skill: Plan Feature Implementation

## Description
This skill ensures that features are carefully planned to prevent unauthorized architectural changes and waste of resources.

## Trigger
- Request for a new feature, significant architectural change, or when explicitly asked to "plan a feature".

## Instructions
1. **Analyze Requirements:** Understand the user's request and read relevant docs in `docs/`.
2. **Create Plan Document:** Create a new markdown file in `docs/plans/` (e.g., `docs/plans/news_extraction_logic.md`).
3. **Structure the Plan:** Include Goal, Requirements, Affected Files, Implementation Phases, and Verification steps.
4. **Wait for Approval:** Do NOT start coding until the user reviews and approves the plan in the chat.
