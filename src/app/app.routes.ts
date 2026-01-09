import { Routes } from '@angular/router';
import { AppLayout } from './layout/app-layout/app-layout';
import { SimulatorShell } from './pages/simulator-shell/simulator-shell';
import { HistoryShell } from './pages/history-shell/history-shell';
import { BestiaryShell } from './pages/bestiary-shell/bestiary-shell';
export const routes: Routes = [
  {
    path: '',
    component: AppLayout,
    children: [
      { path: '', component: SimulatorShell },
      { path: 'history', component: HistoryShell },
      { path: 'bestiary', component: BestiaryShell },
    ],
  },
  { path: '**', redirectTo: '' },
];
