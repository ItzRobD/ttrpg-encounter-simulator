import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {MapperService} from './mapper.service';
import { Actor, ActorSummary, Armor } from '../models';
import {catchError, forkJoin, Observable, of, retry, tap, throwError} from 'rxjs';
import {map, switchMap} from 'rxjs/operators';
import {ActorService} from './actor.service.interface';
import { CustomContentService } from './custom-content.service';
import { EquipmentService } from './equipment.service';
import { SpellsService } from './spells.service';

import { ApiResponse } from '../models';

@Injectable({
  providedIn: 'root',
})
export class CharacterService implements ActorService<Actor, ActorSummary> {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/characters`;

  private readonly mapperService = inject(MapperService);
  private readonly customContentService = inject(CustomContentService);

  private readonly equipmentService = inject(EquipmentService);
  private readonly spellsService = inject(SpellsService);

  private readonly _characters = signal<Actor[]>([]);
  public readonly characters = computed(() => {
    return [...(this.customContentService.customCharacters() as unknown as Actor[]), ...this._characters()];
  });

  private readonly _summaries = signal<ActorSummary[]>([]);
  public readonly summaries = computed(() => {
    const customSummaries: ActorSummary[] = (this.customContentService.customCharacters() as unknown as Actor[]).map(c => {
      let armorName = undefined;
      const eq = c.equipment;
      const metadata = c.metadata;

      if (eq?.armorId) {
        const armorSummary = (this.equipmentService.summaries() as any[]).find((s: any) =>
          s.id.toString() === eq.armorId!.toString() && (s.type === 'Armor' || s.type === 'Shield')
        );
        armorName = armorSummary ? armorSummary.name : `Armor #${eq.armorId}`;
      }

      if (eq?.hasShieldEquipped) {
        let shieldName = 'Shield';
        if (eq.shieldId) {
          const shieldSummary = (this.equipmentService.summaries() as any[]).find((s: any) =>
            s.id.toString() === eq.shieldId!.toString() && s.type === 'Shield'
          );
          shieldName = shieldSummary ? shieldSummary.name : `Shield #${eq.shieldId}`;
        }
        armorName = armorName ? `${armorName} (+ ${shieldName})` : shieldName;
      }

      const weaponNames: string[] = [];
      if (eq) {
        const primaryIds = (eq.primarySlot || []).map((w: any) => w.weaponId);
        const secondaryIds = (eq.secondarySlot || []).map((w: any) => w.weaponId);
        const rangedIds = (eq.rangedSlot || []).map((w: any) => w.weaponId);

        const allWeaponIds = [...primaryIds, ...secondaryIds, ...rangedIds];

        allWeaponIds.forEach(id => {
          const ws = (this.equipmentService.summaries() as any[]).find((s: any) =>
            s.id.toString() === id.toString() && s.type === 'Weapon'
          );
          weaponNames.push(ws ? ws.name : `Weapon #${id}`);
        });
      }

      return {
        id: c.id,
        name: c.name,
        isCustom: true,
        race: this.mapperService.getRaceName(metadata?.raceId || 0),
        class: this.mapperService.getClassName(metadata?.classId || 0),
        level: metadata?.level || 1,
        classId: metadata?.classId || 0,
        raceId: metadata?.raceId || 0,
        isSpellcaster: !!metadata?.spellcasterMetadata?.isSpellcaster,
        isInnateCaster: !!metadata?.spellcasterMetadata?.isInnateCaster,
        armorName: armorName,
        weapons: weaponNames
      } as any;
    });
    return [...customSummaries, ...this._summaries()];
  });

  private readonly _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  private readonly _selectedActor = signal<Actor | null>(null);
  public readonly selectedActor = this._selectedActor.asReadonly();

  // Selected Character alias
  public readonly selectedCharacter = this.selectedActor;

  // Deprecated naming
  public get characterSummaries() { return this.summaries; }

  getSummaries(forceRefresh = false): Observable<ActorSummary[]> {
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
          console.log('[CharacterService] getSummaries raw response:', response);
          let rawData: any[] = [];
          const responseData = (response as any).data || response;

          if (Array.isArray(responseData)) {
            rawData = responseData;
          } else if (responseData && typeof responseData === 'object') {
            rawData = Object.values(responseData);
          }

          console.log('[CharacterService] getSummaries extracted rawData:', rawData);

          const mapped = rawData.map((c) => {
            const mapped = c as any;
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
              mapped.weapons = mapped.weaponIds.map((id: string | number) => {
                const weapon = this.equipmentService.summaries().find(s =>
                  s.id.toString() === id.toString() && s.type === 'Weapon'
                );
                return weapon ? weapon.name : `Weapon #${id}`;
              });
            }

            return mapped as ActorSummary;
          });
          console.log('[CharacterService] getSummaries mapped result:', mapped);
          return mapped;
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
  getCharacterSummaries(forceRefresh = false): Observable<ActorSummary[]> {
    return this.getSummaries(forceRefresh);
  }

  selectActorByID(id: string): Observable<Actor> {
    // Try finding in custom characters first
    const customCharacter = this.customContentService.customCharacters().find(c => c.id.toString() === id);
    if (customCharacter) {
      return this.hydrateCharacter(customCharacter as unknown as Actor).pipe(
        tap(hydrated => this._selectedActor.set(hydrated))
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
        const data = (response as any)?.data || response;
        return data as Actor;
      }),
      switchMap(character => this.hydrateCharacter(character)),
      tap((character) => {
        this._selectedActor.set(character);
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
  private hydrateCharacter(character: Actor): Observable<Actor> {
    const hydrationTasks: Observable<unknown>[] = [];

    // 1. Hydrate Spells
    const spellcasting = character.spellcasting;
    const spellIds = character.knownSpellIDs || (character.spellcasting as any)?.spellIds;

    if (spellIds && spellIds.length > 0) {
      // If not already hydrated
      if (!(spellcasting as any)?.spells || (spellcasting as any).spells.length !== (spellIds as any).length) {
        const spellRequests = (spellIds as any).map((id: any) =>
          this.spellsService.selectActorByID(id.toString()).pipe(
            catchError(err => {
              console.error(`Failed to hydrate spell ${id}`, err);
              return of(null);
            })
          )
        );
        hydrationTasks.push(forkJoin(spellRequests).pipe(
          tap((spells: any) => {
            if (!character.spellcasting) {
              (character as any).spellcasting = {
                casterType: (character.metadata?.spellcasterMetadata?.isSpellcaster ? 'full' : 'none') as any,
                casterLevel: character.metadata?.spellcasterMetadata?.spellcastingLevel || 1,
                spellSlots: {},
                spellSaveDC: 10,
                spellAttackBonus: 0
              };
            }
            (character.spellcasting as any).spells = spells.filter((s: any) => s !== null);
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
  selectCharacterByID(id: string): Observable<Actor> {
    return this.selectActorByID(id);
  }

  selectActor(actor: Actor | null): void {
    this._selectedActor.set(actor);
  }

  deleteCharacter(id: string | number): Observable<void> {
    this._loading.set(true);
    return this.customContentService.deleteActor('characters', id).pipe(
      tap(() => {
        if (this._selectedActor()?.id === id) {
          this._selectedActor.set(null);
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
