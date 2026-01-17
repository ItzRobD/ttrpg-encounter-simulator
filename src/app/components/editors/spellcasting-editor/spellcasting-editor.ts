import { Component, input, OnInit, inject, OnChanges, SimpleChanges, signal, ChangeDetectionStrategy, computed, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormGroup, ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { SelectModule } from 'primeng/select';
import { InputNumberModule } from 'primeng/inputnumber';
import { CheckboxModule } from 'primeng/checkbox';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { FormsModule } from '@angular/forms';
import { CasterType, Ability, Spell, SpellSummary } from '../../../models';
import { SpellsService } from '../../../services/spells.service';

@Component({
  selector: 'app-spellcasting-editor',
  imports: [CommonModule, ReactiveFormsModule, SelectModule, InputNumberModule, CheckboxModule, ButtonModule, FormsModule, TooltipModule],
  templateUrl: './spellcasting-editor.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class SpellcastingEditorComponent implements OnInit, OnChanges {
  public readonly group = input.required<FormGroup>(); // Expects Spellcasting group structure
  public readonly classId = input<string | number>();

  private readonly spellsService = inject(SpellsService);

  protected readonly abilities = Object.values(Ability).map(a => ({
    label: a.charAt(0).toUpperCase() + a.slice(1),
    value: a
  }));

  protected availableSummaries = signal<SpellSummary[]>([]);
  protected selectedSummary: SpellSummary | null = null;
  protected loadingSpells = signal(false);

  private readonly casterTypeSignal = signal<CasterType>(CasterType.None);

  public readonly slotLevels = computed(() => {
    const casterType = this.casterTypeSignal();
    if (casterType === CasterType.None) return [];
    return [1, 2, 3, 4, 5, 6, 7, 8, 9];
  });

  constructor() {
    effect(() => {
      const group = this.group();
      if (group) {
        const ctrl = group.get('casterType');
        if (ctrl) {
          this.casterTypeSignal.set(ctrl.value);
          const sub = ctrl.valueChanges.subscribe(v => this.casterTypeSignal.set(v));
          return () => sub.unsubscribe();
        }
      }
      return;
    });
  }

  ngOnInit(): void {
    this.loadSummaries();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['classId'] && !changes['classId'].firstChange) {
      this.loadSummaries();
    }
  }

  private loadSummaries(): void {
    this.loadingSpells.set(true);
    const cId = this.classId();
    if (cId) {
      this.spellsService.getSummariesByClass(cId).subscribe({
        next: (summaries) => {
          this.availableSummaries.set(summaries);
          this.loadingSpells.set(false);
        },
        error: () => this.loadingSpells.set(false)
      });
    } else {
      this.spellsService.getSummaries().subscribe({
        next: (summaries) => {
          this.availableSummaries.set(summaries);
          this.loadingSpells.set(false);
        },
        error: () => this.loadingSpells.set(false)
      });
    }
  }

  get spells(): Spell[] {
    return this.group().get('spells')?.value || [];
  }

  onAddSpell(): void {
    if (!this.selectedSummary) return;

    const summary = this.selectedSummary;
    const currentSpells = this.spells;

    if (currentSpells.some(s => s.id === summary.id)) {
      this.selectedSummary = null;
      return;
    }

    this.loadingSpells.set(true);
    this.spellsService.selectSpellByID(summary.id.toString()).subscribe({
      next: (fullSpell) => {
        const updatedSpells = [...currentSpells, fullSpell];
        this.group().get('spells')?.setValue(updatedSpells);
        this.selectedSummary = null;
        this.loadingSpells.set(false);
      },
      error: () => {
        this.loadingSpells.set(false);
      }
    });
  }

  removeSpell(index: number): void {
    const currentSpells = [...this.spells];
    currentSpells.splice(index, 1);
    this.group().get('spells')?.setValue(currentSpells);
  }
}
