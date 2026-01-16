import { inject, Injectable, signal, computed, effect} from '@angular/core';
import { LocalStorageService } from './local-storage.service';
import { SubscriptionService } from './subscription.service';
import { HttpClient } from '@angular/common/http';
import {
  Monster, MonsterSummary,
  Character, CharacterSummary,
  Spell, SpellSummary,
  EquipmentItem, EquipmentSummary,
  UserTier
} from '../models';
import { Observable, of, tap } from 'rxjs';
import { environment } from '../../environments/environment';

export type CustomEntityType = 'monsters' | 'characters' | 'spells' | 'equipment';

@Injectable({
  providedIn: 'root'
})
export class CustomContentService {
  private readonly localStorage = inject(LocalStorageService);
  private readonly subscriptionService = inject(SubscriptionService);
  private readonly http = inject(HttpClient);

  // Local signals for each type
  private readonly _customMonsters = signal<Monster[]>([]);
  private readonly _customCharacters = signal<Character[]>([]);
  private readonly _customSpells = signal<Spell[]>([]);
  private readonly _customEquipment = signal<EquipmentItem[]>([]);

  // Public readonly signals
  public readonly customMonsters = this._customMonsters.asReadonly();
  public readonly customCharacters = this._customCharacters.asReadonly();
  public readonly customSpells = this._customSpells.asReadonly();
  public readonly customEquipment = this._customEquipment.asReadonly();

  constructor() {
    this.loadInitialData();

    // Automatically persist to localStorage whenever signals change
    effect(() => {
      this.localStorage.setItem('custom_monsters', this._customMonsters());
    });
    effect(() => {
      this.localStorage.setItem('custom_characters', this._customCharacters());
    });
    effect(() => {
      this.localStorage.setItem('custom_spells', this._customSpells());
    });
    effect(() => {
      this.localStorage.setItem('custom_equipment', this._customEquipment());
    });
  }

  private loadInitialData() {
    // For now, always load from local storage as we might have guest data
    // TODO: Later, if premium, we might want to fetch from API and merge or replace
    const monsters = this.localStorage.getItem<Monster[]>('custom_monsters') || [];
    const characters = this.localStorage.getItem<Character[]>('custom_characters') || [];
    const spells = this.localStorage.getItem<Spell[]>('custom_spells') || [];
    const equipment = this.localStorage.getItem<EquipmentItem[]>('custom_equipment') || [];

    this._customMonsters.set(monsters);
    this._customCharacters.set(characters);
    this._customSpells.set(spells);
    this._customEquipment.set(equipment);
  }

  /**
   * Saves a custom entity.
   * Handles LocalStorage for free users and API for premium users.
   */
  saveEntity<T extends { id?: string | number, isCustom?: boolean }>(type: CustomEntityType, entity: T): Observable<T> {
    entity.isCustom = true;

    if (this.subscriptionService.isPremium()) {
      // If it has a local ID, remove it before sending to API so server can generate a fresh one
      if (entity.id && typeof entity.id === 'string' && entity.id.startsWith('local-')) {
        delete entity.id;
      }
      return this.saveToApi(type, entity);
    } else {
      // For local storage, if no ID, generate one (since there is no backend)
      if (!entity.id) {
        entity.id = `local-${crypto.randomUUID()}`;
      }
      this.saveToLocal(type, entity as T & { id: string | number });
      return of(entity as T & { id: string | number });
    }
  }

  /**
   * Deletes a custom entity.
   */
  deleteEntity(type: CustomEntityType, id: string | number): Observable<void> {
    if (this.subscriptionService.isPremium()) {
      return this.deleteFromApi(type, id);
    } else {
      this.deleteFromLocal(type, id);
      return of(undefined);
    }
  }

  private saveToLocal<T extends { id: string | number }>(type: CustomEntityType, entity: T) {
    const signalMap = {
      monsters: this._customMonsters,
      characters: this._customCharacters,
      spells: this._customSpells,
      equipment: this._customEquipment
    };

    const targetSignal = signalMap[type];
    targetSignal.update(items => {
      const index = items.findIndex(i => i.id === entity.id);
      if (index > -1) {
        const newItems = [...items];
        newItems[index] = entity as any;
        return newItems;
      }
      return [...items, entity as any];
    });

    // Refresh limits if needed (though backend is source of truth,
    // we might want to trigger a fetch or just wait for next poll)
    this.subscriptionService.fetchLimits();
  }

  private deleteFromLocal(type: CustomEntityType, id: string | number) {
     const signalMap = {
      monsters: this._customMonsters,
      characters: this._customCharacters,
      spells: this._customSpells,
      equipment: this._customEquipment
    };

    const targetSignal = signalMap[type];
    (targetSignal as any).update((items: any[]) => items.filter(i => i.id !== id));
    this.subscriptionService.fetchLimits();
  }

  private saveToApi<T extends { id?: string | number }>(type: CustomEntityType, entity: T): Observable<T> {
    const url = `${environment.apiUrl}/custom/${type}`;
    return this.http.post<T>(url, entity).pipe(
      tap(saved => {
        // Also update local signals so UI is snappy
        this.saveToLocal(type, saved as T & { id: string | number });
      })
    );
  }

  private deleteFromApi(type: CustomEntityType, id: string | number): Observable<void> {
    const url = `${environment.apiUrl}/custom/${type}/${id}`;
    return this.http.delete<void>(url).pipe(
      tap(() => {
        this.deleteFromLocal(type, id);
      })
    );
  }
}
