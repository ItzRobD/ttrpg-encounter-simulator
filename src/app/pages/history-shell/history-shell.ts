import { Component } from '@angular/core';
@Component({
  selector: 'app-history-shell',
  standalone: true,
  imports: [],
  templateUrl: './history-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
})
export class HistoryShell {}
