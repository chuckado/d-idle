# rune: palette

Arcane dark: an original sigil in a ring (a stem, a triangular bowl, a diamond dot) with twelve tick marks like a clock face.

## Colours

| Name | Hex | Use |
|---|---|---|
| night | `#100d1c` | dark background |
| mist | `#f5f2fa` | light background |
| violet P | `#8a63e8` | logo/mark primary (both) |
| teal-dark Q | `#1f8f82` | logo/mark secondary (both) |
| violet bright | `#a98bff` | header/icon ring and rune (dark only) |
| teal | `#5fd4c4` | header/icon accents (dark only) |
| text | `#e6e0f5` | header wordmark |
| muted | `#a79fc4` | tagline |
| icon night | `#161128` | icon tile |

## Contrast (WCAG 2.1 relative luminance; AA = 4.5:1 text, 3:1 large text and graphics)

| Foreground | Background | Ratio | Result |
|---|---|---|---|
| logo colour P `#8a63e8` | dark `#100d1c` | 4.59 | AA text |
| logo colour P `#8a63e8` | light `#f5f2fa` | 3.77 | AA large/graphics |
| logo colour Q `#1f8f82` | dark `#100d1c` | 4.84 | AA text |
| logo colour Q `#1f8f82` | light `#f5f2fa` | 3.57 | AA large/graphics |
| text `#e6e0f5` | header `#100d1c` | 14.91 | AA text |
| muted tagline `#a79fc4` | header `#100d1c` | 7.66 | AA text |
| violet `#a98bff` | header `#100d1c` | 7.14 | AA text |
| icon violet `#a98bff` | icon `#161128` | 6.83 | AA text |

## Usage

- logo.svg and mark.svg use only P and Q, chosen so both pass AA for large text and graphics (3:1) on the dark and the light background above; the wordmark is large display type (letters are 70 px tall at 1x).
- header.svg and icon.svg are opaque and use the brighter dark-UI colours; they are meant for dark surfaces.
- On dark UI (stream page `#0d1016`, status page dark theme `#14110f`) use logo.svg as is, or the brighter header colours.
- On light surfaces (status page light theme `#f3efe6`) use logo.svg as is; never put the header colours on light.
