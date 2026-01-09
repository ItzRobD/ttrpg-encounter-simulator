import { Component } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import {EntityCard} from '../../components/entity-card/entity-card';
@Component({
  selector: 'app-simulator-shell',
  standalone: true,
  imports: [ButtonModule, EntityCard],
  templateUrl: './simulator-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
})
export class SimulatorShell {}
