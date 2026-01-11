import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';
import { TimelineService } from './timeline.service';

describe('TimelineService', () => {
  let service: TimelineService;

  beforeEach(() => {
    const localStorageSpy = {
      getItem: vi.fn().mockReturnValue(null),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    };
    (window as any).localStorage = localStorageSpy;

    TestBed.configureTestingModule({});
    service = TestBed.inject(TimelineService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
