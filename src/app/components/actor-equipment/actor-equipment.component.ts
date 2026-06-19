import { Component, inject, input, ChangeDetectionStrategy } from '@angular/core';
import { Actor, WeaponSlot, WeaponModifiers } from '../../models';
import { CardModule } from 'primeng/card';
import { EquipmentService } from '../../services/equipment.service';

@Component({
  selector: 'app-actor-equipment',
  standalone: true,
  imports: [CardModule],
  templateUrl: './actor-equipment.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './actor-equipment.component.css',
})
export class ActorEquipment {
  private readonly equipmentService = inject(EquipmentService);
  public readonly actor = input.required<Actor>();

  protected readonly weaponSlots: { key: WeaponSlot; label: string }[] = [
    { key: WeaponSlot.Primary, label: 'Primary' },
    { key: WeaponSlot.Secondary, label: 'Secondary' },
    { key: WeaponSlot.Ranged, label: 'Ranged' },
  ];

  getWeaponName(weaponId: number | string): string {
    const summary = this.equipmentService.summaries().find(s =>
      s.id.toString() === weaponId.toString() && s.type === 'Weapon'
    );
    return summary ? summary.name : `Weapon #${weaponId}`;
  }

  getArmorName(armorId: number | string): string {
    const summary = this.equipmentService.summaries().find(s =>
      s.id.toString() === armorId.toString() && (s.type === 'Armor' || s.type === 'Shield')
    );
    return summary ? summary.name : `Armor/Shield #${armorId}`;
  }

  getShieldName(shieldId: number | string): string {
    const summary = this.equipmentService.summaries().find(s =>
      s.id.toString() === shieldId.toString() && s.type === 'Shield'
    );
    return summary ? summary.name : `Shield #${shieldId}`;
  }

  formatModifiers(modifiers: WeaponModifiers): string {
    const parts = [];
    if (modifiers.attackBonus) parts.push(`+${modifiers.attackBonus} to hit`);
    if (modifiers.damageBonus) parts.push(`+${modifiers.damageBonus} damage`);
    if (modifiers.isMagic) parts.push('Magic');
    return parts.length > 0 ? ` (${parts.join(', ')})` : '';
  }
}
