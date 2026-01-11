import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SimulationResults } from './simulation-results';

describe('SimulationResults', () => {
  let component: SimulationResults;
  let fixture: ComponentFixture<SimulationResults>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SimulationResults]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SimulationResults);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
