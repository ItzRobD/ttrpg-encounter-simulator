import { Component, inject, OnInit, signal } from '@angular/core';
import { SharedTable } from '../../components/shared-table/shared-table.component';
import { EquipmentCard } from '../../components/equipment-card/equipment-card';
import { EquipmentService } from '../../services/equipment.service';
import { ButtonModule } from 'primeng/button';
import { IconFieldModule } from 'primeng/iconfield';
import { InputIconModule } from 'primeng/inputicon';
import { InputTextModule } from 'primeng/inputtext';
import { TooltipModule } from 'primeng/tooltip';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-equipment-shell',
  standalone: true,
  imports: [
    SharedTable,
    EquipmentCard,
    ButtonModule,
    IconFieldModule,
    InputIconModule,
    InputTextModule,
    TooltipModule,
    CommonModule
  ],
  templateUrl: './equipment-shell.html',
  styleUrl: './equipment-shell.css',
  styles: [`
    :host {
      display: block;
    }
  `]
})
export class EquipmentShell implements OnInit {
  public readonly equipmentService = inject(EquipmentService);
  public readonly searchTerm = signal('');

  ngOnInit(): void {
    this.equipmentService.getSummaries().subscribe();
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchTerm.set(target.value);
  }

  onClearSearch(): void {
    this.searchTerm.set('');
  }
}
