// Dragon-scale accent themes for the simulator's theme switcher.
//
// Surfaces/text stay neutral across every theme — only the accent (PrimeNG
// `primary`) changes. Each dragon carries a full 50→950 ramp whose 500 stop is
// its defining hex, plus the button-label contrast color for dark/light modes
// (dark mode fills with the bright 500 stop; light mode fills with the deeper
// 600 stop, so the contrast differs per scheme).

export type DragonGroup = 'Chromatic' | 'Metallic' | 'Gem';

export interface DragonRamp {
  50: string; 100: string; 200: string; 300: string; 400: string;
  500: string; 600: string; 700: string; 800: string; 900: string; 950: string;
}

export interface Dragon {
  key: string;
  name: string;
  group: DragonGroup;
  hex: string;          // defining color (== ramp.500)
  desc: string;
  ramp: DragonRamp;
  darkContrast: string;  // button-label color on the bright (dark-mode) fill
  lightContrast: string; // button-label color on the deeper (light-mode) fill
}

export const DEFAULT_DRAGON = 'garnet';

export const DRAGONS: Dragon[] = [
  // ── Chromatic ──────────────────────────────────────────────────────────────
  {
    key: 'emerald', name: 'Emerald', group: 'Chromatic', hex: '#10b981',
    desc: 'Chromatic green.',
    ramp: { 50:'#ecfdf5',100:'#d1fae5',200:'#a7f3d0',300:'#6ee7b7',400:'#34d399',500:'#10b981',600:'#059669',700:'#047857',800:'#065f46',900:'#064e3b',950:'#022c22' },
    darkContrast: '#022c22', lightContrast: '#ffffff',
  },
  {
    key: 'garnet', name: 'Garnet', group: 'Chromatic', hex: '#dc2626',
    desc: 'Chromatic red. Fierce crimson. Default theme.',
    ramp: { 50:'#fef2f2',100:'#fee2e2',200:'#fecaca',300:'#fca5a5',400:'#f87171',500:'#dc2626',600:'#b91c1c',700:'#991b1b',800:'#7f1d1d',900:'#671818',950:'#450a0a' },
    darkContrast: '#ffffff', lightContrast: '#ffffff',
  },
  {
    key: 'sapphire', name: 'Sapphire', group: 'Chromatic', hex: '#2563eb',
    desc: 'Chromatic blue. Deep ocean.',
    ramp: { 50:'#eff6ff',100:'#dbeafe',200:'#bfdbfe',300:'#93c5fd',400:'#60a5fa',500:'#2563eb',600:'#1d4ed8',700:'#1e40af',800:'#1e3a8a',900:'#172554',950:'#0f172a' },
    darkContrast: '#ffffff', lightContrast: '#ffffff',
  },
  {
    key: 'onyx', name: 'Onyx', group: 'Chromatic', hex: '#64748b',
    desc: 'Chromatic black. Steel-graphite accent.',
    ramp: { 50:'#f8fafc',100:'#f1f5f9',200:'#e2e8f0',300:'#cbd5e1',400:'#94a3b8',500:'#64748b',600:'#475569',700:'#334155',800:'#1e293b',900:'#0f172a',950:'#020617' },
    darkContrast: '#ffffff', lightContrast: '#ffffff',
  },
  // ── Metallic ───────────────────────────────────────────────────────────────
  {
    key: 'gold', name: 'Gold Dragon', group: 'Metallic', hex: '#eab308',
    desc: 'Rich treasure gold.',
    ramp: { 50:'#fefce8',100:'#fef9c3',200:'#fef08a',300:'#fde047',400:'#facc15',500:'#eab308',600:'#ca8a04',700:'#a16207',800:'#854d0e',900:'#713f12',950:'#422006' },
    darkContrast: '#1a1505', lightContrast: '#1a1505',
  },
  {
    key: 'silver', name: 'Silver Dragon', group: 'Metallic', hex: '#c4ccd6',
    desc: 'Cool polished steel.',
    ramp: { 50:'#f8fafb',100:'#f0f3f6',200:'#e3e8ee',300:'#d4dbe3',400:'#cdd4dd',500:'#c4ccd6',600:'#94a0b0',700:'#6b7785',800:'#4b5563',900:'#374151',950:'#1f2937' },
    darkContrast: '#1a1d22', lightContrast: '#1a1d22',
  },
  {
    key: 'bronze', name: 'Bronze Dragon', group: 'Metallic', hex: '#cd7f32',
    desc: 'Warm brown metal.',
    ramp: { 50:'#faf3eb',100:'#f3e1cc',200:'#e8c39b',300:'#dca468',400:'#d28f47',500:'#cd7f32',600:'#b06a26',700:'#8c531f',800:'#6b3f19',900:'#4f2f14',950:'#2e1b0b' },
    darkContrast: '#1f1206', lightContrast: '#ffffff',
  },
  {
    key: 'copper', name: 'Copper Dragon', group: 'Metallic', hex: '#c2410c',
    desc: 'Reddish-orange metal.',
    ramp: { 50:'#fdf4ef',100:'#fbe3d4',200:'#f6c2a0',300:'#ef9d6b',400:'#e57a3c',500:'#c2410c',600:'#a3360a',700:'#822a08',800:'#642107',900:'#4a1905',950:'#2a0e03' },
    darkContrast: '#ffffff', lightContrast: '#ffffff',
  },
  {
    key: 'brass', name: 'Brass Dragon', group: 'Metallic', hex: '#c9a227',
    desc: 'Yellow-green alloy.',
    ramp: { 50:'#faf7e8',100:'#f3ecc3',200:'#e8d98a',300:'#dcc555',400:'#d3b53a',500:'#c9a227',600:'#a8841d',700:'#856717',800:'#654e13',900:'#4a390f',950:'#281e07' },
    darkContrast: '#1a1605', lightContrast: '#1a1605',
  },
  // ── Gem ────────────────────────────────────────────────────────────────────
  {
    key: 'amethyst', name: 'Amethyst', group: 'Gem', hex: '#a855f7',
    desc: 'Arcane purple gem dragon.',
    ramp: { 50:'#faf5ff',100:'#f3e8ff',200:'#e9d5ff',300:'#d8b4fe',400:'#c084fc',500:'#a855f7',600:'#9333ea',700:'#7e22ce',800:'#6b21a8',900:'#581c87',950:'#3b0764' },
    darkContrast: '#ffffff', lightContrast: '#ffffff',
  },
];

export const DRAGON_BY_KEY: Record<string, Dragon> =
  Object.fromEntries(DRAGONS.map(d => [d.key, d]));

export function dragonOrDefault(key: string | null | undefined): Dragon {
  return (key && DRAGON_BY_KEY[key]) || DRAGON_BY_KEY[DEFAULT_DRAGON];
}
