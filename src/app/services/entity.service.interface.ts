import { Signal } from '@angular/core';
import { Observable } from 'rxjs';
import { Entity, EntitySummary } from '../models';

export interface EntityService<T, S> {
  summaries: Signal<S[]>;
  loading: Signal<boolean>;
  error: Signal<string | null>;
  selectedEntity: Signal<T | null>;

  getSummaries(forceRefresh?: boolean): Observable<S[]>;
  selectEntityByID(id: string): Observable<T>;
  selectEntity(entity: T | null): void;
  clearError(): void;
}
