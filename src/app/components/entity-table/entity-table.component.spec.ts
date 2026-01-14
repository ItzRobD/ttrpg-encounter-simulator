import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EntityTable } from './entity-table.component';

describe('EntityTable', () => {
  let component: EntityTable;
  let fixture: ComponentFixture<EntityTable>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EntityTable]
    })
    .compileComponents();

    fixture = TestBed.createComponent(EntityTable);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
