import {Action} from './actor.model';

export interface Lair {
  name: string;
  actions: Action[];
  availability: { [actionId: number]: boolean };
}
