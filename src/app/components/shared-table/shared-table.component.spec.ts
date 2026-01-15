import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SharedTable } from './shared-table.component';

describe('SharedTable', () => {
  let component: SharedTable;
  let fixture: ComponentFixture<SharedTable>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SharedTable]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SharedTable);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
