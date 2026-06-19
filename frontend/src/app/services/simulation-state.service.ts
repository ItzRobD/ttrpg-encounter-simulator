import { Injectable, signal } from '@angular/core';
import { SimulationResult, SimulationLog } from '../models';

@Injectable({
  providedIn: 'root',
})
export class SimulationStateService {
  private readonly _simulationResult = signal<SimulationResult | null>(null);
  readonly simulationResult = this._simulationResult.asReadonly();

  private readonly _loading = signal(false);
  readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  readonly error = this._error.asReadonly();

  private readonly _selectedSimulationLog = signal<SimulationLog | null>(null);
  readonly selectedSimulationLog = this._selectedSimulationLog.asReadonly();

  setSimulationResult(result: SimulationResult | null): void {
    this._simulationResult.set(result);
  }

  setLoading(loading: boolean): void {
    this._loading.set(loading);
  }

  setError(error: string | null): void {
    this._error.set(error);
  }

  setSelectedSimulationLog(log: SimulationLog | null): void {
    this._selectedSimulationLog.set(log);
  }
}
