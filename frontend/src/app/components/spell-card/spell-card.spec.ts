import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SpellCard } from './spell-card';

describe('SpellCard', () => {
  let component: SpellCard;
  let fixture: ComponentFixture<SpellCard>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SpellCard]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SpellCard);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
