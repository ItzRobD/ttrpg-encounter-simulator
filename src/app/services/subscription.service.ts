import { inject, Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { UserLimits, UserTier } from '../models/subscription.model';
import { catchError, tap } from 'rxjs/operators';
import { of } from 'rxjs';
import {environment} from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class SubscriptionService {
  private readonly http = inject(HttpClient);
  // Using the provided endpoint
  private readonly apiUrl = `${environment.apiUrl}/users/limits`;

  private readonly _limits = signal<UserLimits | null>({
    tier: UserTier.Premium,
    monsters: { current: 0, max: 10 },
    characters: { current: 0, max: 10 },
    spells: { current: 0, max: 10 },
    equipment: { current: 0, max: 10 }
  });
  public readonly limits = this._limits.asReadonly();

  private readonly _loading = signal(false);
  public readonly loading = this._loading.asReadonly();

  private readonly _error = signal<string | null>(null);
  public readonly error = this._error.asReadonly();

  /**
   * Fetches user limits from the backend.
   */
  fetchLimits() {
    this._loading.set(true);
    this._error.set(null);

    return this.http.get<UserLimits>(this.apiUrl).pipe(
      tap((limits) => {
        this._limits.set(limits);
        this._loading.set(false);
      }),
      catchError((err) => {
        console.error('Failed to fetch user limits', err);
        // Do not set error signal here to avoid UI noise for non-logged in users
        // this._error.set('Failed to load subscription limits.');
        this._loading.set(false);
        return of(null);
      })
    );
  }

  /**
   * Manually update the current usage for a specific type.
   * Useful for local storage mode to keep UI in sync.
   */
  updateLocalUsage(type: keyof Omit<UserLimits, 'tier' | 'userId'>, count: number): void {
    const currentLimits = this._limits();
    if (currentLimits && currentLimits.tier === UserTier.Free) {
      this._limits.set({
        ...currentLimits,
        [type]: {
          ...currentLimits[type],
          current: count
        }
      });
    }
  }

  /**
   * Helper to check if a user can create another item of a specific type.
   */
  canCreate(type: keyof Omit<UserLimits, 'tier' | 'userId'>): boolean {
    const l = this._limits();
    if (!l) return true; // Fallback to true if somehow null, allowing save attempt

    const usage = l[type];
    return usage.current < usage.max;
  }

  /**
   * Returns the current tier of the user.
   */
  getTier(): UserTier {
    const tier = this._limits()?.tier;
    if (!tier) return UserTier.Free;
    if (typeof tier === 'string') {
      return tier.toLowerCase() as UserTier;
    }
    return tier;
  }

  /**
   * Returns true if the user is a premium or pro member.
   */
  isPremium(): boolean {
    const tier = this.getTier();
    return tier === UserTier.Premium || tier === UserTier.Pro;
  }
}
