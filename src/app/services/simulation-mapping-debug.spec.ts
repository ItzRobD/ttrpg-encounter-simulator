import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { SimulationService } from './simulation.service';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
import { environment } from '../../environments/environment';
import { signal } from '@angular/core';

describe('SimulationMappingDebug', () => {
  let service: SimulationService;
  let httpMock: HttpTestingController;

  const simulationId = '019bd814-7b89-7022-9f8e-3e73ed185b5f';

  beforeEach(() => {
    const combatantServiceSpy = {
      combatants: signal([]),
      count: signal(0)
    };

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        SimulationService,
        MapperService,
        { provide: CombatantService, useValue: combatantServiceSpy }
      ]
    });

    service = TestBed.inject(SimulationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should fetch and map simulation results for ID: ' + simulationId, async () => {
    // We're using the real URL but mocking the response to see how it flows through fetchSimulationResult
    // Accessing private method for debugging
    const promise = new Promise((resolve, reject) => {
      (service as any).fetchSimulationResult(simulationId).subscribe({
        next: (result: any) => {
          expect(result).toBeTruthy();
          expect(result.totalRuns).toBeGreaterThan(0);
          expect(result.logs.length).toBeGreaterThan(0);
          resolve(true);
        },
        error: (err: any) => {
          console.error('Mapping FAILED');
          console.error('Error:', err);
          reject(err);
        }
      });

      const req = httpMock.expectOne(`${environment.apiUrl}/simulation/results/${simulationId}`);

      // We'll simulate the response based on what the user said it looks like
      // {data: {created_at: "...", entity_configs: {...}, results: {total_runs: 2, individual_results: [...]}}}
      req.flush({
        data: {
          created_at: "2026-01-19T16:02:10.135146Z",
          entity_configs: {
            monster_configs: [],
            character_configs: [
              {
                instance_id: 1,
                name: "Henry",
                as_config: { ability_scores: {}, proficiencies: {} },
                hp: { value: 119 }
              },
              {
                instance_id: 2,
                name: "Frank",
                as_config: { ability_scores: {}, proficiencies: {} },
                hp: { value: 14 }
              }
            ]
          },
          results: {
            total_runs: 2,
            character_victories: 0,
            monster_victories: 2,
            other_victories: 0,
            average_rounds: 13.5,
            win_rate_percentage: 100,
            individual_results: [
              {
                run_id: 1,
                victory_status: "monsters",
                rounds: 16,
                logs: [
                  { id: "e1", round: 1, type: "round", data: { note: "Start" } },
                  { id: "e1-2", round: 1, attack_name: "Longsword", attack_roll: 10, target: "Monster", success: true }
                ]
              },
              {
                run_id: 2,
                victory_status: "monsters",
                rounds: 11,
                logs: [
                  { id: "e2", round: 1, type: "round", data: { note: "Start" } }
                ]
              }
            ],
            initial_state: {
              "0": { state: { current_hp: 119, max_hp: 119, conditions: {} } },
              "1": { state: { current_hp: 14, max_hp: 14, conditions: {} } },
              "2": { state: { current_hp: 165, max_hp: 165, conditions: {} } }
            }
          }
        },
        simulation_id: simulationId,
        status: "completed",
        updated_at: "2026-01-19T16:02:10.135146Z"
      });
    });

    await promise;
  });
});
