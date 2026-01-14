import { Signal } from '@angular/core';
import { Observable } from 'rxjs';
import { Entity, EntitySummary } from '../models';

export interface EntityService<T extends Entity, S extends EntitySummary> {
  summaries: Signal<S[]>;
  loadingSummaries: Signal<boolean>;
  loading: Signal<boolean>;
  error: Signal<string | null>;
  selectedEntity: Signal<T | null>;

  getSummaries(forceRefresh?: boolean): Observable<S[]>;
  selectEntityByID(id: string): Observable<T>;
  selectEntity(entity: T | null): void;
  clearError(): void;
}
