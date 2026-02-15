import {Component, computed, inject, signal} from '@angular/core';
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

  protected readonly activeTabIndex = signal(0);

  protected readonly logIndices = computed(() => {
    const result = this.simulationService.simulationResult();
    if (!result) return [];

    return result.individualResults.map((_, i) => i);
  });

  protected readonly activeEncounterIndex = signal(0);

  protected readonly encounterIndices = computed(() => {
    const result = this.simulationService.simulationResult();
    const runIndex = this.activeTabIndex();
    if (!result || !result.individualResults[runIndex]) return [];

    return result.individualResults[runIndex].encounterResults.map((_, i) => i);
  });

  protected readonly treeNodes = computed(() => {
    const result = this.simulationService.simulationResult();
    const runIndex = this.activeTabIndex();
    const encounterIndex = this.activeEncounterIndex();

    if (!result || !result.individualResults[runIndex]) return [];

    const run = result.individualResults[runIndex];
    const encounter = run.encounterResults[encounterIndex];

    if (!encounter) return [];

    // log.events is already mapped to camelCase by SimulationService
    return this.mapperService.mapSimulationLog(encounter.logs);
  });

  protected readonly performance = computed(() => {
    return this.simulationService.simulationResult()?.performance;
  });

  onTabChange(value: string | number | undefined): void {
    if (value === undefined) return;
    const index = typeof value === 'string' ? parseInt(value, 10) : value;
    this.activeTabIndex.set(index);
    this.activeEncounterIndex.set(0);
    this.updateSelectedLog();
  }

  onEncounterChange(value: string | number | undefined): void {
    if (value === undefined) return;
    const index = typeof value === 'string' ? parseInt(value, 10) : value;
    this.activeEncounterIndex.set(index);
    this.updateSelectedLog();
  }

  private updateSelectedLog(): void {
    const result = this.simulationService.simulationResult();
    const runIndex = this.activeTabIndex();
    const encounterIndex = this.activeEncounterIndex();

    if (result && result.individualResults[runIndex]) {
        const run = result.individualResults[runIndex];
        const encounter = run.encounterResults[encounterIndex];
        if (encounter) {
            this.timelineService.setSelectedSimulationLog({
                actors: [],
                events: encounter.logs,
                initialState: run.initialState,
                actorInitialStates: run.actorInitialStates,
                actorConfigs: result.actorConfigs
            });
        }
    }
  }

  isEventActive(id: string): boolean {
    const active = this.timelineService.activeEvent();
    if (active?.id === id) return true;

    // Special case for rounds and turns which might not have parentIds in the same way
    if (id.startsWith('round-')) {
      const roundNum = parseInt(id.replace('round-', ''), 10);
      return active?.round === roundNum;
    }

    // Check if the event is an ancestor of the active event
    const log = this.stateService.selectedSimulationLog();
    if (!log) return false;

    let current: SimulationEvent | undefined | null = active;
    while (current?.parentId) {
      if (current.parentId === id) return true;
      current = log.events.find(e => e.id === current?.parentId);
    }

    return false;
  }

  logEvent(event: SimulationEvent): void {
    console.log('Simulation Event:', event);
  }

  getEventLabel(event: SimulationEvent): string {
    if (event.type === EventType.CombatStart) {
      return 'Combat Started';
    }

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

    if (event.type === EventType.CombatStart) {
      details = 'Initial health and conditions set.';
    }

    if (event.type === EventType.Resolution) {
      details = '';
    }

    if (event.type === EventType.AttackRoll || event.type === EventType.SavingThrow || event.type === EventType.DamageRoll) {
      const roll = data.roll;
      if (roll) {
        const diceType = (roll.dice as unknown as DiceType) || DiceType.D20;
        const damageType = this.titleCasePipe.transform(roll.rollType) || '';

        details = `Total: ${roll.total}, Dice: ${formatDice(
          roll.numberOfDice,
          diceType,
          roll.modifier
        )} ${damageType}`.trim();

        if (event.type === EventType.SavingThrow || event.type === EventType.AttackRoll || roll.rollType === 'attack' || roll.rollType === 'saving throw') {
          const isAttack = event.type === EventType.AttackRoll || roll.rollType === 'attack';
          const isSuccess = isAttack ? data.isHit : data.saveSuccess;
          const targetValue = isAttack
            ? (data.targetAc)
            : (data.dc);

          const label = isAttack ? 'vs AC' : 'vs DC';
          const status = isSuccess ? `Success ${label} ${targetValue}` : `Failed ${label} ${targetValue}`;
          details = `${details} (${status})`;
        }
      }
    }

    if (event.type === EventType.Victory) {
      if (data.winner) {
        details = `${data.winner} won the combat`;
      }
    }

    if (data.roll && (event.type as any) !== EventType.AttackRoll && (event.type as any) !== EventType.SavingThrow && (event.type as any) !== EventType.DamageRoll && (event.type as any) !== EventType.Initiative) {
      const diceType = (data.roll.dice as unknown as DiceType);
      details += ` Total: ${data.roll.total}, Dice: ${formatDice(
        data.roll.numberOfDice,
        diceType,
        data.roll.modifier
      )}`;
    }

    return details;
  }
}
