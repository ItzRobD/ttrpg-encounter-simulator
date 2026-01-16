import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {ApiResponse, Spell, SpellSummary} from '../models';
import {catchError, map, Observable, of, retry, tap, throwError} from 'rxjs';
import {MapperService} from './mapper.service';
import { CustomContentService } from './custom-content.service';

@Injectable({
  providedIn: 'root',
})
export class SpellsService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/spells`;
  private mapperService = inject(MapperService);
  private customContentService = inject(CustomContentService);

  private _summaries = signal<SpellSummary[]>([]);
  public readonly summaries = computed(() => {
    const customSummaries: SpellSummary[] = this.customContentService.customSpells().map(s => ({
      id: s.id,
      name: s.name,
      isCustom: true,
      isConcentration: s.isConcentration,
      isRitual: s.isRitual,
      level: s.level,
      spellType: s.spellType,
      isAOE: s.isAOE,
      isTouch: s.isTouch,
      hasDC: s.hasDC
    }));
    return [...customSummaries, ...this._summaries()];
  });

  private _spells = signal<Spell[]>([]);
  public readonly spells = computed(() => {
    return [...this.customContentService.customSpells(), ...this._spells()];
  });

  private _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private _selectedSpell = signal<Spell | null>(null);
  public readonly selectedSpell = this._selectedSpell.asReadonly();

  private _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  getSummaries(forceRefresh = false): Observable<SpellSummary[]> {
    if (!forceRefresh && this._summaries().length > 0) {
      return of(this._summaries());
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http.get<ApiResponse<unknown>>(`${this.apiUrl}/summaries`)
      .pipe(
        retry({
          count: environment.httpRetryCount,
          delay: environment.httpRetryDelay,
        }),
        map((resp) => {
          let rawData: unknown[] = [];
          if (resp && typeof resp === 'object') {
            const data = resp.data;
            if (data) {
              if (Array.isArray(data)) {
                rawData = data;
              } else {
                rawData = Object.values(data as Record<string, unknown>);
              }
            } else if (resp && 'spells' in (resp as unknown as Record<string, unknown>)) {
              rawData = Object.values((resp as unknown as Record<string, Record<string, unknown>>)['spells']);
            }
          }

          if (rawData.length === 0 && Array.isArray(resp)) {
            rawData = resp as unknown[];
          }

          return rawData.map((s) => this.mapperService.mapKeys(s)) as SpellSummary[];
        }),
      tap((summaries) => {
        this._summaries.set(summaries);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        this._error.set('Failed to load spell summaries. Please try again later.');
        return throwError(() => err);
      })
      );
  };

  /**
   * Fetches all spells from the backend.
   */
  getSpells(): Observable<Spell[]> {
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
        } else if (response && 'spells' in (response as unknown as Record<string, unknown>)) {
          rawData = Object.values((response as unknown as Record<string, Record<string, unknown>>)['spells']);
        }
        return rawData.map(m => this.mapperService.mapKeys(m) as Spell);
      }),
      tap((spell) => {
        this._spells.set(spell);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        this._error.set('Failed to load spells. Please try again later.');
        return throwError(() => err);
      })
    );
    }

  selectSpellByID(id: string): Observable<Spell> {
    // Try finding in custom spells first
    const customSpell = this.customContentService.customSpells().find(s => s.id.toString() === id);
    if (customSpell) {
      this._selectedSpell.set(customSpell);
      return of(customSpell);
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http.get<unknown>(`${this.apiUrl}/${id}`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => this.mapperService.mapKeys(response) as Spell),
      tap((spell) => {
        this._selectedSpell.set(spell);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        this._error.set(`Failed to load spell with ID ${id}.`);
        return throwError(() => err);
      })
    );
  }

  clearError(): void {
    this._error.set(null);
  }

  selectSpell(spell: Spell | null): void {
    this._selectedSpell.set(spell);
  }
}
