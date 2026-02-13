import { Directive, inject, input, output, signal, computed } from '@angular/core';
import { CustomContentService, CustomActorType } from '../../services/custom-content.service';
import { SubscriptionService } from '../../services/subscription.service';
import { Observable, finalize } from 'rxjs';
import { MenuItem } from 'primeng/api';

@Directive()
export abstract class BaseEditorDirective<T extends { id?: string | number; name: string; isCustom?: boolean }> {
  protected readonly customContentService = inject(CustomContentService);
  protected readonly subscriptionService = inject(SubscriptionService);

  // Visibility of the dialog
  public readonly visible = input.required<boolean>();
  public readonly visibleChange = output<boolean>();

  // Item being edited (null if creating new)
  public readonly itemToEdit = input<T | null>(null);

  // State management
  public readonly loading = signal(false);
  public readonly error = signal<string | null>(null);

  /**
   * Computed property to determine the primary save label and whether to show split options.
   */
  public readonly isPremium = computed(() => this.subscriptionService.isPremium());

  public readonly saveOptions = computed<MenuItem[]>(() => {
    const isPremium = this.isPremium();
    const type = this.getActorType();

    const options: MenuItem[] = [
      {
        label: 'Save to Cloud',
        icon: 'pi pi-cloud-upload',
        disabled: !isPremium,
        tooltipOptions: !isPremium ? {
          tooltipLabel: 'Premium feature: Store your creations in the cloud to access them from anywhere!',
          tooltipPosition: 'top'
        } : undefined,
        command: () => {
          if (isPremium) {
            this.onSaveAttempt('api');
          }
        }
      },
      {
        label: 'Save Locally',
        icon: 'pi pi-desktop',
        command: () => this.onSaveAttempt('local')
      }
    ];

    return options;
  });

  public readonly primarySaveLabel = computed(() => {
    return this.isPremium() ? 'Save to Cloud' : 'Save Locally';
  });

  /**
   * Abstract method for getting the actor type (monsters, spells, etc.)
   */
  protected abstract getActorType(): CustomActorType;

  /**
   * This should be implemented by the component to trigger the validation and call saveActor
   */
  public abstract onSave(): void;

  /**
   * Helper to handle the save attempt from the UI
   */
  protected onSaveAttempt(target?: 'local' | 'api'): void {
    // This will be called by the component's onSave or by SplitButton items
    // If target is provided, we pass it down
    this._saveTarget = target;
    this.onSave();
  }

  private _saveTarget?: 'local' | 'api';

  /**
   * Closes the dialog
   */
  close(): void {
    this.visibleChange.emit(false);
    this.error.set(null);
    this._saveTarget = undefined;
  }

  /**
   * Generic save method
   */
  protected saveActor(actor: T): void {
    const type = this.getActorType();
    const target = this._saveTarget;

    // Check limits for new items only
    // Note: If saving locally, we still check limits (for free tier)
    // If saving to API, the server handles limits, but we can do a client side check too if we have the info
    if (!this.itemToEdit() && !this.subscriptionService.canCreate(type)) {
      this.error.set(`You have reached the limit for custom ${type} in your current tier.`);
      return;
    }

    this.loading.set(true);
    this.error.set(null);

    this.customContentService.saveActor(type, actor, target)
      .pipe(finalize(() => {
        this.loading.set(false);
        this._saveTarget = undefined;
      }))
      .subscribe({
        next: () => {
          this.close();
        },
        error: (err: any) => {
          console.error(`Failed to save ${type}`, err);
          this.error.set(`Failed to save ${type}. Please try again.`);
        }
      });
  }
}
