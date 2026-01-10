import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { SimulationService } from './simulation.service';
import { CombatantService } from './combatant.service';
import { MapperService } from './mapper.service';
import { environment } from '../../environments/environment';
import { Entity } from '../models';
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
    abilityScores: {} as any,
    abilityScoreProficiency: {} as any,
    state: {} as any
  };

  beforeEach(() => {
    combatantServiceSpy = {
      combatants: signal([mockCombatant]),
      count: signal(1)
    };

    mapperServiceSpy = {
      mapKeys: (obj: any) => {
        const result: any = {};
        Object.keys(obj).forEach(key => {
          result[key.toLowerCase()] = obj[key];
        });
        return result;
      }
    };

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        SimulationService,
        { provide: CombatantService, useValue: combatantServiceSpy },
        { provide: MapperService, useValue: mapperServiceSpy }
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
    service.runSimulation();
    httpMock.expectNone(`${environment.apiUrl}/simulate`);
    expect(service.loading()).toBe(false);
  });

  it('should run simulation, decompress and update result on success', async () => {
    // Mock data
    const mockJson = [{ ID: '1', Type: 'round' }];
    const jsonString = JSON.stringify(mockJson);
    const encoder = new TextEncoder();
    const data = encoder.encode(jsonString);

    // Compress data using CompressionStream (simulating backend)
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(data);
        controller.close();
      }
    }).pipeThrough(new CompressionStream('gzip'));

    const arrayBuffer = await new Response(stream).arrayBuffer();

    service.runSimulation();

    expect(service.loading()).toBe(true);

    const req = httpMock.expectOne(`${environment.apiUrl}/simulate`);
    expect(req.request.method).toBe('POST');
    expect(req.request.responseType).toBe('arraybuffer');

    req.flush(arrayBuffer);

    // We need to wait for the async decompression to finish
    // Since it's a promise inside an observable, we might need a small delay or use fakeAsync
    await new Promise(resolve => setTimeout(resolve, 50));

    expect(service.loading()).toBe(false);
    expect(service.simulationResult()).not.toBeNull();
    expect(service.simulationResult()?.count).toBe(1);
    expect(service.simulationResult()?.events[0].id).toBe('1');
  });

  it('should handle errors and stop loading', () => {
    service.runSimulation();

    const req = httpMock.expectOne(`${environment.apiUrl}/simulate`);
    req.error(new ErrorEvent('Network error'));

    expect(service.loading()).toBe(false);
    expect(service.simulationResult()).toBeNull();
  });
});
