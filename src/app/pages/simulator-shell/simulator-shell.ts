import { Component, ChangeDetectionStrategy, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { toSignal } from '@angular/core/rxjs-interop';
import { map } from 'rxjs/operators';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { MessageModule } from 'primeng/message';
import { EntityCard } from '../../components/entity-card/entity-card';
import { CombatantService } from '../../services/combatant.service';
import { SimulationService } from '../../services/simulation.service';
import { environment } from '../../../environments/environment';
import {TimelineService} from '../../services/timeline.service';
import {SimulationResults} from '../../components/simulation-results/simulation-results';
import { SimulationOptionsComponent } from '../../components/simulation-options/simulation-options';
import { EntitySelectorDialog } from '../../components/entity-selector-dialog/entity-selector-dialog';

@Component({
  selector: 'app-simulator-shell',
  standalone: true,
  imports: [
    CommonModule,
    ButtonModule,
    TooltipModule,
    MessageModule,
    EntityCard,
    SimulationResults,
    SimulationOptionsComponent,
    EntitySelectorDialog
  ],
  templateUrl: './simulator-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SimulatorShell {
  public readonly combatantService = inject(CombatantService);
  public readonly simulationService = inject(SimulationService);
  public readonly timelineService = inject(TimelineService);
  private readonly breakpointObserver = inject(BreakpointObserver);
  protected readonly layout = environment.layout;

  public readonly isCombatantsCollapsed = signal(false);
  public readonly isOptionsVisible = signal(false);
  public readonly isCharacterSelectorVisible = signal(false);
  public readonly isMonsterSelectorVisible = signal(false);

  // Use BreakpointObserver with custom query from environment
  public readonly isXL = toSignal(
    this.breakpointObserver
      .observe(`(min-width: ${this.layout.breakpointXL})`)
      .pipe(map((result) => result.matches)),
    { initialValue: false }
  );

  onAddDummyData() {
    this.combatantService.seedDummyData();
  }

  onAddTimelineDummyData() {
    this.combatantService.seedTimelineDummyData();
    this.simulationService.seedDummyData();
  }

  toggleCombatants() {
    this.isCombatantsCollapsed.update((v) => !v);
  }

  toggleOptions() {
    this.isOptionsVisible.update(v => !v);
  }
}
