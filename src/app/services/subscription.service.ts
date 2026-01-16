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

  private readonly _limits = signal<UserLimits | null>(null);
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
        this._error.set('Failed to load subscription limits.');
        this._loading.set(false);
        return of(null);
      })
    );
  }

  /**
   * Helper to check if a user can create another item of a specific type.
   */
  canCreate(type: keyof UserLimits['usage']): boolean {
    const l = this._limits();
    if (!l) return false;

    const usage = l.usage[type];
    return usage.current < usage.max;
  }

  /**
   * Returns the current tier of the user.
   */
  getTier(): UserTier {
    return this._limits()?.tier || UserTier.Free;
  }

  /**
   * Returns true if the user is a premium or pro member.
   */
  isPremium(): boolean {
    const tier = this.getTier();
    return tier === UserTier.Premium || tier === UserTier.Pro;
  }
}
