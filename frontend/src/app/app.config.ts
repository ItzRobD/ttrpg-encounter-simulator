import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { providePrimeNG } from 'primeng/config';
import { DndPreset } from './theme/dnd-preset';
import { routes } from './app.routes';
import { mappingInterceptor } from './interceptors/mapping.interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideHttpClient(withInterceptors([mappingInterceptor])),
    provideAnimationsAsync(),
    providePrimeNG({
      theme: {
        preset: DndPreset,
        options: {
          darkModeSelector: '.app-dark',
          cssLayer: { name: 'primeng', order: 'reset, primeng' },
        },
      },
    }),
  ],
};
