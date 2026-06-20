import { ChangeDetectionStrategy, Component, inject, model } from '@angular/core';
import { DialogModule } from 'primeng/dialog';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { ProgressBarModule } from 'primeng/progressbar';
import { TagModule } from 'primeng/tag';
import { ThemeSwitcher } from '../theme-switcher/theme-switcher';
import { ThemeService } from '../../services/theme.service';

// Live theme preview: pick a dragon accent on the left and watch it apply to real
// PrimeNG components on the right. Toggled from the header theme popover.
@Component({
  selector: 'app-theme-preview-dialog',
  imports: [DialogModule, ButtonModule, CardModule, InputTextModule, ProgressBarModule, TagModule, ThemeSwitcher],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <p-dialog
      [(visible)]="visible"
      [modal]="true"
      header="Theme Preview"
      [draggable]="false"
      [resizable]="false"
      [breakpoints]="{ '720px': '95vw' }"
      [style]="{ width: '720px' }"
    >
      <p class="tp-active">Active: <strong>{{ theme.dragon().name }}</strong> · {{ theme.dragon().hex }}</p>

      <div class="tp-cols">
        <app-theme-switcher />

        <p-card>
          <ng-template pTemplate="title"><span class="tp-dot"></span> Sample</ng-template>
          <div class="tp-sample">
            <div class="tp-actions">
              <p-button label="Run" />
              <p-button label="CR 5" [outlined]="true" />
              <p-button label="Cancel" severity="secondary" />
            </div>
            <p-progressBar [value]="78" />
            <p-progressBar [value]="42" />
            <div class="tp-tags">
              <p-tag value="Active" />
              <p-tag value="Boss" severity="secondary" />
            </div>
            <input pInputText placeholder="Encounter name…" />
            <a href="#" class="tp-link" (click)="$event.preventDefault()">A primary-colored link</a>
          </div>
        </p-card>
      </div>

      <ng-template pTemplate="footer">
        <p-button label="Done" icon="pi pi-check" (onClick)="visible.set(false)" />
      </ng-template>
    </p-dialog>
  `,
  styles: `
    .tp-active { color: var(--p-text-muted-color); margin-bottom: 1rem; }
    .tp-cols { display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; align-items: start; }
    .tp-sample { display: flex; flex-direction: column; gap: 1rem; }
    .tp-actions { display: flex; gap: .5rem; flex-wrap: wrap; }
    .tp-tags { display: flex; gap: .5rem; }
    .tp-dot { display: inline-block; width: .7rem; height: .7rem; border-radius: 50%;
              background: var(--p-primary-color); margin-right: .4rem; }
    .tp-link { color: var(--p-primary-color); font-weight: 600; }
    @media (max-width: 720px) { .tp-cols { grid-template-columns: 1fr; } }
  `,
})
export class ThemePreviewDialog {
  readonly visible = model.required<boolean>();
  protected readonly theme = inject(ThemeService);
}
