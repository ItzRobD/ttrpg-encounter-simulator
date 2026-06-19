import {computed, inject, Injectable, signal} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {environment} from '../../environments/environment';
import {MapperService} from './mapper.service';
import { Actor, ActorSummary, Armor, Weapon, Spell, CasterType, WeaponSlotData, DataEnvelope } from '../models';
import {catchError, forkJoin, Observable, of, retry, tap, throwError} from 'rxjs';
import {map, switchMap} from 'rxjs/operators';
import {ActorService} from './actor.service.interface';
import { CustomContentService } from './custom-content.service';
import { EquipmentService } from './equipment.service';
import { SpellsService } from './spells.service';

import { ApiResponse } from '../models';

/**
 * The character summaries endpoint returns objects that are mostly ActorSummary
 * plus a few raw fields used only to resolve display names. Fields are camelCase
 * (the mapping interceptor normalizes the snake_case backend payload).
 */
interface RawCharacterSummary extends ActorSummary {
  armorId?: number | string;
  weaponIds?: (number | string)[];
  equipment?: { hasShieldEquipped?: boolean; shieldId?: number | string };
}

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
        const armorSummary = this.equipmentService.summaries().find(s =>
          s.id.toString() === eq.armorId!.toString() && (s.type === 'Armor' || s.type === 'Shield')
        );
        armorName = armorSummary ? armorSummary.name : `Armor #${eq.armorId}`;
      }

      if (eq?.hasShieldEquipped) {
        let shieldName = 'Shield';
        if (eq.shieldId) {
          const shieldSummary = this.equipmentService.summaries().find(s =>
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
          const ws = this.equipmentService.summaries().find(s =>
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
      };
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

  constructor() {
    // Listen for cloud content changes to refresh the UI
    this.customContentService.apiContentChange$.subscribe(({ type }) => {
      if (type === 'characters') {
        this.getSummaries(true).subscribe();
      }
    });
  }

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
          // Response is camelCased + 'data'-unwrapped by the mapping interceptor.
          const responseData = response.data ?? response;

          let rawData: RawCharacterSummary[] = [];
          if (Array.isArray(responseData)) {
            rawData = responseData as RawCharacterSummary[];
          } else if (responseData && typeof responseData === 'object') {
            rawData = Object.values(responseData) as RawCharacterSummary[];
          }

          const mapped = rawData.map((c) => {
            // Inject display names for components that still expect .race and .class
            c.race = this.mapperService.getRaceName(c.raceId ?? 0);
            c.class = this.mapperService.getClassName(c.classId ?? 0);

            // Correctly resolve armor and weapon names for the summary
            if (c.armorId) {
              const armor = this.equipmentService.summaries().find(s =>
                s.id.toString() === c.armorId?.toString() && (s.type === 'Armor' || s.type === 'Shield')
              );
              c.armorName = armor ? armor.name : `Armor #${c.armorId}`;
            }

            const hasShield = !!c.equipment?.hasShieldEquipped;
            const shieldId = c.equipment?.shieldId;
            if (hasShield) {
              let sName = 'Shield';
              if (shieldId) {
                const shield = this.equipmentService.summaries().find(s =>
                  s.id.toString() === shieldId.toString() && s.type === 'Shield'
                );
                sName = shield ? shield.name : `Shield #${shieldId}`;
              }
              c.armorName = c.armorName ? `${c.armorName} (+ ${sName})` : sName;
            }

            if (c.weaponIds && Array.isArray(c.weaponIds)) {
              c.weapons = c.weaponIds.map((id) => {
                const weapon = this.equipmentService.summaries().find(s =>
                  s.id.toString() === id.toString() && s.type === 'Weapon'
                );
                return weapon ? weapon.name : `Weapon #${id}`;
              });
            }

            return c as ActorSummary;
          });
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
        const data = (response as DataEnvelope<Actor>)?.data || response;
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
    const spellIds = character.knownSpellIDs || spellcasting?.spellIds;

    if (spellIds && spellIds.length > 0) {
      // If not already hydrated
      if (!spellcasting?.spells || spellcasting.spells.length !== spellIds.length) {
        const spellRequests = spellIds.map((id) =>
          this.spellsService.selectActorByID(id.toString()).pipe(
            catchError(err => {
              console.error(`Failed to hydrate spell ${id}`, err);
              return of(null);
            })
          )
        );
        hydrationTasks.push(forkJoin(spellRequests).pipe(
          tap((spells: (Spell | null)[]) => {
            if (!character.spellcasting) {
              character.spellcasting = {
                casterType: (character.metadata?.spellcasterMetadata?.isSpellcaster ? CasterType.Full : CasterType.None),
                casterLevel: character.metadata?.spellcasterMetadata?.spellcastingLevel || 1,
                spellSlots: {},
                spellSaveDC: 10,
                spellAttackBonus: 0
              };
            }
            character.spellcasting.spells = spells.filter((s): s is Spell => s !== null);
          })
        ));
      }
    }

    // 2. Hydrate Armor and Weapons from EquipmentConfigs if equipment object is missing
    if (character.equipmentConfigs && (!character.equipment || (!character.equipment.armorId && !character.equipment.primarySlot))) {
      if (!character.equipment) {
        character.equipment = { hasShieldEquipped: false };
      }

      character.equipmentConfigs.forEach(config => {
        if (config.type === 'armor') {
          character.equipment!.armorId = config.id;
        } else if (config.type === 'shield') {
          character.equipment!.shieldId = config.id;
          character.equipment!.hasShieldEquipped = true;
        } else if (config.type === 'weapon') {
          const slotData: WeaponSlotData = {
            weaponId: config.id,
            isProficient: true, // Assume proficient for custom chars
            modifiers: { attackBonus: 0, damageBonus: 0, isMagic: false, isSilvered: false, isAdamantine: false, isColdForgedIron: false }
          };
          if (config.slot === 'primary') {
            character.equipment!.primarySlot = [slotData];
          } else if (config.slot === 'secondary') {
            character.equipment!.secondarySlot = [slotData];
          } else if (config.slot === 'ranged') {
            character.equipment!.rangedSlot = [slotData];
          }
        }
      });
    }

    // 3. Hydrate Armor
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

    // 4. Hydrate Shield
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

    // 5. Hydrate Weapons
    const weaponSlots = [
      character.equipment?.primarySlot,
      character.equipment?.secondarySlot,
      character.equipment?.rangedSlot
    ].filter(slot => slot && slot.length > 0);

    weaponSlots.forEach(slot => {
      slot!.forEach(ws => {
        if (ws.weaponId && !ws.weapon) {
          hydrationTasks.push(this.equipmentService.selectItemByID(ws.weaponId.toString(), 'Weapon').pipe(
            tap(weapon => {
              ws.weapon = weapon as Weapon;
            }),
            catchError(err => {
              console.error(`Failed to hydrate weapon ${ws.weaponId}`, err);
              return of(null);
            })
          ));
        }
      });
    });

    if (hydrationTasks.length === 0) {
      return of(character);
    }

    return forkJoin(hydrationTasks).pipe(
      map(() => character)
    );
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
