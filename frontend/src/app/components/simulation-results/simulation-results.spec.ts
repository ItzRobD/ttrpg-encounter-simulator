import { ComponentFixture, TestBed } from '@angular/core/testing';
import { vi } from 'vitest';
import { SimulationResults } from './simulation-results';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('SimulationResults', () => {
  let component: SimulationResults;
  let fixture: ComponentFixture<SimulationResults>;

  beforeEach(async () => {
    const localStorageSpy = {
      getItem: vi.fn().mockReturnValue(null),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    };
    (window as any).localStorage = localStorageSpy;

    await TestBed.configureTestingModule({
      imports: [SimulationResults, HttpClientTestingModule]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SimulationResults);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
