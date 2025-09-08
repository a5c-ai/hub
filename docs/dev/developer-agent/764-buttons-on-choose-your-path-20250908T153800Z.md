# Issue #764 – Update /choose-your-path buttons

By: developer-agent(https://app.a5c.ai/a5c/agents/development/developer-agent)

## Context
Replace the buttons "add to an existing repo" and "create a new repo" on the `/choose-your-path` page with links that look like buttons. Targets:
- Add to an existing repo → `/enable`
- Create a new repo → `/new-repo`

## Plan
- Add `/choose-your-path` page (not found in repo) with two CTA links styled as buttons.
- Implement route alias `/new-repo` to redirect to existing `/repositories/new`.
- Add minimal `/enable` page as a placeholder route (destination not present), to be wired later.
- Validate frontend build locally.

## Notes
- Follow existing UI patterns using `Button asChild` with `next/link` to render actual anchors styled as buttons.

## Results
- [x] `/choose-your-path` page added with two link CTAs styled as buttons
  - "Add to existing repo" → `/enable`
  - "Create a new repo" → `/new-repo`
- [x] Route alias `/new-repo` redirects to `/repositories/new`
- [x] Added minimal `/enable` page as a placeholder route
- [x] Frontend production build completed successfully

