import { Component, ChangeDetectionStrategy, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SimulationResult } from '../../models';
import { CardModule } from 'primeng/card';

@Component({
  selector: 'app-simulation-stats',
  standalone: true,
  imports: [CommonModule, CardModule],
  template: `
    <div class="flex flex-column gap-3 h-full overflow-auto p-2">
      <h3 class="text-xl font-semibold m-0">Quick Stats</h3>

      @if (data()) {
        <div class="grid grid-nogutter gap-2">
          <div class="col-12">
            <p-card styleClass="bg-primary-50">
              <div class="flex flex-column align-items-center">
                <span class="text-500 text-sm">Win Rate</span>
                <span class="text-3xl font-bold text-primary">{{ data()?.winRatePercentage | number:'1.1-1' }}%</span>
              </div>
            </p-card>
          </div>
          <div class="col-12">
            <p-card>
              <div class="flex justify-content-between align-items-center mb-2">
                <span class="text-500">Party Victories</span>
                <span class="font-bold text-green-600">{{ data()?.characterVictories }}</span>
              </div>
              <div class="flex justify-content-between align-items-center mb-2">
                <span class="text-500">Monster Victories</span>
                <span class="font-bold text-red-600">{{ data()?.monsterVictories }}</span>
              </div>
              <div class="flex justify-content-between align-items-center">
                <span class="text-500">Avg Rounds</span>
                <span class="font-bold">{{ data()?.averageRounds | number:'1.1-1' }}</span>
              </div>
            </p-card>
          </div>
        </div>

        <div class="mt-4 border-1 border-dashed border-round surface-border p-3 text-center text-500 italic">
          Additional combat analysis coming soon...
        </div>
      } @else {
        <div class="flex flex-column align-items-center justify-content-center h-full text-400 p-5 border-1 border-dashed surface-border border-round">
           <i class="pi pi-chart-bar text-4xl mb-2"></i>
           <p class="m-0">Run simulation to see stats.</p>
        </div>
      }
    </div>
  `,
  styles: [`
    :host {
      display: block;
      height: 100%;
    }
  `],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SimulationStatsComponent {
  public readonly data = input<SimulationResult | null>(null);
}
