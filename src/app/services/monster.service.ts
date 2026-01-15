import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Monster, MonsterSummary} from '../models';
import { environment } from '../../environments/environment';
import { catchError, map, Observable, of, retry, tap, throwError } from 'rxjs';

import { MapperService } from './mapper.service';
import {EntityService} from './entity.service.interface';

import { ApiResponse } from '../models';

@Injectable({
  providedIn: 'root',
})
export class MonsterService implements EntityService<Monster, MonsterSummary> {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/monsters`;

  private readonly mapperService = inject(MapperService);

  private readonly _monsters = signal<Monster[]>([]);
  public readonly monsters = this._monsters.asReadonly();

  private readonly _summaries = signal<MonsterSummary[]>([]);
  public readonly summaries = this._summaries.asReadonly();

  private readonly _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  private readonly _selectedEntity = signal<Monster | null>(null);
  public readonly selectedEntity = this._selectedEntity.asReadonly();

  // Selected Monster alias
  public readonly selectedMonster = this.selectedEntity;

  // Deprecated naming for backward compatibility if needed, but we'll update usages
  public get monsterSummaries() { return this.summaries; }
  public get selectedBestiaryMonster() { return this.selectedMonster; }

  // Backward compatibility method
  selectMonster(monster: Monster | null): void {
    this.selectEntity(monster);
  }

  getSummaries(forceRefresh = false): Observable<MonsterSummary[]> {
    if (!forceRefresh && this._summaries().length > 0) {
      return of(this._summaries());
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http
      .get<ApiResponse<unknown>>(`${this.apiUrl}/summaries`)
      .pipe(
        retry({
          count: environment.httpRetryCount,
          delay: environment.httpRetryDelay
        }),
        map((response) => {
          let rawData: unknown[] = [];
          if (response && typeof response === 'object') {
            const data = response.data;
            if (data) {
              if (Array.isArray(data)) {
                rawData = data;
              } else {
                // Handle dictionary format in data
                rawData = Object.values(data as Record<string, unknown>);
              }
            } else if (response && 'monsters' in (response as unknown as Record<string, unknown>)) {
              rawData = Object.values((response as unknown as Record<string, Record<string, unknown>>)['monsters']);
            }
          }

          if (rawData.length === 0 && Array.isArray(response)) {
            rawData = response as unknown[];
          }

          return rawData.map((m) => this.mapperService.mapKeys(m)) as MonsterSummary[];
        }),
        tap((summaries) => {
          this._summaries.set(summaries);
          this._loading.set(false);
        }),
        catchError((err) => {
          this._loading.set(false);
          this._error.set('Failed to load monster summaries. Please try again later.');
          return throwError(() => err);
        })
      );
  }

  // Keep old method name for compatibility during migration
  getMonsterSummaries(forceRefresh = false): Observable<MonsterSummary[]> {
    return this.getSummaries(forceRefresh);
  }

  /**
   * Fetches all monsters from the backend.
   */
  getMonsters(): Observable<Monster[]> {
    this._loading.set(true);
    this._error.set(null);
    return this.http.get<ApiResponse<unknown>>(this.apiUrl).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map((response) => {
        let rawData: unknown[] = [];
        if (response && response.data) {
          rawData = Array.isArray(response.data) ? response.data : Object.values(response.data as Record<string, unknown>);
        } else if (response && 'monsters' in (response as unknown as Record<string, unknown>)) {
          rawData = Object.values((response as unknown as Record<string, Record<string, unknown>>)['monsters']);
        }
        return rawData.map(m => this.mapperService.mapKeys(m) as Monster);
      }),
      tap((monsters) => {
        this._monsters.set(monsters);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        this._error.set('Failed to load monsters. Please try again later.');
        return throwError(() => err);
      })
    );
  }

  selectEntityByID(id: string): Observable<Monster> {
    this._loading.set(true);
    this._error.set(null);
    return this.http.get<unknown>(`${this.apiUrl}/${id}`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => this.mapperService.mapKeys(response) as Monster),
      tap((monster) => {
        this._selectedEntity.set(monster);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        this._error.set(`Failed to load monster with ID ${id}.`);
        return throwError(() => err);
      })
    );
  }

  // Keep old method name for compatibility during migration
  selectBestiaryMonsterByID(id: string): Observable<Monster> {
    return this.selectEntityByID(id);
  }

  /**
   * Creates a new monster.
   */
  createMonster(monster: Partial<Monster>): Observable<Monster> {
    return this.http.post<Monster>(this.apiUrl, monster).pipe(
      tap((newMonster) => {
        this._monsters.update((m) => [...m, newMonster]);
      })
    );
  }

  /**
   * Updates an existing monster.
   */
  // updateMonster(id: string, monster: Partial<Monster>): Observable<Monster> {
  //   return this.http.put<Monster>(`${this.apiUrl}/${id}`, monster).pipe(
  //     tap((updatedMonster) => {
  //       this._monsters.update((m) =>
  //         m.map((mon) => (mon.id === id ? updatedMonster : mon))
  //       );
  //       if (this._selectedMonster()?.id === id) {
  //         this._selectedMonster.set(updatedMonster);
  //       }
  //     })
  //   );
  // }

  /**
   * Deletes a monster.
   */
  // deleteMonster(id: string): Observable<void> {
  //   return this.http.delete<void>(`${this.apiUrl}/${id}`).pipe(
  //     tap(() => {
  //       this._monsters.update((m) => m.filter((mon) => mon.id !== id));
  //       if (this._selectedMonster()?.id === id) {
  //         this._selectedMonster.set(null);
  //       }
  //     })
  //   );
  // }

  /**
   * Sets the currently selected monster.
   */
  /**
   * Clears any current error state.
   */
  clearError(): void {
    this._error.set(null);
  }

  selectEntity(entity: Monster | null): void {
    this._selectedEntity.set(entity);
  }
}
