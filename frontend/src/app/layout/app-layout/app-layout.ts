import { Component, inject, signal, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { ButtonModule } from 'primeng/button';
import { DrawerModule } from 'primeng/drawer';
import { TooltipModule } from 'primeng/tooltip';
import { PopoverModule } from 'primeng/popover';
import { ThemeSwitcher } from '../../theme/theme-switcher/theme-switcher';
import { ThemePreviewDialog } from '../../theme/theme-preview-dialog/theme-preview-dialog';
import { ThemeService } from '../../services/theme.service';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-app-layout',
  imports: [
    CommonModule,
    RouterOutlet,
    RouterLink,
    RouterLinkActive,
    ButtonModule,
    DrawerModule,
    TooltipModule,
    PopoverModule,
    ThemeSwitcher,
    ThemePreviewDialog
  ],
  templateUrl: './app-layout.html',
  styleUrl: './app-layout.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AppLayout {
  visible = false;
  protected readonly theme = inject(ThemeService);
  protected readonly environmentName = environment.name;

  protected readonly navItems = [
    { label: 'Simulator', icon: 'pi pi-bolt', route: '/', exact: true },
    { label: 'Characters', icon: 'pi pi-users', route: '/characters' },
    { label: 'Bestiary', icon: 'pi pi-book', route: '/bestiary' },
    { label: 'Equipment', icon: 'pi pi-shield', route: '/equipment' },
    { label: 'Spells', icon: 'pi pi-sparkles', route: '/spells' },
    { label: 'History', icon: 'pi pi-history', route: '/history' },
  ];

  public readonly isSidebarCollapsed = signal(true);
  public readonly autoSlideSidebar = signal(true);
  public readonly themePreviewVisible = signal(false);

  toggleSidebar() {
    this.isSidebarCollapsed.update(v => !v);
  }

  toggleAutoSlide() {
    this.autoSlideSidebar.update(v => !v);
  }
}
