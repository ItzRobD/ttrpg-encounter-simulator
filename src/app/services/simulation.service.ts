import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SimulationLog, SimulationResult, EncounterConfig, SimulationEvent, EventType, SimulationConfig, isMonster,
  isCharacter, SimulationPayload, Entity, Monster, Character, SimulationStatusResponse, SimulationStatus
} from '../models';
import { SimulationOptions } from '../models/simoptions.model';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
import { TimelineService } from './timeline.service';
import { environment } from '../../environments/environment';
import { finalize, map, switchMap, takeWhile, tap, delay, repeat } from 'rxjs/operators';
import { catchError, from, Observable, retry, throwError, timer, of } from 'rxjs';
import {MonsterConfig} from '../models/configs/monster-config.model';

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

  private readonly _error = signal<string | null>(null);
  readonly error = this._error.asReadonly();

  private readonly _options = signal<SimulationOptions>(this.getDefaultOptions());
  readonly options = this._options.asReadonly();

  private readonly _config = signal<SimulationConfig>(this.getDefaultConfig());
  readonly config = this._config.asReadonly();

  private readonly _currentSimulationId = signal<string | null>(null);
  readonly currentSimulationId = this._currentSimulationId.asReadonly();

  private readonly timelineService = inject(TimelineService);

  private getDefaultConfig(): SimulationConfig {
    return {
      numberOfRuns: environment.config.defaultNumberOfRuns,
      maxRounds: environment.config.defaultMaxRounds,
      includeLogs: environment.config.defaultIncludeLogs,
    };
  }

  updateConfig(config: Partial<SimulationConfig>): void {
    this._config.update(current => ({ ...current, ...config }));
  }

  private getDefaultOptions(): SimulationOptions {
    return {
      seed: { seed1: 0, seed2: 0 },
      useHPAverageMonster: true,
      useHPAverageCharacter: false,
      canMonstersCrit: true,
      canCharactersCrit: true,
      hasIncreasedCrits: false,
      useImprovedCritical: false,
      charactersAlwaysUpcast: false,
      monstersAlwaysUpcast: false,
      allowCharacterHeals: true,
      allowMonsterHeals: true,
      aoeHitsAllEnemies: false,
      characterHealThresholdPct: 30,
      monsterHealThresholdPct: 30,
      limitedLegendaryActions: true,
      allowLairActions: true,
      allowDragonbornBreathAttack: true,
      enableClassFeatures: true,
      enableRacialFeatures: true,
      barbarianAlwaysRecklessAttack: false,
      paladinAlwaysSmite: false,
      paladinUseHighestSmiteSlot: false,
      useMassiveDamage: false,
      enableSpecialAbilities: true,
      monsterDeathEffectsHitAllies: true,
      alwaysUseSneakAttack: true,
      useWeightedAI: true,
      debugAI: false,
      hpVisibilityMode: 'visible',
      enableMonsterNoise: false,
      monsterNoiseWeight: 0.05
    };
  }

  updateOptions(options: Partial<SimulationOptions>): void {
    this._options.update(current => ({ ...current, ...options }));
  }

  /**
   * Decompresses a GZIP ArrayBuffer into a JSON object.
   */
  private async decompress(buffer: ArrayBuffer): Promise<unknown> {
    const ds = new DecompressionStream('gzip');
    const decompressedStream = new Response(buffer).body!.pipeThrough(ds);
    const response = new Response(decompressedStream);
    return response.json();
  }

  makeSimulationPayload(combatants: Entity[]): SimulationPayload | null {
    const monsters = combatants.filter(e => isMonster(e)) as Monster[];
    const characters = combatants.filter(e => isCharacter(e)) as Character[];

    if (monsters.length === 0 || characters.length === 0) {
      this._error.set('No monsters or characters to simulate.');
      return null;
    }

    const monsterIds: number[] = [];
    const monsterConfigs: MonsterConfig[] = [];

    for (const m of monsters) {
      if (!m.isCustom) {
        monsterIds.push(+m.id);
      } else {
        monsterConfigs.push(m as MonsterConfig);
      }
    }

    return {
      base_options: this.options(),
      character_configs: characters,
      monster_ids: monsterIds,
      monster_configs: monsterConfigs,
      lair_config: null,
      number_of_runs: this.config().numberOfRuns,
      max_rounds: this.config().maxRounds,
      include_logs: this.config().includeLogs,
    };
  }

  createSimulation(): void {
    const combatants = this.combatantService.combatants();
    if (combatants.length === 0) return;

    this._loading.set(true);
    this._error.set(null);
    this._currentSimulationId.set(null);

    const url = `${environment.apiUrl}/simulation/create`;

    const payload = this.makeSimulationPayload(combatants);

    if (!payload) {
      this._loading.set(false);
      return;
    }

    // Transform the payload to snake_case before sending
    const snakePayload = this.mapperService.serializeKeys(payload);

    this.http.post(url, { payload: snakePayload }, { observe: 'response' })
      .pipe(
        switchMap(response => {
          if (response.status === 202) {
            const body = response.body as any;
            const simulationId = body?.data?.simulation_id;
            if (simulationId) {
              this._currentSimulationId.set(simulationId);
              return this.pollSimulationStatus(simulationId);
            }
          }
          // If not 202 or no ID, handle as legacy or error
          this._error.set('Failed to start simulation: No simulation ID received.');
          return throwError(() => new Error('No simulation ID'));
        }),
        catchError((err) => {
          console.error('Simulation request error:', err);
          this._error.set('Simulation request failed. Please check your connection or try again later.');
          this._loading.set(false);
          return throwError(() => err);
        })
      )
      .subscribe({
        next: (result) => {
          if (result && result.logs && result.logs.length > 0) {
            this._simulationResult.set(result);
            this._loading.set(false);
            console.log('Simulation complete', result);
          }
        },
        error: (err) => {
          console.error('Simulation failed', err);
          this._loading.set(false);
        }
      });
  }

  private pollSimulationStatus(id: string): Observable<SimulationResult> {
    const statusUrl = `${environment.apiUrl}/simulation/status/${id}`;

    return timer(0, 2000).pipe(
      switchMap(() => this.http.get<any>(statusUrl)),
      map(response => {
        // Map keys to camelCase
        const mapped = this.mapperService.mapKeys(response) as SimulationStatusResponse;
        return mapped;
      }),
      // Continue polling until status is Completed or Failed
      takeWhile(status =>
        status.status !== SimulationStatus.Completed &&
        status.status !== SimulationStatus.Failed,
        true // Include the final value (Completed or Failed) in the stream
      ),
      switchMap(status => {
        if (status.status === SimulationStatus.Completed) {
          return this.fetchSimulationResult(id);
        } else if (status.status === SimulationStatus.Failed || status.error) {
          const errorMessage = status.error || 'Simulation failed on the server.';
          this._error.set(errorMessage);
          return throwError(() => new Error(errorMessage));
        }
        // Still pending/running, return empty or wait for next timer tick
        return of(null);
      }),
      // Filter out nulls (pending ticks) so subscribe only gets the final result
      takeWhile(result => result === null, true),
      map(result => result as SimulationResult),
      // We want to stop the timer once we have a result
      finalize(() => {
        // Optional cleanup
      })
    ).pipe(
      // Ensure we only emit the final result
      map(res => res),
      catchError(err => {
        this._error.set('Error polling simulation status.');
        return throwError(() => err);
      })
    ).pipe(
      // Filter to only emit the actual SimulationResult
      switchMap(res => res ? of(res) : [])
    );
  }

  private fetchSimulationResult(id: string): Observable<SimulationResult> {
    const resultUrl = `${environment.apiUrl}/simulation/results/${id}`;
    return this.http.get<any>(resultUrl).pipe(
      map(response => {
        // Map keys to camelCase first to handle the 'data' and 'results' nesting consistently
        const mappedResponse = this.mapperService.mapKeys(response) as any;

        // The backend structure is now: { data: { results: { individual_results: [...], ... }, ... } }
        const data = mappedResponse?.results;

        if (!data || !Array.isArray(data.individualResults)) {
          console.error('Invalid simulation result structure:', mappedResponse);
          throw new Error('Invalid simulation result structure');
        }

        // Convert the backend "IndividualResult" (which has a 'logs' array of events)
        // into the frontend "SimulationLog" structure.
        const logs: SimulationLog[] = data.individualResults.map((res: any) => ({
          entities: [], // If the backend provides entities per run, extract them here
          events: res.logs || []
        }));

        const simulationResult: SimulationResult = {
          ...data,
          logs: logs,
          count: data.totalRuns || logs.length
        };

        return simulationResult;
      })
    );
  }

  /**
   * Clears any current error state.
   */
  clearError(): void {
    this._error.set(null);
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
