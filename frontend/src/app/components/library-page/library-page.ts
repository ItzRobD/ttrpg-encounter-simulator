import { Component, ChangeDetectionStrategy, input, model } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { IconFieldModule } from 'primeng/iconfield';
import { InputIconModule } from 'primeng/inputicon';
import { InputTextModule } from 'primeng/inputtext';

@Component({
  selector: 'app-library-page',
  imports: [FormsModule, IconFieldModule, InputIconModule, InputTextModule],
  templateUrl: './library-page.html',
  styleUrl: './library-page.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LibraryPage {
  readonly title = input.required<string>();
  readonly description = input('');
  readonly listTitle = input.required<string>();
  readonly detailTitle = input.required<string>();
  readonly emptyMessage = input.required<string>();
  readonly hasSelection = input(false);
  readonly searchPlaceholder = input('Search...');
  readonly searchLabel = input('');
  readonly search = model('');
}
