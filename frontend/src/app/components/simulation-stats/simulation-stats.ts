import { Component, ChangeDetectionStrategy, input } from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { SimulationResult } from '../../models';
import { CardModule } from 'primeng/card';

@Component({
  selector: 'app-simulation-stats',
  imports: [DecimalPipe, CardModule],
  template: `
    <div class="stats-panel">
      <h3 class="stats-title">Quick Stats</h3>

      @if (data()) {
        <div class="stats-cards">
          <p-card styleClass="win-card">
            <div class="stats-winrate">
              <span class="stats-winrate-label">Win Rate</span>
              <span class="stats-winrate-value">{{ data()?.winRatePercentage | number:'1.1-1' }}%</span>
            </div>
          </p-card>
          <p-card>
            <div class="stats-row">
              <span class="stats-row-label">Party Victories</span>
              <span class="stats-row-value party">{{ data()?.characterVictories }}</span>
            </div>
            <div class="stats-row">
              <span class="stats-row-label">Monster Victories</span>
              <span class="stats-row-value monster">{{ data()?.monsterVictories }}</span>
            </div>
            <div class="stats-row">
              <span class="stats-row-label">Avg Rounds</span>
              <span class="stats-row-value">{{ data()?.averageRounds | number:'1.1-1' }}</span>
            </div>
          </p-card>
        </div>

        <div class="stats-soon">Additional combat analysis coming soon...</div>
      } @else {
        <div class="stats-empty">
          <i class="pi pi-chart-bar stats-empty-icon"></i>
          <p class="stats-empty-text">Run simulation to see stats.</p>
        </div>
      }
    </div>
  `,
  styleUrl: './simulation-stats.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SimulationStatsComponent {
  public readonly data = input<SimulationResult | null>(null);
}
