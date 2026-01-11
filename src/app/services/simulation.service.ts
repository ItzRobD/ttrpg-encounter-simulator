import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SimulationLog, SimulationResult, EncounterConfig, SimulationEvent, EventType } from '../models';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
import { TimelineService } from './timeline.service';
import { environment } from '../../environments/environment';
import { finalize, map, switchMap } from 'rxjs/operators';
import { from, Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class SimulationService {
  private readonly http = inject(HttpClient);
  private readonly combatantService = inject(CombatantService);
  private readonly mapperService = inject(MapperService);

  private readonly _simulationResult = signal<SimulationResult | null>(null);
  readonly simulationResult = this._simulationResult.asReadonly();

  private readonly _loading = signal(false);
  readonly loading = this._loading.asReadonly();

  private readonly timelineService = inject(TimelineService);

  /**
   * Decompresses a GZIP ArrayBuffer into a JSON object.
   */
  private async decompress(buffer: ArrayBuffer): Promise<unknown> {
    const ds = new DecompressionStream('gzip');
    const decompressedStream = new Response(buffer).body!.pipeThrough(ds);
    const response = new Response(decompressedStream);
    return response.json();
  }

  runSimulation(): void {
    const combatants = this.combatantService.combatants();
    if (combatants.length === 0) return;

    const config: EncounterConfig = {
      combatants,
      options: {
        flanking: false,
        sleepRule: 'official',
        averageDamage: true
      },
      maxRounds: 20,
      iterations: 100
    };

    this._loading.set(true);

    // TODO: Update endpoint
    const url = `${environment.apiUrl}/simulate`;

    this.http.post(url, config, { responseType: 'arraybuffer' })
      .pipe(
        switchMap(buffer => from(this.decompress(buffer))),
        map(rawResult => {
          // Map the entire structure to handle nested logs
          const mappedResult = this.mapperService.mapKeys(rawResult) as SimulationResult;

          if (!mappedResult || !Array.isArray(mappedResult.logs)) {
            throw new Error('Invalid simulation result: expected an object with a logs array.');
          }

          return mappedResult;
        }),
        finalize(() => this._loading.set(false))
      )
      .subscribe({
        next: (result) => {
          this._simulationResult.set(result);
          console.log('Simulation complete', result);
        },
        error: (err) => {
          console.error('Simulation failed', err);
        }
      });
  }

  clearResult(): void {
    this._simulationResult.set(null);
  }

  /**
   * Seeds the simulation result with dummy data from timeline_output.json.
   */
  async seedDummyData(): Promise<void> {
    try {
      const response = await fetch('/timeline_output.json');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const rawResult = await response.json();
      const result = this.mapperService.mapKeys(rawResult) as SimulationResult;

      if (!result || !Array.isArray(result.logs)) {
        throw new Error('Invalid dummy data: expected an object with a logs array.');
      }

      this._simulationResult.set(result);
      if (result.logs.length > 0) {
        this.timelineService.setSelectedSimulationLog(result.logs[0]);
        // Also update combatant service if needed, though usually the entity card
        // in simulation mode might be showing entities from the log
      }
      console.log('Dummy simulation data seeded:', result);
      console.log('Entities in seeded log:', result.logs[0]?.entities);
    } catch (err) {
      console.error('Failed to seed dummy simulation data', err);
    }
  }
}
