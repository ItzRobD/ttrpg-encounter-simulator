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
        map((resp) => this.mapSummaries(resp)),
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

  getSummariesByClass(classId: string | number): Observable<SpellSummary[]> {
    this._loading.set(true);
    this._error.set(null);
    return this.http.get<ApiResponse<unknown>>(`${this.apiUrl}/summaries/class/${classId}`)
      .pipe(
        retry({
          count: environment.httpRetryCount,
          delay: environment.httpRetryDelay,
        }),
        map((resp) => this.mapSummaries(resp)),
        tap(() => this._loading.set(false)),
        catchError((err) => {
          this._loading.set(false);
          this._error.set('Failed to load class spell summaries.');
          return throwError(() => err);
        })
      );
  }

  private mapSummaries(resp: ApiResponse<unknown> | unknown[]): SpellSummary[] {
    const rawData = resp;

    if (Array.isArray(rawData)) {
      return rawData as SpellSummary[];
    }

    if (rawData && typeof rawData === 'object') {
      // If it's a dictionary of spells
      return Object.values(rawData as any as Record<string, SpellSummary>);
    }

    return [];
  }

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
        const mapped = response;
        const spells = Array.isArray(mapped) ? mapped : (mapped && typeof mapped === 'object' ? Object.values(mapped) : []);
        return spells as Spell[];
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

  selectActorByID(id: string): Observable<Spell> {
    // Try finding in custom spells first
    const customSpell = this.customContentService.customSpells().find(s => s.id.toString() === id);
    if (customSpell) {
      this._selectedSpell.set(customSpell);
      return of(customSpell);
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http.get<ApiResponse<Record<string, Spell>>>(`${this.apiUrl}/${id}`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => {
        // Handle dictionary response { "119": { ... } }
        // Note: MapperService.mapKeys might have already unwrapped the outer 'data' envelope
        // or the response might still be { data: { "119": { ... } } } if MapperService didn't unwrap it.

        let data = (response as any)?.data || response;

        // If it's a dictionary of spells
        if (data && typeof data === 'object' && !Array.isArray(data) && !data.id && !data.name) {
          const keys = Object.keys(data);
          // If the key is the ID, return the value
          if (keys.includes(id)) {
            return data[id];
          }
          // Fallback: if there's only one key and it's numeric/id-like
          if (keys.length === 1 && !isNaN(Number(keys[0]))) {
            return data[keys[0]];
          }
        }
        return data as Spell;
      }),
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

  selectActor(actor: Spell | null): void {
    this._selectedSpell.set(actor);
  }

  deleteSpell(id: string | number): Observable<void> {
    return this.customContentService.deleteActor('spells', id).pipe(
      tap(() => {
        if (this._selectedSpell()?.id === id) {
          this._selectedSpell.set(null);
        }
      })
    );
  }
}
