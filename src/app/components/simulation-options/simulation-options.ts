import { Component, inject, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DrawerModule } from 'primeng/drawer';
import { ButtonModule } from 'primeng/button';
import { ToggleSwitchModule } from 'primeng/toggleswitch';
import { InputNumberModule } from 'primeng/inputnumber';
import { SelectButtonModule } from 'primeng/selectbutton';
import { TooltipModule } from 'primeng/tooltip';
import { AccordionModule, AccordionPanel, AccordionHeader, AccordionContent } from 'primeng/accordion';
import { SimulationOptions, HPVisibilityMode, IntermissionConfig } from '../../models';
import { SimulationService } from '../../services/simulation.service';
import {environment} from '../../../environments/environment';

@Component({
  selector: 'app-simulation-options',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    DrawerModule,
    ButtonModule,
    ToggleSwitchModule,
    InputNumberModule,
    SelectButtonModule,
    TooltipModule,
    AccordionModule,
    AccordionPanel,
    AccordionHeader,
    AccordionContent
  ],
  templateUrl: './simulation-options.html',
  styleUrl: './simulation-options.css'
})
export class SimulationOptionsComponent {
  private readonly simulationService = inject(SimulationService);

  public readonly visible = input.required<boolean>();
  public readonly visibleChange = output<boolean>();

  protected readonly options = this.simulationService.options;
  protected readonly config = this.simulationService.config;
  protected readonly intermissionConfig = this.simulationService.intermissionConfig;

  protected readonly hpVisibilityOptions = [
    { label: 'Visible', value: 'visible' },
    { label: 'Hidden', value: 'hidden' },
    { label: 'Percentage', value: 'percentage' }
  ];

  protected readonly multiattackPolicyOptions = [
    { label: 'Aggressive', value: 'aggressive' },
    { label: 'Random', value: 'random' },
    { label: 'Smart', value: 'smart' }
  ];

  protected readonly actionSelectionPolicyOptions = [
    { label: 'Weighted', value: 'weighted' },
    { label: 'Random', value: 'random' },
    { label: 'Highest Damage', value: 'highest_damage' },
    { label: 'Utility', value: 'utility' }
  ];

  onHide() {
    this.visibleChange.emit(false);
  }

  updateOption(key: keyof SimulationOptions, value: any) {
    this.simulationService.updateOptions({ [key]: value });
  }

  updateSeed(part: 'seed1' | 'seed2', value: number | null) {
    const currentSeed = this.options().seed;
    this.simulationService.updateOptions({
      seed: { ...currentSeed, [part]: value || 0 }
    });
  }

  updateConfig(part: 'numberOfRuns' | 'maxRounds' | 'includeLogs', value: number | boolean | null) {
    const currentConfig = this.config();
    this.simulationService.updateConfig({
      ...currentConfig,
      [part]: value || false
    });
  }

  updateIntermission(key: keyof IntermissionConfig, value: number | null) {
    this.simulationService.updateIntermissionConfig({ [key]: value || 0 });
  }

  protected readonly environment = environment;
}
