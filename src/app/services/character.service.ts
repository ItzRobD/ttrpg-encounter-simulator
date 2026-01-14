import {inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {MapperService} from './mapper.service';
import {Character, CharacterSummary, Class, Race} from '../models';
import {catchError, Observable, of, retry, tap, throwError} from 'rxjs';
import {map} from 'rxjs/operators';
import {EntityService} from './entity.service.interface';

@Injectable({
  providedIn: 'root',
})
export class CharacterService implements EntityService<Character, CharacterSummary> {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/characters`;

  private readonly mapperService = inject(MapperService);

  private readonly _characters = signal<Character[]>([]);
  public readonly characters = this._characters.asReadonly();

  private readonly _summaries = signal<CharacterSummary[]>([]);
  public readonly summaries = this._summaries.asReadonly();

  private readonly _loadingSummaries = signal(false);
  public readonly loadingSummaries = this._loadingSummaries.asReadonly();

  private readonly _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  private readonly _selectedEntity = signal<Character | null>(null);
  public readonly selectedEntity = this._selectedEntity.asReadonly();

  // Deprecated naming
  public get characterSummaries() { return this.summaries; }
  public get selectedCharacter() { return this.selectedEntity; }

  getSummaries(forceRefresh = false): Observable<CharacterSummary[]> {
    if (!forceRefresh && this._summaries().length > 0) {
      return of(this._summaries());
    }

    this._loadingSummaries.set(true);
    this._error.set(null);
    return this.http.get<any>(`${this.apiUrl}/summaries`)
      .pipe(
        retry({
          count: environment.httpRetryCount,
          delay: environment.httpRetryDelay
        }),
        map((response) => {
          let rawData: any[] = [];
          if (response && typeof response === 'object') {
            if (response.data && Array.isArray(response.data)) {
              rawData = response.data;
            } else if (response.characters) {
              rawData = Object.values(response.characters);
            }
          } else if (Array.isArray(response)) {
            rawData = response;
          }

          return rawData.map((c: any) => {
            const mapped = this.mapperService.mapKeys(c) as any;

            // Map IDs to Enums (1-based index)
            if (c.race_id) {
              mapped.race = Object.values(Race)[c.race_id - 1] as Race;
            }
            if (c.class_id) {
              mapped.class = Object.values(Class)[c.class_id - 1] as Class;
            }

            return mapped as CharacterSummary;
          });
        }),
        tap((summaries) => {
          this._summaries.set(summaries);
          this._loadingSummaries.set(false);
        }),
        catchError((err) => {
          this._loadingSummaries.set(false);
          this._error.set('Failed to load character summaries. Please try again later.');
          return of([]);
        })
      );
  }

  // Deprecated
  getCharacterSummaries(forceRefresh = false): Observable<CharacterSummary[]> {
    return this.getSummaries(forceRefresh);
  }

  selectEntityByID(id: string): Observable<Character> {
    this._loading.set(true);
    this._error.set(null);
    return this.http.get<any>(`${this.apiUrl}/${id}`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => this.mapperService.mapKeys(response) as Character),
      tap((character) => {
        this._selectedEntity.set(character);
        this._loading.set(false);
      }),
      catchError((err) => {
        this._loading.set(false);
        this._error.set(`Failed to load character with ID ${id}.`);
        return throwError(() => err);
      })
    );
  }

  // Deprecated
  selectCharacterByID(id: string): Observable<Character> {
    return this.selectEntityByID(id);
  }

  selectEntity(entity: Character | null): void {
    this._selectedEntity.set(entity);
  }

  clearError(): void {
    this._error.set(null);
  }
}
