import { definePreset } from '@primeuix/themes';
import Aura from '@primeuix/themes/aura';
import { DRAGON_BY_KEY, DEFAULT_DRAGON } from './dragons';

// D&D dragon-scale theme. Neutral surfaces (near-black in dark, paper in light)
// with an accent-only `primary` that the ThemeService swaps per dragon.
//
// IMPORTANT: surfaces use Aura's STANDARD orientation — surface.0 = lightest,
// surface.950 = darkest — in BOTH color schemes (dark mode just maps backgrounds
// to the dark end of the ramp). This keeps Aura's own dark defaults working. Do
// NOT invert the dark ramp (the EveTrace approach) — that needs per-token
// re-pointing hacks across every component.

const def = DRAGON_BY_KEY[DEFAULT_DRAGON];

export const DndPreset = definePreset(Aura, {
  semantic: {
    // Default accent = Emerald. Swapped at runtime via updatePreset (ThemeService).
    primary: { ...def.ramp },
    borderRadius: {
      none: '0', xs: '2px', sm: '4px', md: '8px', lg: '10px', xl: '14px',
    },
    colorScheme: {
      light: {
        primary: {
          color: '{primary.600}',
          contrastColor: def.lightContrast,
          hoverColor: '{primary.700}',
          activeColor: '{primary.800}',
        },
        // Neutral paper ramp (standard orientation: 0 lightest → 950 darkest).
        surface: {
          0:  '#ffffff',
          50: '#f7f7f8',
          100:'#f4f4f5',
          200:'#e9e9ec',
          300:'#e4e4e7',
          400:'#d4d4d8',
          500:'#a1a1aa',
          600:'#71717a',
          700:'#52525b',
          800:'#3f3f46',
          900:'#27272a',
          950:'#18181b',
        },
      },
      dark: {
        primary: {
          color: '{primary.500}',
          contrastColor: def.darkContrast,
          hoverColor: '{primary.400}',
          activeColor: '{primary.600}',
        },
        // Neutral near-black ramp (standard orientation). Dark backgrounds map to
        // the dark end: ground → 950 (#050507), card/content → 900 (#0c0c0e),
        // borders → 700 (#3a3a42); text maps to the light end (0 → #fafafa). This
        // reproduces the preferred (Dark-Reader'd) near-black + crisp-border look.
        surface: {
          0:  '#fafafa',  // text-900 / strong text
          50: '#f4f4f5',
          100:'#e4e4e7',  // text-700
          200:'#c8c8cd',  // text-600
          300:'#a1a1aa',  // text-500 / muted
          400:'#8a8a90',
          500:'#5c5c66',
          600:'#44444c',  // form-field border
          700:'#3a3a42',  // crisp borders (surface-border, content border)
          800:'#1c1c20',
          900:'#0c0c0e',  // card / content background
          950:'#050507',  // page canvas / form-field background
        },
      },
    },
  },
});
