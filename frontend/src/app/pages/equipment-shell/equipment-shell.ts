import { Component, inject, OnInit, signal, ChangeDetectionStrategy } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { TabsModule } from 'primeng/tabs';
import { LibraryPage } from '../../components/library-page/library-page';
import { SharedTable } from '../../components/shared-table/shared-table.component';
import { EquipmentCard } from '../../components/equipment-card/equipment-card.component';
import { EquipmentService } from '../../services/equipment.service';
import { EquipmentEditorComponent } from '../../components/editors/equipment-editor/equipment-editor';
import { EquipmentItem } from '../../models';

@Component({
  selector: 'app-equipment-shell',
  imports: [
    ButtonModule,
    TooltipModule,
    TabsModule,
    LibraryPage,
    SharedTable,
    EquipmentCard,
    EquipmentEditorComponent,
  ],
  templateUrl: './equipment-shell.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class EquipmentShell implements OnInit {
  public readonly equipmentService = inject(EquipmentService);
  public readonly searchTerm = signal('');
  public readonly isEditorVisible = signal(false);
  public readonly itemToEdit = signal<EquipmentItem | null>(null);
  public readonly activeTab = signal('all');

  onTabChange(event: string | number | undefined): void {
    if (typeof event === 'string') {
      this.activeTab.set(event);
    }
  }

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
}
