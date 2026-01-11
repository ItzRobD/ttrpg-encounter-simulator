import { Component, input } from '@angular/core';
import { Character, WeaponSlot } from '../../models';
import { CardModule } from 'primeng/card';
import { TitleCasePipe } from '@angular/common';
import { formatWeaponData } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-entity-equipment',
  standalone: true,
  imports: [CardModule, TitleCasePipe],
  templateUrl: './entity-equipment.html',
  styleUrl: './entity-equipment.css',
})
export class EntityEquipment {
  public readonly character = input.required<Character>();

  protected readonly weaponSlots: WeaponSlot[] = [
    WeaponSlot.Primary,
    WeaponSlot.Secondary,
    WeaponSlot.Ranged,
  ];

  protected readonly formatWeaponData = formatWeaponData;
}
