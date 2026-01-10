import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SimulationLog, EncounterConfig, SimulationEvent } from '../models';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
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

  private readonly _simulationResult = signal<SimulationLog | null>(null);
  readonly simulationResult = this._simulationResult.asReadonly();

  private readonly _loading = signal(false);
  readonly loading = this._loading.asReadonly();

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

    const url = `${environment.apiUrl}/simulate`;

    this.http.post(url, config, { responseType: 'arraybuffer' })
      .pipe(
        switchMap(buffer => from(this.decompress(buffer))),
        map(rawEvents => {
          if (!Array.isArray(rawEvents)) {
            throw new Error('Invalid simulation result: expected an array of events.');
          }
          const events = rawEvents.map(e => this.mapperService.mapKeys(e) as SimulationEvent);
          return {
            events,
            count: events.length
          } as SimulationLog;
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
}
