import {Action} from '../core';

export interface Lair {
  name: string;
  actions: Action[];
  availability: { [actionId: number]: boolean };
}
