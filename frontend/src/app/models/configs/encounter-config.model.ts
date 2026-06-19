import { Actor } from '../combatants';
import { SimulationOptions } from '../simoptions.model';

export interface EncounterConfig {
  combatants: Actor[];
  options: SimulationOptions;
  maxRounds: number;
  iterations: number;
}
