# tick: palette

Cool and lockstep: a pixel mage with a glowing staff on an isometric tile, and a segmented progress bar that never finishes.

## Colours

| Name | Hex | Use |
|---|---|---|
| deep | `#0a111d` | dark background |
| frost | `#f1f6fa` | light background |
| tile blue P | `#1c8fb4` | logo/mark primary (both) |
| royal Q | `#3d63c4` | logo/mark secondary (both) |
| cyan | `#4fd0f0` | header/icon accent (dark only) |
| glow | `#b8f4ff` | staff glow |
| robe | `#4d86e8` | hero robe (header/icon) |
| face | `#f2d4b0` | hero face |
| text | `#e2ecf8` | header wordmark |
| muted | `#9fb0c8` | tagline |
| icon deep | `#0f1a2c` | icon tile |

## Contrast (WCAG 2.1 relative luminance; AA = 4.5:1 text, 3:1 large text and graphics)

| Foreground | Background | Ratio | Result |
|---|---|---|---|
| logo colour P `#1c8fb4` | dark `#0a111d` | 5.07 | AA text |
| logo colour P `#1c8fb4` | light `#f1f6fa` | 3.42 | AA large/graphics |
| logo colour Q `#3d63c4` | dark `#0a111d` | 3.40 | AA large/graphics |
| logo colour Q `#3d63c4` | light `#f1f6fa` | 5.11 | AA text |
| text `#e2ecf8` | header `#0a111d` | 15.83 | AA text |
| muted tagline `#9fb0c8` | header `#0a111d` | 8.57 | AA text |
| cyan `#4fd0f0` | header `#0a111d` | 10.45 | AA text |
| icon cyan `#4fd0f0` | icon `#0f1a2c` | 9.64 | AA text |

## Usage

- logo.svg and mark.svg use only P and Q, chosen so both pass AA for large text and graphics (3:1) on the dark and the light background above; the wordmark is large display type (letters are 70 px tall at 1x).
- header.svg and icon.svg are opaque and use the brighter dark-UI colours; they are meant for dark surfaces.
- On dark UI (stream page `#0d1016`, status page dark theme `#14110f`) use logo.svg as is, or the brighter header colours.
- On light surfaces (status page light theme `#f3efe6`) use logo.svg as is; never put the header colours on light.
