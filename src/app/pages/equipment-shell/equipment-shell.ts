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
import { EquipmentEditorComponent } from '../../components/editors/equipment-editor/equipment-editor';
import { EquipmentItem } from '../../models';

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
    CommonModule,
    EquipmentEditorComponent
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
  public readonly isEditorVisible = signal(false);
  public readonly itemToEdit = signal<EquipmentItem | null>(null);

  ngOnInit(): void {
    this.equipmentService.getSummaries().subscribe();
  }

  onCreateItem(): void {
    this.itemToEdit.set(null);
    this.isEditorVisible.set(true);
  }

  onEditItem(item: EquipmentItem): void {
    this.itemToEdit.set(item);
    this.isEditorVisible.set(true);
  }

  onDeleteItem(item: EquipmentItem): void {
    if (confirm(`Are you sure you want to delete the custom item "${item.name}"?`)) {
      this.equipmentService.deleteItem(item.id!).subscribe();
    }
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchTerm.set(target.value);
  }

  onClearSearch(): void {
    this.searchTerm.set('');
  }
}
