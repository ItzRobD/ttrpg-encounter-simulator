import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SimulationLog, SimulationResult, EncounterConfig, SimulationEvent, EventType, SimulationConfig,
  SimulationPayload, Actor, SimulationStatusResponse, SimulationStatus, ApiResponse, IndividualResult, isMonster, isCharacter
} from '../models';
import { SimulationOptions } from '../models/simoptions.model';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
import { SimulationStateService } from './simulation-state.service';
import { environment } from '../../environments/environment';
import { filter, finalize, map, switchMap, take, takeWhile, tap } from 'rxjs/operators';
import { catchError, Observable, throwError, timer, of } from 'rxjs';
import {MonsterConfig} from '../models/configs/monster-config.model';

@Injectable({
  providedIn: 'root',
})
export class SimulationService {
  private readonly http = inject(HttpClient);
  private readonly combatantService = inject(CombatantService);
  private readonly mapperService = inject(MapperService);
  private readonly stateService = inject(SimulationStateService);

  readonly simulationResult = this.stateService.simulationResult;
  readonly loading = this.stateService.loading;
  readonly error = this.stateService.error;

  private readonly _options = signal<SimulationOptions>(this.getDefaultOptions());
  readonly options = this._options.asReadonly();

  private readonly _config = signal<SimulationConfig>(this.getDefaultConfig());
  readonly config = this._config.asReadonly();

  private readonly _currentSimulationId = signal<string | null>(null);
  readonly currentSimulationId = this._currentSimulationId.asReadonly();

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

  makeSimulationPayload(combatants: Actor[]): SimulationPayload | null {
    const monsters = combatants.filter(a => isMonster(a)) as Actor[];
    const characters = combatants.filter(a => isCharacter(a)) as Actor[];

    if (monsters.length === 0 || characters.length === 0) {
      this.stateService.setError('No monsters or characters to simulate.');
      return null;
    }

    const monsterIds: number[] = [];
    const actorConfigs: any[] = [];

    // Process Monsters
    for (const m of monsters) {
      if (!m.isCustom) {
        monsterIds.push(+m.id);
      } else {
        const { state, ...config } = m;
        actorConfigs.push(config);
      }
    }

    // Process Characters
    for (const c of characters) {
      const { state, ...config } = c;
      actorConfigs.push(config);
    }

    return {
      base_options: this.options(),
      actor_configs: actorConfigs,
      monster_ids: monsterIds,
      number_of_runs: this.config().numberOfRuns,
      max_rounds: this.config().maxRounds,
      include_logs: this.config().includeLogs,
    };
  }

  createSimulation(): void {
    const combatants = this.combatantService.combatants();
    if (combatants.length === 0) return;

    this.stateService.setLoading(true);
    this.stateService.setError(null);
    this._currentSimulationId.set(null);

    const url = `${environment.apiUrl}/simulation/create`;

    const payload = this.makeSimulationPayload(combatants);

    if (!payload) {
      this.stateService.setLoading(false);
      return;
    }

    this.http.post(url, { payload }, { observe: 'response' })
      .pipe(
        switchMap(response => {
          if (response.status === 202) {
            const body = response.body as any;
            const simulationId = body?.data?.simulationId || body?.simulationId;
            if (simulationId) {
              this._currentSimulationId.set(simulationId);
              return this.pollSimulationStatus(simulationId);
            }
          }
          // If not 202 or no ID, handle as legacy or error
          this.stateService.setError('Failed to start simulation: No simulation ID received.');
          return throwError(() => new Error('No simulation ID'));
        }),
        catchError((err) => {
          console.error('Simulation request error:', err);
          this.stateService.setError('Simulation request failed. Please check your connection or try again later.');
          this.stateService.setLoading(false);
          return throwError(() => err);
        })
      )
      .subscribe({
        next: (result) => {
          if (result && result.logs && result.logs.length > 0) {
            this.stateService.setSimulationResult(result);
            this.stateService.setLoading(false);
            console.log('Simulation complete', result);
          }
        },
        error: (err) => {
          console.error('Simulation failed', err);
          this.stateService.setLoading(false);
        }
      });
  }

  pollSimulationStatus(id: string): Observable<SimulationResult> {
    const statusUrl = `${environment.apiUrl}/simulation/status/${id}`;

    return timer(0, 2000).pipe(
      switchMap(() => this.http.get<SimulationStatusResponse>(statusUrl)),
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
          this.stateService.setError(errorMessage);
          return throwError(() => new Error(errorMessage));
        }
        return of(null);
      }),
      filter((result): result is SimulationResult => result !== null),
      take(1),
      catchError(err => {
        if (err.message && err.message.includes('missing individualResults')) {
          this.stateService.setError('Simulation completed but returned invalid data structure.');
        } else if (this.stateService.error() === null) {
          this.stateService.setError('Error polling simulation status.');
        }
        return throwError(() => err);
      })
    );
  }

  public fetchSimulationResult(id: string): Observable<SimulationResult> {
    const resultUrl = `${environment.apiUrl}/simulation/results/${id}`;
    return this.http.get<any>(resultUrl).pipe(
      map(response => {
        // The backend structure is now: { data: { results: { individual_results: [...], ... }, actor_configs: { ... } } }
        const unwrapped = response?.data || response;
        const data = unwrapped?.results || unwrapped;
        const actorConfigs = unwrapped?.actorConfigs;

        if (!data || !Array.isArray(data.individualResults)) {
          console.error('Invalid simulation result structure. Full response:', response);
          throw new Error('Invalid simulation result structure: missing individualResults');
        }

        // Convert the backend "IndividualResult" (which has a 'logs' array of events)
        // into the frontend "SimulationLog" structure for backward compatibility/UI.
        const simulationLogs: SimulationLog[] = data.individualResults.map((res: any) => ({
          actors: [], // We'll populate this if needed, but TimelineService will use initialState
          events: res.logs || [],
          initialState: res.initialState // Pass along individual run initial state if present
        }));

        const simulationResult: SimulationResult = {
          ...data,
          actorConfigs: actorConfigs,
          logs: simulationLogs,
          count: data.totalRuns || simulationLogs.length
        };

        return simulationResult;
      }),
      tap(result => {
        this.stateService.setSimulationResult(result);
        if (result.logs.length > 0) {
          // Default to the logs of the first run
          this.stateService.setSelectedSimulationLog(result.logs[0]);
        }
      })
    );
  }

  /**
   * Clears any current error state.
   */
  clearError(): void {
    this.stateService.setError(null);
  }

  clearResult(): void {
    this.stateService.setSimulationResult(null);
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

      this.stateService.setSimulationResult(result);
      if (result.logs.length > 0) {
        // Dummy data might not have individualResults, so we use logs[0]
        this.stateService.setSelectedSimulationLog(result.logs[0]);
      }
      console.log('Dummy simulation data seeded:', result);
    } catch (err) {
      console.error('Failed to seed dummy simulation data', err);
    }
  }
}
