import { Component, inject, signal, ChangeDetectionStrategy } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ThemeService } from './services/theme.service';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet],
  templateUrl: './app.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './app.css',
})
export class App {
  // Instantiating ThemeService applies the persisted mode + dragon accent on startup.
  protected readonly theme = inject(ThemeService);
  protected readonly title = signal('dnd5e-encounter-simulator-frontend');
}
