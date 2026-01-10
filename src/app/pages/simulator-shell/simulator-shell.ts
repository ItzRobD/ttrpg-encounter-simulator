import { Component, ChangeDetectionStrategy, inject, signal } from '@angular/core';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { toSignal } from '@angular/core/rxjs-interop';
import { map } from 'rxjs/operators';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { EntityCard } from '../../components/entity-card/entity-card';
import { CombatantService } from '../../services/combatant.service';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-simulator-shell',
  standalone: true,
  imports: [ButtonModule, TooltipModule, EntityCard],
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
  private readonly breakpointObserver = inject(BreakpointObserver);
  protected readonly layout = environment.layout;

  public readonly isCombatantsCollapsed = signal(false);

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

  toggleCombatants() {
    this.isCombatantsCollapsed.update((v) => !v);
  }
}
