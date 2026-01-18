import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { ButtonModule } from 'primeng/button';
import { DrawerModule } from 'primeng/drawer';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-app-layout',
  standalone: true,
  imports: [CommonModule, RouterOutlet, RouterLink, RouterLinkActive, ButtonModule, DrawerModule],
  templateUrl: './app-layout.html',
  styles: [
    `
      :host {
        display: block;
        height: 100vh;
      }

      .env-banner {
        height: 1px;
        background-color: #6a1b9a;
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        z-index: 2000;
      }

      .env-tab {
        position: fixed;
        top: 0;
        left: 50%;
        transform: translateX(-50%);
        background-color: #6a1b9a;
        color: white;
        padding: 0.25rem 2rem;
        font-size: 1rem;
        font-weight: bold;
        white-space: nowrap;
        box-shadow: 0 2px 4px rgba(0,0,0,0.2);
        z-index: 2001;
        font-family: monospace;
        clip-path: polygon(0 0, 100% 0, 85% 100%, 15% 100%);
      }

      @media screen and (min-width: 768px) {
        .layout-wrapper {
          padding-top: 1.5rem;
        }
      }
    `,
  ],
})
export class AppLayout {
  visible = false;
  protected readonly environmentName = environment.name;
}
