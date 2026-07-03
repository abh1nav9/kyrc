# kyrc landing page

The marketing + docs home page for [kyrc](https://github.com/abh1nav9/kyrc).
Vite + React + TypeScript + Tailwind CSS v4 + Motion, static output, no backend.

```sh
npm install
npm run dev       # local dev server
npm run build     # production build → dist/
npm run preview   # serve the production build
```

## Deploy

The build is a static SPA (`base: "./"`), so it deploys anywhere:

- **Vercel** — import the repo, set the root directory to `site/`. `vercel.json`
  is included.
- **Netlify** — base `site/`, build `npm run build`, publish `site/dist`.
- **GitHub Pages** — publish `site/dist`; the relative `base` handles the
  project subpath.

## Structure

```
src/App.tsx          page composition + Motion scroll-reveal (Reveal wrapper)
src/TerminalDemo.tsx browser replay of the TUI (typing → results, looped)
src/InstallTabs.tsx  npm / bun / pnpm / npx install commands + copy
src/styles.css       Tailwind entry + @theme design tokens + caret keyframe
```

Styling is Tailwind v4 (CSS-first: theme tokens live in `@theme` in
`styles.css`, no `tailwind.config.js`). Animations use `motion` — entrance +
scroll reveals, the sliding install-tab underline (`layoutId`), and the
test→results transition (`AnimatePresence`).

The terminal demo is a scripted animation that mirrors the real TUI's visual
states (untyped / correct / wrong / caret → results); it is not the actual
engine, but it shows exactly what the tool looks like.
