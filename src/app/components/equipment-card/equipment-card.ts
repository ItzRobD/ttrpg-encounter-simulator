import { Component, computed, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { CardModule } from 'primeng/card';
import { Weapon, Armor, DiceType } from '../../models';
import { formatDice, formatModifier, getEquipmentDetail } from '../../shared/utils/dnd-utils';
import { TagModule } from 'primeng/tag';

@Component({
  selector: 'app-equipment-card',
  standalone: true,
  imports: [CardModule, CommonModule, TagModule],
  templateUrl: './equipment-card.html',
  styleUrl: './equipment-card.css',
})
export class EquipmentCard {
  public readonly item = input.required<Weapon | Armor>();

  protected readonly isWeapon = computed(() => 'damageBlocks' in this.item() || 'die' in this.item());
  protected readonly isArmor = computed(() => 'ac' in this.item() && !('die' in this.item()) && !('damageBlocks' in this.item()));

  asWeapon(item: Weapon | Armor): Weapon {
    return item as Weapon;
  }

  asArmor(item: Weapon | Armor): Armor {
    return item as Armor;
  }

  protected readonly weaponProperties = computed(() => {
    if (!this.isWeapon()) return [];
    const w = this.asWeapon(this.item());
    const props: string[] = [];
    if (w.properties.isVersatile) props.push('Versatile');
    if (w.properties.isFinesse) props.push('Finesse');
    if (w.properties.isRanged) props.push('Ranged');
    if (w.properties.isHeavy) props.push('Heavy');
    if (w.properties.isTwoHanded) props.push('Two-Handed');
    if (w.properties.isLight) props.push('Light');
    if (w.properties.isThrown) props.push('Thrown');
    return props;
  });

  protected readonly weaponModifiers = computed(() => {
    if (!this.isWeapon()) return [];
    const w = this.asWeapon(this.item());
    const mods: string[] = [];
    if (w.modifiers.isMagic) mods.push('Magic');
    if (w.modifiers.isSilvered) mods.push('Silvered');
    if (w.modifiers.isAdamantine) mods.push('Adamantine');
    if (w.modifiers.isColdForgedIron) mods.push('Cold-Forged Iron');
    return mods;
  });

  protected readonly formatDice = formatDice;
  protected readonly formatModifier = formatModifier;
  protected readonly getEquipmentDetail = getEquipmentDetail;
}
