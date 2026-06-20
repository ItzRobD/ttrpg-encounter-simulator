import { Injectable, computed, effect, signal } from '@angular/core';
import { updatePreset } from '@primeuix/themes';
import { DRAGONS, DEFAULT_DRAGON, dragonOrDefault, type Dragon } from '../theme/dragons';

export type ThemeMode = 'dark' | 'light';

const MODE_KEY = 'dnd-theme-mode';
const DRAGON_KEY = 'dnd-theme-dragon';

// Owns the two theme dimensions:
//  • mode   → toggles `.app-dark` on <html> (PrimeNG darkModeSelector + our canvas var key off it)
//  • dragon → swaps the PrimeNG `primary` accent via updatePreset, keeping surfaces neutral
// Both default to (dark, emerald) and persist to localStorage. Surfaces never change
// with the dragon — only the accent does.
@Injectable({ providedIn: 'root' })
export class ThemeService {
  readonly dragons = DRAGONS;

  readonly mode = signal<ThemeMode>(this.initialMode());
  readonly dragonKey = signal<string>(this.initialDragon());
  readonly dragon = computed<Dragon>(() => dragonOrDefault(this.dragonKey()));
  readonly isDark = computed(() => this.mode() === 'dark');

  constructor() {
    effect(() => this.applyMode(this.mode()));
    effect(() => this.applyDragon(this.dragon()));
  }

  toggleMode(): void {
    this.mode.update(m => (m === 'dark' ? 'light' : 'dark'));
  }

  setMode(mode: ThemeMode): void {
    this.mode.set(mode);
  }

  setDragon(key: string): void {
    this.dragonKey.set(dragonOrDefault(key).key);
  }

  private initialMode(): ThemeMode {
    return localStorage.getItem(MODE_KEY) === 'light' ? 'light' : 'dark';
  }

  private initialDragon(): string {
    return dragonOrDefault(localStorage.getItem(DRAGON_KEY) ?? DEFAULT_DRAGON).key;
  }

  private applyMode(mode: ThemeMode): void {
    document.documentElement.classList.toggle('app-dark', mode === 'dark');
    localStorage.setItem(MODE_KEY, mode);
  }

  private applyDragon(d: Dragon): void {
    updatePreset({
      semantic: {
        primary: { ...d.ramp },
        colorScheme: {
          light: { primary: { color: '{primary.600}', contrastColor: d.lightContrast, hoverColor: '{primary.700}', activeColor: '{primary.800}' } },
          dark:  { primary: { color: '{primary.500}', contrastColor: d.darkContrast,  hoverColor: '{primary.400}', activeColor: '{primary.600}' } },
        },
      },
    });
    localStorage.setItem(DRAGON_KEY, d.key);
  }
}
