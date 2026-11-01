# cursor: palette

Terminal ember: a command prompt with a block cursor waiting for input. The bot sits at the prompt for days.

## Colours

| Name | Hex | Use |
|---|---|---|
| coal | `#0e0c0a` | dark background |
| paper | `#f6f1e7` | light background |
| ember P | `#e0630f` | logo/mark primary (both backgrounds) |
| amber-dark Q | `#a8700a` | logo/mark cursor (both backgrounds) |
| ember bright | `#ff8a1f` | header/icon accent (dark only) |
| amber | `#ffc15a` | icon cursor (dark only) |
| ash | `#ece2d4` | header wordmark |
| muted | `#b4a797` | tagline |
| icon coal | `#15100c` | icon tile |

## Contrast (WCAG 2.1 relative luminance; AA = 4.5:1 text, 3:1 large text and graphics)

| Foreground | Background | Ratio | Result |
|---|---|---|---|
| logo colour P `#e0630f` | dark `#0e0c0a` | 5.54 | AA text |
| logo colour P `#e0630f` | light `#f6f1e7` | 3.13 | AA large/graphics |
| logo colour Q `#a8700a` | dark `#0e0c0a` | 4.64 | AA text |
| logo colour Q `#a8700a` | light `#f6f1e7` | 3.74 | AA large/graphics |
| ash text `#ece2d4` | header `#0e0c0a` | 15.24 | AA text |
| muted tagline `#b4a797` | header `#0e0c0a` | 8.29 | AA text |
| ember `#ff8a1f` | header `#0e0c0a` | 8.28 | AA text |
| icon ember `#ff8a1f` | icon `#15100c` | 8.02 | AA text |

## Usage

- logo.svg and mark.svg use only P and Q, chosen so both pass AA for large text and graphics (3:1) on the dark and the light background above; the wordmark is large display type (letters are 70 px tall at 1x).
- header.svg and icon.svg are opaque and use the brighter dark-UI colours; they are meant for dark surfaces.
- On dark UI (stream page `#0d1016`, status page dark theme `#14110f`) use logo.svg as is, or the brighter header colours.
- On light surfaces (status page light theme `#f3efe6`) use logo.svg as is; never put the header colours on light.
