import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'snakeToSpace',
  standalone: true
})
export class SnakeToSpacePipe implements PipeTransform {
  transform(value: string | undefined): string {
    if (!value) return '';
    return value.replace(/_/g, ' ');
  }
}
