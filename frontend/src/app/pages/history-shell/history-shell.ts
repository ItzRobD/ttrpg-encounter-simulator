import { Component, ChangeDetectionStrategy } from '@angular/core';
@Component({
  selector: 'app-history-shell',
  standalone: true,
  imports: [],
  templateUrl: './history-shell.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
})
export class HistoryShell {}
