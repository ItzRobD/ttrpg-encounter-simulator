import { inject, Injectable, signal, computed, effect} from '@angular/core';
import { LocalStorageService } from './local-storage.service';
import { SubscriptionService } from './subscription.service';
import { HttpClient } from '@angular/common/http';
import {
  Spell, SpellSummary,
  EquipmentItem, EquipmentSummary,
  UserTier,
  Actor,
  ActorSummary
} from '../models';
import { Observable, of, tap, Subject } from 'rxjs';
import { environment } from '../../environments/environment';

export type CustomActorType = 'monsters' | 'characters' | 'spells' | 'equipment';

@Injectable({
  providedIn: 'root'
})
export class CustomContentService {
  private readonly localStorage = inject(LocalStorageService);
  private readonly subscriptionService = inject(SubscriptionService);
  private readonly http = inject(HttpClient);

  // Local signals for each type
  private readonly _customMonsters = signal<Actor[]>([]);
  private readonly _customCharacters = signal<Actor[]>([]);
  private readonly _customSpells = signal<Spell[]>([]);
  private readonly _customEquipment = signal<EquipmentItem[]>([]);

  // Public readonly signals
  public readonly customMonsters = this._customMonsters.asReadonly();
  public readonly customCharacters = this._customCharacters.asReadonly();
  public readonly customSpells = this._customSpells.asReadonly();
  public readonly customEquipment = this._customEquipment.asReadonly();

  // Notification for API changes
  private readonly _apiContentChange = new Subject<{type: CustomActorType, action: 'save' | 'delete'}>();
  public readonly apiContentChange$ = this._apiContentChange.asObservable();

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
    const monsters = this.localStorage.getItem<Actor[]>('custom_monsters') || [];
    const characters = this.localStorage.getItem<Actor[]>('custom_characters') || [];
    const spells = this.localStorage.getItem<Spell[]>('custom_spells') || [];
    const equipment = this.localStorage.getItem<EquipmentItem[]>('custom_equipment') || [];

    this._customMonsters.set(monsters);
    this._customCharacters.set(characters);
    this._customSpells.set(spells);
    this._customEquipment.set(equipment);

    // Sync initial limits for free tier
    this.subscriptionService.updateLocalUsage('monsters', monsters.length);
    this.subscriptionService.updateLocalUsage('characters', characters.length);
    this.subscriptionService.updateLocalUsage('spells', spells.length);
    this.subscriptionService.updateLocalUsage('equipment', equipment.length);
  }

  /**
   * Saves a custom actor.
   * Handles LocalStorage for free users and API for premium users.
   */
  saveActor<T extends { id?: string | number, isCustom?: boolean }>(
    type: CustomActorType,
    actor: T,
    forceTarget?: 'local' | 'api'
  ): Observable<T> {
    actor.isCustom = true;

    const useApi = forceTarget ? forceTarget === 'api' : this.subscriptionService.isPremium();

    if (useApi) {
      // If it has a local ID, remove it before sending to API so server can generate a fresh one
      if (actor.id && typeof actor.id === 'string' && actor.id.startsWith('local-')) {
        delete actor.id;
      }
      return this.saveToApi(type, actor);
    } else {
      // For local storage, if no ID, generate one (since there is no backend)
      if (!actor.id) {
        actor.id = `local-${crypto.randomUUID()}`;
      }
      this.saveToLocal(type, actor as T & { id: string | number });
      return of(actor as T & { id: string | number });
    }
  }

  /**
   * Deletes a custom actor.
   */
  deleteActor(type: CustomActorType, id: string | number): Observable<void> {
    if (this.subscriptionService.isPremium()) {
      return this.deleteFromApi(type, id);
    } else {
      this.deleteFromLocal(type, id);
      return of(undefined);
    }
  }

  private saveToLocal<T extends { id: string | number }>(type: CustomActorType, actor: T) {
    const signalMap = {
      monsters: this._customMonsters,
      characters: this._customCharacters,
      spells: this._customSpells,
      equipment: this._customEquipment
    };

    const targetSignal = signalMap[type];
    (targetSignal as any).update((items: any[]) => {
      const index = items.findIndex(i => i.id === actor.id);
      if (index > -1) {
        const newItems = [...items];
        newItems[index] = actor;
        return newItems;
      }
      const newItems = [...items, actor];
      // Update usage limit if it's a new item
      this.subscriptionService.updateLocalUsage(type, newItems.length);
      return newItems;
    });

    // Refresh limits if needed (though backend is source of truth,
    // we might want to trigger a fetch or just wait for next poll)
    this.subscriptionService.fetchLimits();
  }

  private deleteFromLocal(type: CustomActorType, id: string | number) {
     const signalMap = {
      monsters: this._customMonsters,
      characters: this._customCharacters,
      spells: this._customSpells,
      equipment: this._customEquipment
    };

    const targetSignal = signalMap[type] as any;
    targetSignal.update((items: any[]) => {
      const newItems = items.filter((i: any) => i.id !== id);
      this.subscriptionService.updateLocalUsage(type, newItems.length);
      return newItems;
    });
    this.subscriptionService.fetchLimits();
  }

  private saveToApi<T extends { id?: string | number }>(type: CustomActorType, actor: T): Observable<T> {
    let url = `${environment.apiUrl}/custom/${type}`;

    // Specific endpoints for types if they don't follow the generic pattern
    if (type === 'characters') {
      url = `${environment.apiUrl}/characters/save`;
    } else if (type === 'monsters') {
      url = `${environment.apiUrl}/monsters`;
    } else if (type === 'spells') {
      url = `${environment.apiUrl}/spells`;
    } else if (type === 'equipment') {
      url = `${environment.apiUrl}/equipment`;
    }

    return this.http.post<T>(url, actor).pipe(
      tap(() => {
        // Notify that a cloud item has changed so services can refresh
        this._apiContentChange.next({ type, action: 'save' });
      })
    );
  }

  private deleteFromApi(type: CustomActorType, id: string | number): Observable<void> {
    let url = `${environment.apiUrl}/custom/${type}/${id}`;

    if (type === 'characters') {
      url = `${environment.apiUrl}/characters/delete/${id}`;
    } else if (type === 'monsters') {
      url = `${environment.apiUrl}/monsters/${id}`;
    } else if (type === 'spells') {
      url = `${environment.apiUrl}/spells/${id}`;
    } else if (type === 'equipment') {
      url = `${environment.apiUrl}/equipment/${id}`;
    }

    return this.http.delete<void>(url).pipe(
      tap(() => {
        this._apiContentChange.next({ type, action: 'delete' });
      })
    );
  }
}
