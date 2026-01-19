import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {MapperService} from './mapper.service';
import {Character, CharacterSummary, Class, EquipmentSummary, Race, WeaponSlotData} from '../models';
import {catchError, Observable, of, retry, tap, throwError} from 'rxjs';
import {map} from 'rxjs/operators';
import {EntityService} from './entity.service.interface';
import { CustomContentService } from './custom-content.service';
import { EquipmentService } from './equipment.service';

import { ApiResponse } from '../models';

@Injectable({
  providedIn: 'root',
})
export class CharacterService implements EntityService<Character, CharacterSummary> {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/characters`;

  private readonly mapperService = inject(MapperService);
  private readonly customContentService = inject(CustomContentService);

  private readonly equipmentService = inject(EquipmentService);

  private readonly _characters = signal<Character[]>([]);
  public readonly characters = computed(() => {
    return [...this.customContentService.customCharacters(), ...this._characters()];
  });

  private readonly _summaries = signal<CharacterSummary[]>([]);
  public readonly summaries = computed(() => {
    const customSummaries: CharacterSummary[] = this.customContentService.customCharacters().map(c => {
      let armorName = undefined;
      if (c.equipment?.armorId) {
        const armorSummary = (this.equipmentService.summaries() as EquipmentSummary[]).find((s: EquipmentSummary) => s.id.toString() === c.equipment!.armorId!.toString());
        armorName = armorSummary ? armorSummary.name : `Armor #${c.equipment!.armorId}`;
      }

      const weaponNames: string[] = [];
      const eq = c.equipment;
      if (eq) {
        const weaponIds = [
          ...(eq.primarySlot || []).map(w => w.weaponId),
          ...(eq.secondarySlot || []).map(w => w.weaponId),
          ...(eq.rangedSlot || []).map(w => w.weaponId),
        ];
        weaponIds.forEach(id => {
          const ws = (this.equipmentService.summaries() as EquipmentSummary[]).find((s: EquipmentSummary) => s.id.toString() === id.toString());
          weaponNames.push(ws ? ws.name : `Weapon #${id}`);
        });
      }

      return {
        id: c.id,
        name: c.name,
        isCustom: true,
        race: c.race,
        class: c.class,
        level: c.level,
        classId: c.classId,
        raceId: c.raceId,
        isSpellcaster: !!c.spellcasting,
        armorName: armorName,
        weapons: weaponNames
      };
    });
    return [...customSummaries, ...this._summaries()];
  });

  private readonly _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  private readonly _selectedEntity = signal<Character | null>(null);
  public readonly selectedEntity = this._selectedEntity.asReadonly();

  // Selected Character alias
  public readonly selectedCharacter = this.selectedEntity;

  // Deprecated naming
  public get characterSummaries() { return this.summaries; }

  getSummaries(forceRefresh = false): Observable<CharacterSummary[]> {
    if (!forceRefresh && this._summaries().length > 0) {
      return of(this._summaries());
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http.get<ApiResponse<unknown>>(`${this.apiUrl}/summaries`)
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
              rawData = Array.isArray(data) ? data : Object.values(data as Record<string, unknown>);
            } else if (response && 'characters' in (response as unknown as Record<string, unknown>)) {
              rawData = Object.values((response as unknown as Record<string, Record<string, unknown>>)['characters']);
            }
          }

          if (rawData.length === 0 && Array.isArray(response)) {
            rawData = response as unknown[];
          }

          return rawData.map((c) => {
            const charData = c as Record<string, unknown>;
            const mapped = this.mapperService.mapKeys(c) as Record<string, unknown>;

            // Map IDs to Enums (1-based index)
            if (charData['race_id']) {
              const raceId = charData['race_id'] as number;
              mapped['race'] = Object.values(Race)[raceId - 1] as Race;
              mapped['raceId'] = raceId;
            }
            if (charData['class_id']) {
              const classId = charData['class_id'] as number;
              mapped['class'] = Object.values(Class)[classId - 1] as Class;
              mapped['classId'] = classId;
            }

            return mapped as unknown as CharacterSummary;
          });
        }),
        tap((summaries) => {
          this._summaries.set(summaries);
          this._loading.set(false);
        }),
        catchError((err) => {
          this._loading.set(false);
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
    // Try finding in custom characters first
    const customCharacter = this.customContentService.customCharacters().find(c => c.id.toString() === id);
    if (customCharacter) {
      this._selectedEntity.set(customCharacter);
      return of(customCharacter);
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http.get<unknown>(`${this.apiUrl}/${id}`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => {
        const respRecord = response as Record<string, unknown>;
        const mapped = this.mapperService.mapKeys(response) as unknown as Character;

        // Map IDs to Enums if they are still IDs
        if (respRecord['race_id'] && typeof mapped.race !== 'string') {
          const raceId = respRecord['race_id'] as number;
          mapped.race = Object.values(Race)[raceId - 1] as Race;
          mapped.raceId = raceId;
        }
        if (respRecord['class_id'] && typeof mapped.class !== 'string') {
          const classId = respRecord['class_id'] as number;
          mapped.class = Object.values(Class)[classId - 1] as Class;
          mapped.classId = classId;
        }

        return mapped;
      }),
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

  deleteCharacter(id: string | number): Observable<void> {
    this._loading.set(true);
    return this.customContentService.deleteEntity('characters', id).pipe(
      tap(() => {
        if (this._selectedEntity()?.id === id) {
          this._selectedEntity.set(null);
        }
        this._loading.set(false);
      }),
      catchError(err => {
        this._loading.set(false);
        this._error.set(`Failed to delete character: ${err.message}`);
        return throwError(() => err);
      })
    );
  }

  clearError(): void {
    this._error.set(null);
  }
}
