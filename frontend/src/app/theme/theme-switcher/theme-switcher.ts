import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { ThemeService } from '../../services/theme.service';
import { DragonGroup } from '../dragons';

// Theme switcher: dark/light mode toggle + dragon-accent picker. Surfaces stay
// neutral; selecting a dragon swaps only the PrimeNG `primary` accent.
@Component({
  selector: 'app-theme-switcher',
  imports: [ButtonModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="switcher">
      <div class="row">
        <span class="label">Appearance</span>
        <p-button
          [text]="true"
          severity="secondary"
          [icon]="theme.isDark() ? 'pi pi-moon' : 'pi pi-sun'"
          [label]="theme.isDark() ? 'Dark' : 'Light'"
          (onClick)="theme.toggleMode()"
          [attr.aria-pressed]="theme.isDark()"
          aria-label="Toggle dark mode" />
      </div>

      @for (group of groups; track group) {
        <div class="group">
          <span class="group-label">{{ group }}</span>
          <div class="swatches">
            @for (d of dragonsIn(group); track d.key) {
              <button
                type="button"
                class="swatch"
                [class.selected]="d.key === theme.dragonKey()"
                [style.--swatch]="d.hex"
                (click)="theme.setDragon(d.key)"
                [attr.aria-pressed]="d.key === theme.dragonKey()"
                [attr.title]="d.name + ' — ' + d.hex">
                <span class="dot"></span>
                <span class="name">{{ d.name }}</span>
              </button>
            }
          </div>
        </div>
      }
    </div>
  `,
  styles: `
    .switcher { display: flex; flex-direction: column; gap: 1rem; max-width: 22rem; }
    .row { display: flex; align-items: center; justify-content: space-between; }
    .label { font-weight: 600; color: var(--p-text-color); }
    .group { display: flex; flex-direction: column; gap: .5rem; }
    .group-label {
      font-size: .7rem; font-weight: 700; letter-spacing: .12em; text-transform: uppercase;
      color: var(--p-text-muted-color);
    }
    .swatches { display: grid; grid-template-columns: repeat(2, 1fr); gap: .5rem; }
    .swatch {
      display: flex; align-items: center; gap: .5rem;
      padding: .5rem .65rem; border-radius: var(--p-content-border-radius, 8px);
      background: var(--p-content-background);
      border: 1px solid var(--p-content-border-color);
      color: var(--p-text-color); cursor: pointer; font: inherit; text-align: left;
      transition: border-color .15s, box-shadow .15s;
    }
    .swatch:hover { border-color: var(--swatch); }
    .swatch.selected { border-color: var(--swatch); box-shadow: 0 0 0 1px var(--swatch); }
    .swatch:focus-visible { outline: 2px solid var(--p-primary-color); outline-offset: 2px; }
    .dot { width: 1rem; height: 1rem; border-radius: 50%; background: var(--swatch); flex: none; }
    .name { font-size: .85rem; }
  `,
})
export class ThemeSwitcher {
  protected readonly theme = inject(ThemeService);
  protected readonly groups: DragonGroup[] = ['Chromatic', 'Metallic', 'Gem'];

  protected dragonsIn(group: DragonGroup) {
    return this.theme.dragons.filter(d => d.group === group);
  }
}
