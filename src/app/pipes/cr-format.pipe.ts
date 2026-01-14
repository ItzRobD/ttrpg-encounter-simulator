import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'crFormat',
  standalone: true
})
export class CrFormatPipe implements PipeTransform {
  transform(value: number | string | null | undefined): string {
    if (value === null || value === undefined) {
      return '';
    }

    const num = typeof value === 'string' ? parseFloat(value) : value;

    if (isNaN(num)) {
      return value.toString();
    }

    if (num === 0.125) {
      return '1/8';
    } else if (num === 0.25) {
      return '1/4';
    } else if (num === 0.5) {
      return '1/2';
    }

    return num.toString();
  }
}
