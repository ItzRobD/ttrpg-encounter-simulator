import { Directive, inject, input, output, signal } from '@angular/core';
import { CustomContentService, CustomActorType } from '../../services/custom-content.service';
import { SubscriptionService } from '../../services/subscription.service';
import { Observable, finalize } from 'rxjs';

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
   * Abstract method for getting the actor type (monsters, spells, etc.)
   */
  protected abstract getActorType(): CustomActorType;

  /**
   * Closes the dialog
   */
  close(): void {
    this.visibleChange.emit(false);
    this.error.set(null);
  }

  /**
   * Generic save method
   */
  protected saveActor(actor: T): void {
    const type = this.getActorType();

    // Check limits for new items only
    if (!this.itemToEdit() && !this.subscriptionService.canCreate(type)) {
      this.error.set(`You have reached the limit for custom ${type} in your current tier.`);
      return;
    }

    this.loading.set(true);
    this.error.set(null);

    this.customContentService.saveActor(type, actor)
      .pipe(finalize(() => this.loading.set(false)))
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
