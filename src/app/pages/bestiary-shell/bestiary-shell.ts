import { Component, ChangeDetectionStrategy, inject, OnInit, signal } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { IconFieldModule } from 'primeng/iconfield';
import { InputIconModule } from 'primeng/inputicon';
import { InputTextModule } from 'primeng/inputtext';
import { EntityTable } from '../../components/entity-table/entity-table.component';
import { MonsterService } from '../../services/monster.service';
import {EntityCard} from '../../components/entity-card/entity-card';

@Component({
  selector: 'app-bestiary-shell',
  standalone: true,
  imports: [
    ButtonModule,
    TooltipModule,
    IconFieldModule,
    InputIconModule,
    InputTextModule,
    EntityTable,
    EntityCard
  ],
  templateUrl: './bestiary-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class BestiaryShell implements OnInit {
  public readonly monsterService = inject(MonsterService);
  public readonly searchTerm = signal('');

  ngOnInit(): void {
    this.monsterService.getSummaries().subscribe();
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchTerm.set(target.value);
  }

  onClearSearch(): void {
    this.searchTerm.set('');
  }
}
