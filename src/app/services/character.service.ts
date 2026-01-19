import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {MapperService} from './mapper.service';
import {Character, CharacterSummary, Class, EquipmentSummary, Race, WeaponSlotData, Spell, Armor} from '../models';
import {catchError, forkJoin, Observable, of, retry, tap, throwError} from 'rxjs';
import {map, switchMap} from 'rxjs/operators';
import {EntityService} from './entity.service.interface';
import { CustomContentService } from './custom-content.service';
import { EquipmentService } from './equipment.service';
import { SpellsService } from './spells.service';

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
  private readonly spellsService = inject(SpellsService);

  private readonly _characters = signal<Character[]>([]);
  public readonly characters = computed(() => {
    return [...this.customContentService.customCharacters(), ...this._characters()];
  });

  private readonly _summaries = signal<CharacterSummary[]>([]);
  public readonly summaries = computed(() => {
    const customSummaries: CharacterSummary[] = this.customContentService.customCharacters().map(c => {
      let armorName = undefined;
      const eq = c.equipment;
      if (eq?.armorId) {
        const armorSummary = (this.equipmentService.summaries() as EquipmentSummary[]).find((s: EquipmentSummary) =>
          s.id.toString() === eq.armorId!.toString() && (s.type === 'Armor' || s.type === 'Shield')
        );
        armorName = armorSummary ? armorSummary.name : `Armor #${eq.armorId}`;
      }

      if (eq?.hasShieldEquipped) {
        let shieldName = 'Shield';
        if (eq.shieldId) {
          const shieldSummary = (this.equipmentService.summaries() as EquipmentSummary[]).find((s: EquipmentSummary) =>
            s.id.toString() === eq.shieldId!.toString() && s.type === 'Shield'
          );
          shieldName = shieldSummary ? shieldSummary.name : `Shield #${eq.shieldId}`;
        }
        armorName = armorName ? `${armorName} (+ ${shieldName})` : shieldName;
      }

      const weaponNames: string[] = [];
      if (eq) {
        const primaryIds = (eq.primarySlot || []).map(w => w.weaponId);
        const secondaryIds = (eq.secondarySlot || []).map(w => w.weaponId);
        const rangedIds = (eq.rangedSlot || []).map(w => w.weaponId);

        const allWeaponIds = [...primaryIds, ...secondaryIds, ...rangedIds];

        allWeaponIds.forEach(id => {
          const ws = (this.equipmentService.summaries() as EquipmentSummary[]).find((s: EquipmentSummary) =>
            s.id.toString() === id.toString() && s.type === 'Weapon'
          );
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
            const mapped = this.mapperService.mapKeys(c) as CharacterSummary;
            // Inject display names for components that still expect .race and .class
            mapped.race = this.mapperService.getRaceName(mapped.raceId);
            mapped.class = this.mapperService.getClassName(mapped.classId);

            // Correctly resolve armor and weapon names for the summary
            if (mapped.armorId) {
              const armor = this.equipmentService.summaries().find(s =>
                s.id.toString() === mapped.armorId?.toString() && (s.type === 'Armor' || s.type === 'Shield')
              );
              mapped.armorName = armor ? armor.name : `Armor #${mapped.armorId}`;
            }

            // Also check for shield in summary mapping if it's there
            const raw = c as any;
            const hasShield = !!(raw.equipment?.has_shield_equipped || raw.has_shield_equipped);
            const shieldId = raw.equipment?.shield_id || raw.shield_id;

            if (hasShield) {
              let sName = 'Shield';
              if (shieldId) {
                const shield = this.equipmentService.summaries().find(s =>
                  s.id.toString() === shieldId.toString() && s.type === 'Shield'
                );
                sName = shield ? shield.name : `Shield #${shieldId}`;
              }
              mapped.armorName = mapped.armorName ? `${mapped.armorName} (+ ${sName})` : sName;
            }

            if (mapped.weaponIds && Array.isArray(mapped.weaponIds)) {
              mapped.weapons = mapped.weaponIds.map(id => {
                const weapon = this.equipmentService.summaries().find(s =>
                  s.id.toString() === id.toString() && s.type === 'Weapon'
                );
                return weapon ? weapon.name : `Weapon #${id}`;
              });
            }

            return mapped;
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
      return this.hydrateCharacter(customCharacter).pipe(
        tap(hydrated => this._selectedEntity.set(hydrated))
      );
    }

    this._loading.set(true);
    this._error.set(null);
    return this.http.get<unknown>(`${this.apiUrl}/${id}`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => {
        const mapped = this.mapperService.mapKeys(response) as Character;
        mapped.race = this.mapperService.getRaceName(mapped.raceId);
        mapped.class = this.mapperService.getClassName(mapped.classId);
        return mapped;
      }),
      switchMap(character => this.hydrateCharacter(character)),
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

  /**
   * Hydrates character spells and equipment using respective services
   */
  private hydrateCharacter(character: Character): Observable<Character> {
    const hydrationTasks: Observable<unknown>[] = [];

    // 1. Hydrate Spells
    const spellcasting = character.spellcasting;
    if (spellcasting && spellcasting.spellIds && spellcasting.spellIds.length > 0) {
      // If not already hydrated
      if (!spellcasting.spells || spellcasting.spells.length !== spellcasting.spellIds.length) {
        const spellRequests = spellcasting.spellIds.map(id =>
          this.spellsService.selectSpellByID(id.toString()).pipe(
            catchError(err => {
              console.error(`Failed to hydrate spell ${id}`, err);
              return of(null);
            })
          )
        );
        hydrationTasks.push(forkJoin(spellRequests).pipe(
          tap(spells => {
            character.spellcasting!.spells = spells.filter((s): s is Spell => s !== null);
          })
        ));
      }
    }

    // 2. Hydrate Armor
    if (character.equipment?.armorId && !character.equipment.armor) {
      hydrationTasks.push(this.equipmentService.selectItemByID(character.equipment.armorId.toString(), 'Armor').pipe(
        tap(armor => {
          character.equipment!.armor = armor as Armor;
        }),
        catchError(err => {
          console.error(`Failed to hydrate armor ${character.equipment?.armorId}`, err);
          return of(null);
        })
      ));
    }

    // 3. Hydrate Shield
    if (character.equipment?.shieldId && !character.equipment.shield) {
      hydrationTasks.push(this.equipmentService.selectItemByID(character.equipment.shieldId.toString(), 'Shield').pipe(
        tap(shield => {
          character.equipment!.shield = shield as Armor;
        }),
        catchError(err => {
          console.error(`Failed to hydrate shield ${character.equipment?.shieldId}`, err);
          return of(null);
        })
      ));
    }

    if (hydrationTasks.length === 0) {
      return of(character);
    }

    return forkJoin(hydrationTasks).pipe(
      map(() => character)
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
