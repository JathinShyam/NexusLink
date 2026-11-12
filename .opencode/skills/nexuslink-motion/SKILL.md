---
name: NexusLink Motion
description: Use when adding, changing, or reviewing any animation, transition, or loading state in the NexusLink web app. Encodes the routing-desk motion vocabulary, the token layer, the frequency gate, and the reduced-motion contract.
---

# NexusLink Motion

The product's motion language is **the desk powering up**: lines draw, nodes
lock in, paper settles. Nothing bounces, nothing overshoots except a single
small marker, and nothing moves for decoration alone.

## The gate (apply before writing any code)

| Frequency | Treatment |
| --- | --- |
| 100+/day — buttons, keyboard shortcuts, theme toggle | Instant feedback only. No entrance animation. |
| Tens/day — hover, focus, copy, disclosure | Near-imperceptible, 120–180ms, colour and opacity only. |
| Occasional — issue a link, open the manage dialog | Standard animation, 200–280ms. |
| Rare / first-run — page arrival, the brand draw | May be expressive, but still under ~550ms total. |

Name the purpose before building: **feedback**, **spatial consistency**,
**state indication**, **preventing a jarring change**, or **explanation**.
If none of them apply, the answer is no animation.

## Tokens — extend, never fork

All motion values live in `web/src/styles/tokens.css`; all keyframes live in
`web/src/styles/motion.css`. Never hardcode a curve or a duration in a
component.

```
--dur-fast 120ms   --dur-base 180ms   --dur-slow 260ms   --dur-scene 420ms
--ease-out         enter / exit
--ease-in-out      on-screen movement, scene changes
--ease-back        small markers only (route nodes, icon snaps)
--motion-rise      travel distance (8px / 12px)
--motion-press     button press scale (0.97)
```

## The vocabulary

| Moment | Motion | Timing |
| --- | --- | --- |
| Brand mark | route stroke draws, nodes pop in | 520ms, delay 100ms |
| Desk arrival | panel rises 8px, route line draws down, node pops | 260 / 280 / 220ms |
| Facts row | fade only | 180ms, delay 180ms |
| Composer focus | accent rule draws left→right | 180ms |
| Options disclosure | height (accordion) + fade + 4px rise | 200ms |
| Ticket issued | ink edge wipes, code rises, tear draws, meta fades | 260ms steps |
| Copy confirmed | icon pop + one outward ring pulse | 180ms / 420ms |
| Theme swap | circular View Transition reveal from the toggle | 420ms |
| Dialog | opacity + 6px rise + 0.98 scale, in and out | 180ms |
| Buttons | press scale 0.97 | 120ms |

## Rules

1. Animate `transform` and `opacity` only. `clip-path` for reveals. Never
   `width`, `height`, `margin`, `padding`, or `top`/`left` — the only
   exception is the options accordion, which is a genuine disclosure.
2. Never `transition: all`. Name every property.
3. Entrances use `cubic-bezier(0.23, 1, 0.32, 1)` (`--ease-out`). Never
   `ease-in` on an entrance.
4. One entrance per container. If a panel moves, its contents are already
   there — do not stagger children of something that is moving.
5. Stagger steps are 30–80ms and vary by importance; never a metronome.
6. Stagger only surfaces the user opens occasionally. Never on lists, menu
   rows, or anything typed into.
7. Hover motion must be gated behind `@media (hover: hover) and (pointer: fine)`.
8. Nothing loops except the loading spinner, and that one stops only for
   genuine errors, never for reduced motion.

## Reduced-motion contract

`prefers-reduced-motion: reduce` must keep the UI fully usable and fully
visible:

- Travel distances collapse to `0px` through `--motion-rise`, so entrances
  become pure fades.
- Drawing animations are switched off **and** their end state is restored
  explicitly (`stroke-dashoffset: 0`, `opacity: 0` on the copy ring) —
  otherwise the element disappears or freezes mid-effect.
- The theme swap skips the View Transition entirely.

When you add an animation, add it to the reduced-motion block in
`web/src/styles/motion.css` in the same change.

## Verification

1. `cd web && npm run lint && npm run build`.
2. Drive the real page with headless Chrome over CDP:
   `document.getAnimations()` must list only the intended animations with the
   intended durations and delays.
3. Capture mid-flight frames (60–200ms) and confirm nothing is half-drawn,
   invisible, or mis-aligned.
4. Re-run with `Emulation.setEmulatedMedia` set to
   `prefers-reduced-motion: reduce` and confirm every element is visible.
5. Check the mobile viewport: the ticket, the options panel, and the recent
   ledger all stack without overflow.
