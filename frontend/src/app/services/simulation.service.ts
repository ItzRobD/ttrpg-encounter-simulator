import { Injectable, signal, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SimulationLog, SimulationResult, SimulationEvent, EventType, SimulationConfig,
  SimulationPayload, Actor, SimulationStatusResponse, SimulationStatus, IndividualResult, isMonster, isCharacter, SimulationResponse, IntermissionConfig, SimulationEncounterConfig
} from '../models';
import { SimulationOptions } from '../models';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
import { SimulationStateService } from './simulation-state.service';
import { environment } from '../../environments/environment';
import { filter, finalize, map, switchMap, take, takeWhile, tap } from 'rxjs/operators';
import { catchError, Observable, throwError, timer, of } from 'rxjs';
import {MonsterConfig} from '../models/configs/monster-config.model';

/**
 * Loosely-typed view of the raw simulation results payload as returned by the
 * API and reshaped here into a SimulationResult. Mirrors the backend
 * MultiSimulationResult after key-mapping (see pkg/simulation/multi_runner.go).
 */
interface RawSimResults {
  individualResults?: IndividualResult[];
  actorConfigs?: Record<string, Actor>;
  initialState?: Record<string, unknown>;
  totalRuns?: number;
  [key: string]: unknown;
}

interface RawSimEnvelope {
  results?: RawSimResults;
}

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

  private readonly _intermissionConfig = signal<IntermissionConfig>(this.getDefaultIntermissionConfig());
  readonly intermissionConfig = this._intermissionConfig.asReadonly();

  private getDefaultIntermissionConfig(): IntermissionConfig {
    return {
      maxShortRests: 2,
      shortRestHealThreshold: 0.7,
      postRestHealThreshold: 0.9
    };
  }

  updateIntermissionConfig(config: Partial<IntermissionConfig>): void {
    this._intermissionConfig.update(current => ({ ...current, ...config }));
  }

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
      characterEmergencyThresholdPct: 15,
      monsterEmergencyThresholdPct: 15,
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
      monsterNoiseWeight: 0.05,
      includeStateSnapshots: true,
      maxLoggedRuns: 10
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
    const party = this.combatantService.party();
    const encountersData = this.combatantService.encounters();

    if (party.length === 0 || encountersData.every(e => e.length === 0)) {
      this.stateService.setError('No party or encounters to simulate.');
      return null;
    }

    const characterConfigs: Actor[] = [];
    for (const c of party) {
      const { state, ...config } = c;
      characterConfigs.push(config as Actor);
    }

    const encounters: SimulationEncounterConfig[] = encountersData.map((e, i) => {
      const monsterIds: number[] = [];
      const monsterConfigs: Actor[] = [];

      for (const m of e) {
        if (m.isCustom) {
          const { state, ...config } = m;
          monsterConfigs.push(config as Actor);
        } else {
          monsterIds.push(+m.id);
        }
      }

      return {
        name: `Encounter ${i + 1}`,
        monsterIds: monsterIds,
        monsterConfigs: monsterConfigs
      };
    });

    return {
      base_options: this.options(),
      character_configs: characterConfigs,
      encounters: encounters,
      intermission: this.intermissionConfig(),
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
            const body = response.body as { data?: { simulationId: string }, simulationId?: string };
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
    return this.http.get<unknown>(resultUrl).pipe(
      map(response => {
        // The mappingInterceptor might automatically unwrap the 'data' envelope.
        // If it's still there, we map it, otherwise we map the top-level response.
        const mappedResponse = this.mapperService.mapKeys(response) as RawSimEnvelope;
        console.log('[SimulationService] fetchSimulationResult mappedResponse:', mappedResponse);

        // Based on the log, results should be either at top level or inside a 'data' key that was already unwrapped.
        const data = mappedResponse;
        const results = data.results;

        if (!results || !Array.isArray(results.individualResults)) {
          console.error('Invalid simulation result structure. Full response:', response);
          console.error('Mapped data:', data);
          if (results) {
            console.error('Results keys:', Object.keys(results));
          }
          throw new Error('Invalid simulation result structure: missing individualResults');
        }

        // actorConfigs is now inside results and contains ALL actors
        const actorConfigsMap = results.actorConfigs || {};

        const actorConfigs: Actor[] = Object.entries(actorConfigsMap).map(([id, config]: [string, Actor]) => {
          const actor = {
            ...config,
            instanceId: config.instanceId || Number(id)
          } as Actor;

          // If ac is 0 or missing, try to set it from the config if available
          if ((!actor.ac || actor.ac === 0) && config.ac) {
            actor.ac = config.ac;
          }

          return actor;
        });

        const simulationLogs: SimulationLog[] = results.individualResults.map((res: IndividualResult) => {
          // Flatten encounters for the current simple view, but preserve encounter context
          const allEvents = res.encounterResults.flatMap(er => er.logs);

          // Find the first event with actor_states as initial states if res.actorInitialStates is missing
          const firstStateEvent = allEvents.find(e => e.type === EventType.CombatStart && e.data?.actorStates);

          // Use the new encounter-level initial_state if available (from the first encounter)
          const encounterInitialState = res.encounterResults[0]?.initialState;

          const actorInitialStates = res.actorInitialStates || encounterInitialState || firstStateEvent?.data?.actorStates;

          return {
            actors: [],
            events: allEvents,
            initialState: res.initialState || results.initialState,
            actorInitialStates: actorInitialStates,
            actorConfigs: actorConfigs
          };
        });

        // `results` is the raw envelope; the win/round aggregate fields are
        // spread through and the reshaped collections overwrite their raw forms.
        const simulationResult = {
          ...results,
          actorConfigs: actorConfigs,
          logs: simulationLogs,
          count: results.totalRuns || simulationLogs.length
        } as unknown as SimulationResult;

        return simulationResult;
      }),
      tap(result => {
        this.stateService.setSimulationResult(result);
        if (result.logs.length > 0) {
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
      const response = await fetch('/sim_adv_day_result_resp.json');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const rawResult = await response.json();

      // Map the full response structure
      const data = this.mapperService.mapKeys(rawResult) as SimulationResponse;
      const results = data.results;
      const actorConfigsMap = results.actorConfigs || {};
      const actorConfigs: Actor[] = Object.entries(actorConfigsMap).map(([id, configRaw]) => {
        const config = configRaw as Actor;
        const actor = {
          ...config,
          instanceId: config.instanceId || Number(id)
        } as Actor;

        if ((!actor.ac || actor.ac === 0) && config.ac) {
          actor.ac = config.ac;
        }

        return actor;
      });

      // Extract and format using the same logic as fetchSimulationResult
      const simulationLogs: SimulationLog[] = results.individualResults.map((res: IndividualResult) => {
        const allEvents = res.encounterResults.flatMap(er => er.logs);
        const firstStateEvent = allEvents.find(e => e.type === EventType.CombatStart && e.data?.actorStates);
        const encounterInitialState = res.encounterResults[0]?.initialState;

        return {
          actors: [],
          events: allEvents,
          initialState: res.initialState || results.initialState,
          actorInitialStates: res.actorInitialStates || encounterInitialState || firstStateEvent?.data?.actorStates,
          actorConfigs: actorConfigs
        };
      });

      const result: SimulationResult = {
        ...results,
        actorConfigs: actorConfigs,
        logs: simulationLogs,
        count: results.totalRuns || simulationLogs.length
      };

      this.stateService.setSimulationResult(result);
      if (result.logs.length > 0) {
        this.stateService.setSelectedSimulationLog(result.logs[0]);
      }
    } catch (err) {
      console.error('Failed to seed dummy simulation data', err);
    }
  }
}
