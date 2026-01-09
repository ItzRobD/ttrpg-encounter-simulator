import { Component } from '@angular/core';
@Component({
  selector: 'app-bestiary-shell',
  standalone: true,
  imports: [],
  templateUrl: './bestiary-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
})
export class BestiaryShell {}
