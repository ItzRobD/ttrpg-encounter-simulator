import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { SimulationService } from './simulation.service';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
import { environment } from '../../environments/environment';
import { Entity, Condition, DamageType, ResistanceType } from '../models';
import { signal } from '@angular/core';

describe('SimulationService', () => {
  let service: SimulationService;
  let httpMock: HttpTestingController;
  let combatantServiceSpy: any;
  let mapperServiceSpy: any;

  const mockCombatant: Entity = {
    id: 1,
    instanceId: 1,
    name: 'Test',
    asConfig: {
      abilityScores: { strength: 10, dexterity: 10, constitution: 10, intelligence: 10, wisdom: 10, charisma: 10 },
      proficiencies: { strength: false, dexterity: false, constitution: false, intelligence: false, wisdom: false, charisma: false }
    },
    state: {
      currentHp: 10,
      maxHp: 10,
      tempHp: 0,
      hitDie: 8,
      conditions: Object.values(Condition).reduce((acc, curr) => ({ ...acc, [curr]: false }), {} as any),
      deathSaves: { successes: 0, failures: 0 },
      resistances: Object.values(DamageType).reduce((acc, curr) => ({ ...acc, [curr]: ResistanceType.None }), {} as any),
      isStable: true,
      isDead: false,
      initiative: 10
    }
  };

  beforeEach(() => {
    combatantServiceSpy = {
      combatants: signal([mockCombatant]),
      count: signal(1)
    };

    mapperServiceSpy = {
      mapKeys: (obj: unknown) => {
        const result: Record<string, unknown> = {};
        const record = obj as Record<string, unknown>;
        Object.keys(record).forEach(key => {
          result[key.toLowerCase()] = record[key];
        });
        return result;
      },
      serializeKeys: (obj: any) => obj
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

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should start with no result and not loading', () => {
    expect(service.simulationResult()).toBeNull();
    expect(service.loading()).toBe(false);
  });

  it('should not run simulation if no combatants', () => {
    combatantServiceSpy.combatants.set([]);
    service.createSimulation();
    httpMock.expectNone(`${environment.apiUrl}/simulation/create`);
    expect(service.loading()).toBe(false);
  });

  it('should run simulation, decompress and update result on success', async () => {
    // Mock data
    const mockJson = {
      total_runs: 1,
      individual_results: [{
        run_id: 1,
        total_rounds: 1,
        victory_status: 'monsters',
        encounter_results: [{
          encounter_name: 'Encounter 1',
          victory_status: 'monsters',
          rounds: 1,
          logs: [{ id: '1', type: 'round' }]
        }]
      }]
    };

    service.createSimulation();

    expect(service.loading()).toBe(true);

    const req = httpMock.expectOne(`${environment.apiUrl}/simulation/create`);
    expect(req.request.method).toBe('POST');
    // It will be a 202 now
    req.flush({ data: { simulation_id: 'test-id', status: 'completed' } }, { status: 202, statusText: 'Accepted' });

    const statusReq = httpMock.expectOne(`${environment.apiUrl}/simulation/status/test-id`);
    statusReq.flush({
      data: {
        simulation_id: 'test-id',
        status: 'completed',
        created_at: '2026-01-18T17:46:19.202803Z',
        updated_at: '2026-01-18T17:46:19.551564Z'
      }
    });

    const resultReq = httpMock.expectOne(`${environment.apiUrl}/simulation/results/test-id`);
    resultReq.flush({
      data: {
        simulation_id: 'test-id',
        status: 'completed',
        results: mockJson
      }
    });

    // Wait for the observables to complete
    await new Promise(resolve => setTimeout(resolve, 100));

    expect(service.loading()).toBe(false);
    expect(service.simulationResult()).not.toBeNull();
    expect(service.simulationResult()?.count).toBe(1);
    expect(service.simulationResult()?.logs[0].events[0].id).toBe('1');
  });

  it('should handle 202 Accepted response and start polling', async () => {
    service.createSimulation();

    const createReq = httpMock.expectOne(`${environment.apiUrl}/simulation/create`);
    createReq.flush({ data: { simulation_id: 'test-id', status: 'pending' } }, { status: 202, statusText: 'Accepted' });

    expect(service.currentSimulationId()).toBe('test-id');
    expect(service.loading()).toBe(true);

    // First status poll
    const statusReq = httpMock.expectOne(`${environment.apiUrl}/simulation/status/test-id`);
    statusReq.flush({
      data: {
        simulation_id: 'test-id',
        status: 'completed',
        created_at: '2026-01-18T17:46:19.202803Z',
        updated_at: '2026-01-18T17:46:19.551564Z'
      }
    });

    // Result fetch
    const resultReq = httpMock.expectOne(`${environment.apiUrl}/simulation/results/test-id`);

    const mockResultJson = {
      total_runs: 0,
      individual_results: [] as any[]
    };
    resultReq.flush({
      data: {
        simulation_id: 'test-id',
        status: 'completed',
        results: mockResultJson
      }
    });

    await new Promise(resolve => setTimeout(resolve, 100));

    expect(service.loading()).toBe(false);
  });

  it('should handle simulation failure with error message from status response', async () => {
    service.createSimulation();

    const createReq = httpMock.expectOne(`${environment.apiUrl}/simulation/create`);
    createReq.flush({ data: { simulation_id: 'test-id', status: 'pending' } }, { status: 202, statusText: 'Accepted' });

    const statusReq = httpMock.expectOne(`${environment.apiUrl}/simulation/status/test-id`);
    statusReq.flush({
      data: {
        simulation_id: 'test-id',
        status: 'failed',
        error: 'Custom server error message'
      }
    });

    expect(service.error()).toBe('Custom server error message');
    expect(service.loading()).toBe(false);
  });

  it('should handle errors and stop loading', () => {
    service.createSimulation();

    const req = httpMock.expectOne(`${environment.apiUrl}/simulation/create`);
    req.error(new ErrorEvent('Network error'));

    expect(service.loading()).toBe(false);
    expect(service.simulationResult()).toBeNull();
  });
});
