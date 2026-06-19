import { Routes } from '@angular/router';
import { AppLayout } from './layout/app-layout/app-layout';
import { SimulatorShell } from './pages/simulator-shell/simulator-shell';
import { HistoryShell } from './pages/history-shell/history-shell';
import { BestiaryShell } from './pages/bestiary-shell/bestiary-shell';
import {CharactersShell} from './pages/characters-shell/characters-shell';
import {EquipmentShell} from './pages/equipment-shell/equipment-shell';
import {SpellsShell} from './pages/spells-shell/spells-shell';
export const routes: Routes = [
  {
    path: '',
    component: AppLayout,
    children: [
      { path: '', component: SimulatorShell },
      { path: 'history', component: HistoryShell },
      { path: 'bestiary', component: BestiaryShell },
      { path: 'characters', component: CharactersShell },
      { path: 'equipment', component: EquipmentShell },
      { path: 'spells', component: SpellsShell },
    ],
  },
  { path: '**', redirectTo: '' },
];
