import {Component, computed, inject} from '@angular/core';
import {TimelineService} from '../../services/timeline.service';
import {SimulationService} from '../../services/simulation.service';
import {Tab, TabList, Tabs} from 'primeng/tabs';
import {MapperService} from '../../services/mapper.service';
import {TreeTableModule} from 'primeng/treetable';
import {SliderModule} from 'primeng/slider';
import {FormsModule} from '@angular/forms';
import {CommonModule, TitleCasePipe} from '@angular/common';
import {ButtonModule} from 'primeng/button';
import {TooltipModule} from 'primeng/tooltip';
import {DiceType, EventType, SimulationEvent} from '../../models';
import {formatDice} from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-simulation-results',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    Tab,
    Tabs,
    TabList,
    TreeTableModule,
    SliderModule,
    ButtonModule,
    TooltipModule,
  ],
  providers: [TitleCasePipe],
  templateUrl: './simulation-results.html',
  styleUrl: './simulation-results.css',
})
export class SimulationResults {
  public readonly simulationService = inject(SimulationService);
  public readonly timelineService = inject(TimelineService);
  private readonly mapperService = inject(MapperService);
  private readonly titleCasePipe = inject(TitleCasePipe);

  protected readonly EventType = EventType;

  protected readonly treeNodes = computed(() => {
    const log = this.timelineService.selectedSimulationLog();
    if (!log) return [];

    // log.events is already mapped to camelCase by SimulationService
    return this.mapperService.mapSimulationLog(log.events);
  });

  protected readonly logIndicies = computed(() => {
    const result = this.simulationService.simulationResult();
    const count = result?.count ?? 0;
    return Array.from({ length: count }, (_, i) => i);
  });

  onTabChange(value: string | number | undefined): void {
    if (value === undefined) return;
    const index = typeof value === 'string' ? parseInt(value, 10) : value;
    const result = this.simulationService.simulationResult();
    if (result && result.logs[index]) {
      this.timelineService.setSelectedSimulationLog(result.logs[index]);
    }
  }

  isEventActive(id: string): boolean {
    return this.timelineService.activeEvent()?.id === id;
  }

  getEventLabel(event: SimulationEvent): string {
    switch (event.type) {
      case EventType.Round:
        return `Round ${event.round}`;
      case EventType.Turn:
        return 'Turn';
      case EventType.Choice:
        return `${this.titleCasePipe.transform(event.data.choiceType)} Choice`;
      case EventType.SavingThrow:
        return `Saving Throw`;
      case EventType.DamageRoll:
        return `Damage Roll`;
      case EventType.DamageModified:
        return `Damage Modified`;
      case EventType.HPModified:
        return `HP Modified`;
      case EventType.Death:
        return `Death`;
      case EventType.Unconscious:
        return `Unconscious`;
      case EventType.Victory:
        return `Victory`;
      case EventType.Equipment:
        return `Weapon`;
      default:
        return this.titleCasePipe.transform(event.type);
    }
  }

  getEventDetails(event: SimulationEvent): string {
    let details = '';

    if (event.type === EventType.Choice) {
      if (event.data.choiceType === 'target') {
        details = event.data.target?.name || event.data.choice || event.data.note || 'None';
      } else if (event.data.choiceType === 'action') {
        details = event.data.choice?.includes('damage') ? 'Damage' : 'Healing';
      } else {
        details = event.data.choice || event.data.note || '';
      }
    } else {
      details = event.data.note || event.data.choice || '';
    }

    if (event.type === EventType.Initiative) {
      const modifier = event.data.roll?.modifier
      const value = event.data.roll?.total
      if (modifier !== undefined && value !== undefined && modifier !== 0) {
        details = `Initiative Roll: ${value} + ${modifier}`;
      } else if (value !== undefined) {
        details = `Initiative Roll: ${value}`;
      } else {
        details = 'Initiative Roll';
      }

      return details;
    }

    if (event.type === EventType.Attack) {
      const isHit = event.data.diceRoll?.success;
      const defType = event.data.attackType === 'melee' ? 'AC' : 'DC';
      const hitStatus = isHit ? `Hit vs ${event.data.diceRoll?.targetValue} ${defType}` : 'Missed';
      details = `${details} ${hitStatus}`.trim();
    }

    if (event.type === EventType.Death || event.type === EventType.Unconscious) {
      if (event.data.target) {
        details = `${event.data.target.name} ${event.type === EventType.Death ? 'died' : 'fell unconscious'}`;
      }
    }

    if (event.type === EventType.SavingThrow) {
      const isSuccess = event.data.diceRoll?.success;
      const targetDC = event.data.diceRoll?.targetValue;
      const status = isSuccess ? `Success vs DC ${targetDC}` : `Failed vs DC ${targetDC}`;
      details = `${details} ${status}`.trim();
    }

    if (event.type === EventType.Victory) {
      const winners = this.titleCasePipe.transform(event.data.winner);
      details = `${winners} won in ${event.data.rounds} rounds`
    }

    if (event.type === EventType.HPModified) {
      const value = event.data.value;
      const targetName = event.data.target?.name || 'HP';
      if (value !== undefined) {
        const verb = value > 0 ? 'increased' : 'decreased';
        details = `${targetName} ${verb} by ${Math.abs(value)}`;
      } else {
        details = `${targetName} modified`;
      }
    }

    if (event.type === EventType.Equipment) {
      if (event.data.die && event.data.name && event.data.numberOfDice) {
        const weaponName = event.data.name;
        const dieValue = parseInt(event.data.die, 10);
        const diceType = dieValue as DiceType;
        const properties = event.data.properties || [];
        const modifiers = event.data.modifiers || [];
        const propString = properties.length > 0 ? ` (${properties.map(p => this.titleCasePipe.transform(p)).join(', ')})` : '';
        const modString = modifiers.length > 0 ? ` (${modifiers.map(m => this.titleCasePipe.transform(m)).join(', ')})` : '';

        details = `${weaponName} - ${formatDice(
          event.data.numberOfDice,
          diceType,
          event.data.damageBonus,
        )} ${event.data.damageType}${propString}${modString}`;``;
      }
    }

    if (event.data.roll) {
      const dieValue = parseInt(event.data.roll.die, 10);
      const diceType = dieValue as DiceType;
      const damageType = this.titleCasePipe.transform(event.data.damageType) || '';
      details += ` Total: ${event.data.roll.total}, Dice: ${formatDice(
        event.data.roll.numberOfDice,
        diceType,
        event.data.roll.modifier
      )} ${damageType}`;
    }

    return details;
  }
}
