import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Actor, ActorSummary, ApiResponse} from '../models';
import {environment} from '../../environments/environment';
import {catchError, map, Observable, of, retry, tap, throwError} from 'rxjs';

import {MapperService} from './mapper.service';
import {ActorService} from './actor.service.interface';
import {CustomContentService} from './custom-content.service';

@Injectable({
  providedIn: 'root',
})
export class MonsterService implements ActorService<Actor, ActorSummary> {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/monsters`;

  private readonly mapperService = inject(MapperService);
  private readonly customContentService = inject(CustomContentService);

  private readonly _monsters = signal<Actor[]>([]);
  public readonly monsters = computed(() => {
    return [...(this.customContentService.customMonsters() as unknown as Actor[]), ...this._monsters()];
  });

  private readonly _summaries = signal<ActorSummary[]>([]);
  public readonly summaries = computed(() => {
    const customSummaries: ActorSummary[] = (this.customContentService.customMonsters() as unknown as Actor[]).map(m => ({
      id: m.id,
      name: m.name,
      isCustom: true,
      cr: m.metadata?.cr || 0,
      type: (m.metadata?.type as any) || 'Unknown',
      size: (m.metadata?.size as any) || 'Medium',
      ac: m.ac || 0,
      isLegendary: !!m.metadata?.isLegendary,
      isSpellcaster: !!m.metadata?.spellcasterMetadata?.isSpellcaster,
      isInnateCaster: !!m.metadata?.spellcasterMetadata?.isInnateCaster
    } as any));

    return [...customSummaries, ...this._summaries()];
  });

  private readonly _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  private readonly _selectedActor = signal<Actor | null>(null);
  public readonly selectedActor = this._selectedActor.asReadonly();

  // Selected Monster alias
  public readonly selectedMonster = this.selectedActor;

  constructor() {
    // Listen for cloud content changes to refresh the UI
    this.customContentService.apiContentChange$.subscribe(({ type }) => {
      if (type === 'monsters') {
        this.getSummaries(true).subscribe();
      }
    });
  }

  // Backward compatibility method
  selectMonster(monster: Actor | null): void {
    this.selectActor(monster);
  }

  deleteMonster(id: string | number): Observable<void> {
    return this.customContentService.deleteActor('monsters', id).pipe(
      tap(() => {
        if (this._selectedActor()?.id === id) {
          this._selectedActor.set(null);
        }
      })
    );
  }

  getSummaries(forceRefresh = false): Observable<ActorSummary[]> {
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
          let rawData: any[] = [];
          const responseData = (response as any).data || response;

          if (Array.isArray(responseData)) {
            rawData = responseData;
          } else if (responseData && typeof responseData === 'object') {
            rawData = Object.values(responseData);
          }

          return rawData.map((m: any) => {
            if (m.metadata) {
              // Actor style summary
              return {
                id: m.id || m.ID,
                name: m.name,
                isCustom: !!m.isCustom,
                cr: m.metadata.cr,
                type: m.metadata.type,
                size: m.metadata.size,
                ac: m.ac || 0,
                isLegendary: !!m.metadata.isLegendary,
                isSpellcaster: !!m.metadata.spellcasterMetadata?.isSpellcaster,
                isInnateCaster: !!m.metadata.spellcasterMetadata?.isInnateCaster
              } as any;
            }
            return m as ActorSummary;
          });
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

  selectActorByID(id: string): Observable<Actor> {
    // Try finding in custom monsters first
    const customMonster = this.customContentService.customMonsters().find(m => m.id.toString() === id);
    if (customMonster) {
      const actor = customMonster as unknown as Actor;
      this._selectedActor.set(actor);
      return of(actor);
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http.get<unknown>(`${this.apiUrl}/${id}`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => {
        const data = (response as any)?.data || response;
        return data as Actor;
      }),
      tap((monster) => {
        this._selectedActor.set(monster);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        this._error.set(`Failed to load monster with ID ${id}.`);
        return throwError(() => err);
      })
    );
  }

  /**
   * Creates a new monster.
   */
  createMonster(monster: Partial<Actor>): Observable<Actor> {
    return this.http.post<Actor>(this.apiUrl, monster).pipe(
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

  selectActor(actor: Actor | null): void {
    this._selectedActor.set(actor);
  }
}
