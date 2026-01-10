import { Component, inject } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { EntityCard } from '../../components/entity-card/entity-card';
import { CombatantService } from '../../services/combatant.service';

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
export class SimulatorShell {
  public readonly combatantService = inject(CombatantService);

  onAddDummyData() {
    this.combatantService.seedDummyData();
  }
}
