import { Component, ChangeDetectionStrategy, inject, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { toSignal } from '@angular/core/rxjs-interop';
import { map } from 'rxjs/operators';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { MessageModule } from 'primeng/message';
import {Tab, TabList, Tabs} from 'primeng/tabs';
import { ActorCard } from '../../components/actor-card/actor-card.component';
import { CombatantService } from '../../services/combatant.service';
import { SimulationService } from '../../services/simulation.service';
import { environment } from '../../../environments/environment';
import {TimelineService} from '../../services/timeline.service';
import {SimulationResults} from '../../components/simulation-results/simulation-results';
import { SimulationOptionsComponent } from '../../components/simulation-options/simulation-options';
import { ActorSelectorDialog } from '../../components/actor-selector-dialog/actor-selector-dialog.component';

import { SimulationStatsComponent } from '../../components/simulation-stats/simulation-stats';

@Component({
  selector: 'app-simulator-shell',
  standalone: true,
  imports: [
    CommonModule,
    ButtonModule,
    TooltipModule,
    MessageModule,
    Tab,
    Tabs,
    TabList,
    ActorCard,
    SimulationResults,
    SimulationOptionsComponent,
    ActorSelectorDialog,
    SimulationStatsComponent
  ],
  templateUrl: './simulator-shell.html',
  styles: [
    `
      :host {
        display: block;
      }

      .combatants-panel {
        transition: all 0.3s ease-in-out;
      }

      .stats-panel {
         transition: all 0.3s ease-in-out;
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
  public readonly isStatsCollapsed = signal(false);
  public readonly isOptionsVisible = signal(false);
  public readonly isCharacterSelectorVisible = signal(false);
  public readonly isMonsterSelectorVisible = signal(false);

  private readonly _timelineDebug = effect(() => {
    this.timelineService.projectedState();
  });

  private readonly _logCombatants = effect(() => {
    const combatants = this.combatantService.combatants();
    console.log(`[SimulatorShell] Current combatants:`, combatants.map(c => ({ name: c.name, instanceId: c.instanceId })));
  });

  // Use BreakpointObserver with custom query from environment
  public readonly isXL = toSignal(
    this.breakpointObserver
      .observe(`(min-width: ${this.layout.breakpointXL})`)
      .pipe(map((result) => result.matches)),
    { initialValue: false }
  );

  toggleCombatants() {
    this.isCombatantsCollapsed.update((v) => !v);
  }

  toggleStats() {
    this.isStatsCollapsed.update((v) => !v);
  }

  toggleOptions() {
    this.isOptionsVisible.update(v => !v);
  }

  fetchSpecificSimulation(): void {
    const id = '019c62ed-304d-793d-9c3a-5411ef60e959';
    // const id = '019c59b0-0909-7053-9273-1926304c6442';
    this.simulationService.fetchSimulationResult(id).subscribe(result => {
        if (result) {
            this.combatantService.loadFromSimulation(result);
        }
    });
  }
}
