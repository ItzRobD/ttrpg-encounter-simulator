import {Component, computed, inject} from '@angular/core';
import {TimelineService} from '../../services/timeline.service';
import {SimulationService} from '../../services/simulation.service';
import {Tab, TabList, Tabs} from 'primeng/tabs';
import {MapperService} from '../../services/mapper.service';
import {SimulationStateService} from '../../services/simulation-state.service';
import {TreeTableModule} from 'primeng/treetable';
import {SliderModule} from 'primeng/slider';
import {FormsModule} from '@angular/forms';
import {CommonModule, TitleCasePipe} from '@angular/common';
import {ButtonModule} from 'primeng/button';
import {TooltipModule} from 'primeng/tooltip';
import {DiceType, EventType, SimulationEvent} from '../../models';
import {formatDice} from '../../shared/utils/dnd-utils';
import {SnakeToSpacePipe} from '../../pipes/snake-to-space.pipe';

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
    SnakeToSpacePipe,
  ],
  providers: [TitleCasePipe],
  templateUrl: './simulation-results.html',
  styleUrl: './simulation-results.css',
})
export class SimulationResults {
  public readonly simulationService = inject(SimulationService);
  public readonly timelineService = inject(TimelineService);
  public readonly stateService = inject(SimulationStateService);
  private readonly mapperService = inject(MapperService);
  private readonly titleCasePipe = inject(TitleCasePipe);

  protected readonly EventType = EventType;

  protected readonly treeNodes = computed(() => {
    const log = this.stateService.selectedSimulationLog();
    if (!log) return [];

    // log.events is already mapped to camelCase by SimulationService
    return this.mapperService.mapSimulationLog(log.events);
  });

  protected readonly logIndicies = computed(() => {
    const result = this.simulationService.simulationResult();
    if (!result) return [];

    return result.individualResults.map((_, i) => i);
  });

  protected readonly performance = computed(() => {
    return this.simulationService.simulationResult()?.performance;
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

  logEvent(event: SimulationEvent): void {
    console.log('Simulation Event:', event);
  }

  getEventLabel(event: SimulationEvent): string {
    const data = event.data || {};
    switch (event.type) {
      case EventType.Round:
        return `Round ${event.round}`;
      case EventType.Turn:
        return 'Turn';
      case EventType.Choice:
        return `${this.titleCasePipe.transform(data.choiceType)} Choice`;
      case EventType.SavingThrow:
        return `Saving Throw`;
      case EventType.DamageRoll:
        return `Damage Roll`;
      case EventType.DamageModified:
        return `Damage Modified`;
      case EventType.HPModified:
        return `HP Modified`;
      case EventType.DecisionStart:
        return `Decision`;
      case EventType.ActionStart:
        return `Action`;
      case EventType.AttackRoll:
        return `Attack Roll`;
      case EventType.Resolution:
        return `Resolution`;
      case EventType.Death:
        return `Death`;
      case EventType.Unconscious:
        return `Unconscious`;
      case EventType.Victory:
        return `Victory`;
      case EventType.Equipment:
        return `Weapon`;
      default:
        return this.titleCasePipe.transform(event.type.replace(/_/g, ' '));
    }
  }

  getEventDetails(event: SimulationEvent): string {
    let details: string;
    const data = event.data || {};

    if (event.type === EventType.Choice) {
      if (data.choiceType === 'target') {
        details = data.target?.name || data.choice || data.note || 'None';
      } else if (data.choiceType === 'action') {
        details = data.choice?.includes('damage') ? 'Damage' : 'Healing';
      } else {
        details = data.choice || data.note || '';
      }
    } else {
      details = data.note || data.choice || '';
    }

    if (event.type === EventType.Initiative) {
      const modifier = data.roll?.modifier
      const value = data.roll?.total
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
      const isHit = data.diceRoll?.success;
      const defType = data.attackType === 'melee' ? 'AC' : 'DC';
      const hitStatus = isHit ? `Hit vs ${data.diceRoll?.targetValue} ${defType}` : 'Missed';
      details = `${details} ${hitStatus}`.trim();
    }

    if (event.type === EventType.Death || event.type === EventType.Unconscious) {
      if (data.target) {
        details = `${data.target.name} ${event.type === EventType.Death ? 'died' : 'fell unconscious'}`;
      }
    }

    if (event.type === EventType.HPModified) {
      const modification = data.result?.modificationValue;
      const targetName = data.target?.name || 'HP';
      if (modification !== undefined) {
        const verb = data.result?.didHealHp ? 'increased' : 'decreased';
        details = `${targetName} ${verb} by ${Math.abs(modification)}`;
      } else {
        details = `${targetName} modified`;
      }
    }

    if (event.type === EventType.DecisionStart) {
      details = data.decision || '';
    }

    if (event.type === EventType.ActionStart) {
      details = data.actionName || '';
    }

    if (event.type === EventType.Resolution) {
      details = '';
    }

    if (event.type === EventType.AttackRoll || event.type === EventType.SavingThrow || event.type === EventType.DamageRoll) {
      const roll = data.roll;
      if (roll) {
        const dieValue = parseInt(roll.dice, 10);
        const diceType = isNaN(dieValue) ? DiceType.D20 : dieValue as DiceType;
        const damageType = this.titleCasePipe.transform(data.damageType || roll.rollType) || '';

        details = `Total: ${roll.total}, Dice: ${formatDice(
          roll.numberOfDice,
          diceType,
          roll.modifier
        )} ${damageType}`.trim();

        if (event.type === EventType.SavingThrow || event.type === EventType.AttackRoll) {
          const isSuccess = roll.isSuccess;
          const targetDC = data.diceRoll?.targetValue || roll.total; // Fallback if diceRoll missing
          const label = event.type === EventType.SavingThrow ? 'vs DC' : 'vs AC';
          const status = isSuccess ? `Success ${label} ${targetDC}` : `Failed ${label} ${targetDC}`;
          details = `${details} (${status})`;
        }
      }
    }

    if (event.type === EventType.Victory) {
      if (data.winner) {
        details = `${data.winner} won the combat`;
      }
    }

    if (data.roll) {
      const dieValue = parseInt(data.roll.dice, 10);
      const diceType = dieValue as DiceType;
      const damageType = this.titleCasePipe.transform(data.damageType) || '';
      details += ` Total: ${data.roll.total}, Dice: ${formatDice(
        data.roll.numberOfDice,
        diceType,
        data.roll.modifier
      )} ${damageType}`;
    }

    return details;
  }
}
