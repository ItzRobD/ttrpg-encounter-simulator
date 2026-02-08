import { Signal } from '@angular/core';
import { Observable } from 'rxjs';
import { Actor } from '../models';

export interface ActorService<T, S> {
  summaries: Signal<S[]>;
  loading: Signal<boolean>;
  error: Signal<string | null>;
  selectedActor: Signal<T | null>;

  getSummaries(forceRefresh?: boolean): Observable<S[]>;
  selectActorByID(id: string): Observable<T>;
  selectActor(actor: T | null): void;
  clearError(): void;
}
