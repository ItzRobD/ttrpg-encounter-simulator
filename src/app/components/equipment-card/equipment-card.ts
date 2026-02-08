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
  public readonly item = input.required<Weapon | Armor | any>();

  protected readonly innerItem = computed(() => {
    const item = this.item();
    if (!item) return null;
    return item.weapon || item.armor || item;
  });

  protected readonly isWeapon = computed(() => {
    const inner = this.innerItem();
    return !!inner && ('damageBlocks' in inner);
  });

  protected readonly isArmor = computed(() => {
    const inner = this.innerItem();
    return !!inner && ('ac' in inner) && !this.isWeapon();
  });

  asWeapon(item: any): Weapon {
    return this.innerItem() as Weapon;
  }

  asArmor(item: any): Armor {
    return this.innerItem() as Armor;
  }

  protected readonly weaponProperties = computed(() => {
    if (!this.isWeapon()) return [];
    const w = this.asWeapon(null);
    const props: string[] = [];
    if (!w?.properties) return props;

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
    const w = this.asWeapon(null);
    const mods: string[] = [];
    if (!w?.modifiers) return mods;

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
