import { computed, inject, Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { MapperService } from './mapper.service';
import { Armor, EquipmentItem, EquipmentSummary, Weapon } from '../models';
import { catchError, forkJoin, Observable, of, retry, tap, throwError } from 'rxjs';
import { map } from 'rxjs/operators';
import { EntityService } from './entity.service.interface';
import { getEquipmentDetail } from '../shared/utils/dnd-utils';
import { CustomContentService } from './custom-content.service';

import { ApiResponse, DataEnvelope } from '../models';

@Injectable({
  providedIn: 'root',
})
export class EquipmentService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/equipment`;
  private readonly mapperService = inject(MapperService);
  private readonly customContentService = inject(CustomContentService);

  private readonly _summaries = signal<EquipmentSummary[]>([]);
  public readonly summaries = computed(() => {
    const customSummaries: EquipmentSummary[] = this.customContentService.customEquipment().map(i => {
      const isWeapon = 'die' in i;
      let type: 'Weapon' | 'Armor' | 'Shield' = 'Armor';
      if (isWeapon) {
        type = 'Weapon';
      } else if (i.name.toLowerCase().includes('shield')) {
        type = 'Shield';
      }

      return {
        id: i.id || 0,
        name: i.name,
        isCustom: true,
        type: type,
        detail: getEquipmentDetail(i),
        properties: isWeapon ? {
          isVersatile: (i as Weapon).properties.isVersatile,
          isFinesse: (i as Weapon).properties.isFinesse,
          isHeavy: (i as Weapon).properties.isHeavy,
          isLight: (i as Weapon).properties.isLight,
          isTwoHanded: (i as Weapon).properties.isTwoHanded,
          isThrown: (i as Weapon).properties.isThrown,
          isRanged: (i as Weapon).properties.isRanged || (i as Weapon).properties.isOnlyRanged
        } : undefined
      };
    });
    return [...customSummaries, ...this._summaries()];
  });

  private readonly _armorList = signal<Armor[]>([]);
  public readonly armorList = computed(() => {
    const customArmor = this.customContentService.customEquipment().filter(i => !('die' in i)) as Armor[];
    return [...customArmor, ...this._armorList()];
  });

  private readonly _weaponList = signal<Weapon[]>([]);
  public readonly weaponList = computed(() => {
    const customWeapons = this.customContentService.customEquipment().filter(i => 'die' in i) as Weapon[];
    return [...customWeapons, ...this._weaponList()];
  });

  private readonly _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  private readonly _selectedItem = signal<EquipmentItem | null>(null);
  public readonly selectedItem = this._selectedItem.asReadonly();

  // Selected Equipment alias
  public readonly selectedEquipment = this.selectedItem;

  getSummaries(forceRefresh = false): Observable<EquipmentSummary[]> {
    if (!forceRefresh && this._summaries().length > 0) {
      return of(this._summaries());
    }

    this._loading.set(true);
    this._error.set(null);

    const weapons$ = this.http.get<ApiResponse<unknown>>(`${this.apiUrl}/weapons`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => {
        const mapped = this.mapperService.mapKeys(response);
        const dataArray = Array.isArray(mapped)
          ? mapped
          : (mapped && typeof mapped === 'object'
            ? Object.values(mapped as Record<string, unknown>)
            : (mapped ? [mapped] : []));

        const weapons = dataArray as Weapon[];
        this._weaponList.set(weapons);
        return weapons;
      }),
      catchError(() => {
        this._weaponList.set([]);
        return of([] as Weapon[]);
      })
    );

    const armor$ = this.http.get<ApiResponse<unknown>>(`${this.apiUrl}/armor`).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => {
        const mapped = this.mapperService.mapKeys(response);
        const dataArray = Array.isArray(mapped)
          ? mapped
          : (mapped && typeof mapped === 'object'
            ? Object.values(mapped as Record<string, unknown>)
            : (mapped ? [mapped] : []));

        const armor = dataArray as Armor[];
        this._armorList.set(armor);
        return armor;
      }),
      catchError(() => {
        this._armorList.set([]);
        return of([] as Armor[]);
      })
    );

    return forkJoin({ weapons: weapons$, armor: armor$ }).pipe(
      map(({ weapons, armor }) => {
        const weaponSummaries = weapons.map((w: Weapon) => {
          return {
            id: w.id || 0,
            name: w.name,
            type: 'Weapon' as const,
            detail: getEquipmentDetail(w),
            properties: {
              isVersatile: w.properties.isVersatile,
              isFinesse: w.properties.isFinesse,
              isHeavy: w.properties.isHeavy,
              isLight: w.properties.isLight,
              isTwoHanded: w.properties.isTwoHanded,
              isThrown: w.properties.isThrown,
              isRanged: w.properties.isRanged || w.properties.isOnlyRanged
            }
          };
        });

        const armorSummaries = armor.map((a: Armor) => {
          const type: 'Shield' | 'Armor' = a.name.toLowerCase().includes('shield') ? 'Shield' : 'Armor';
          return {
            id: a.id || 0,
            name: a.name,
            type: type,
            detail: getEquipmentDetail(a)
          };
        });

        return [...weaponSummaries, ...armorSummaries];
      }),
      tap(summaries => {
        this._summaries.set(summaries);
        this._loading.set(false);
      }),
      catchError(err => {
        this._loading.set(false);
        this._error.set('Failed to load equipment summaries.');
        return of([]);
      })
    );
  }

  selectItemByID(id: string, type?: 'Weapon' | 'Armor' | 'Shield'): Observable<EquipmentItem> {
    // Try finding in custom equipment first
    const customItem = this.customContentService.customEquipment().find(i => i.id?.toString() === id);
    if (customItem) {
      this._selectedItem.set(customItem);
      return of(customItem);
    }

    this._loading.set(true);
    this._error.set(null);

    // Try finding it in the already loaded lists first
    if (!type || type === 'Weapon') {
      const weapon = this._weaponList().find(w => w.id?.toString() === id);
      if (weapon) {
        this._selectedItem.set(weapon);
        this._loading.set(false);
        return of(weapon);
      }
    }

    if (!type || type === 'Armor' || type === 'Shield') {
      const armor = this._armorList().find(a => a.id?.toString() === id);
      if (armor) {
        this._selectedItem.set(armor);
        this._loading.set(false);
        return of(armor);
      }
    }

    // If not found in lists, try generic endpoint (though we might need to know if it's weapon or armor)
    // For now, let's keep the existing logic as fallback
    let url = `${this.apiUrl}/${id}`;
    if (type) {
      const path = (type === 'Armor' || type === 'Shield') ? 'armor' : 'weapons';
      url = `${this.apiUrl}/${path}/${id}`;
    }

    return this.http.get<ApiResponse<unknown>>(url).pipe(
      retry({
        count: environment.httpRetryCount,
        delay: environment.httpRetryDelay
      }),
      map(response => {
        return this.mapperService.mapKeys(response) as EquipmentItem;
      }),
      tap(item => {
        this._selectedItem.set(item);
        this._loading.set(false);
      }),
      catchError(err => {
        this._loading.set(false);
        this._error.set(`Failed to load equipment item with ID ${id}.`);
        return throwError(() => err);
      })
    );
  }

  selectItem(item: EquipmentItem | null): void {
    this._selectedItem.set(item);
  }

  deleteItem(id: string | number): Observable<void> {
    return this.customContentService.deleteEntity('equipment', id).pipe(
      tap(() => {
        if (this._selectedItem()?.id === id) {
          this._selectedItem.set(null);
        }
      })
    );
  }

  clearError(): void {
    this._error.set(null);
  }
}
